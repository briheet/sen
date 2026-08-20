// Package analysis derives a Rust source graph and address map from DWARF.
package analysis

import (
	"debug/dwarf"
	"debug/elf"
	"debug/macho"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"github.com/briheet/sen/internal/model"
	"golang.org/x/arch/arm64/arm64asm"
	"golang.org/x/arch/x86/x86asm"
)

type executable struct {
	dwarf *dwarf.Data
	text  []byte
	base  uint64
	arm64 bool
	x86   bool
}

type function struct {
	name      string
	file      string
	startLine int
	endLine   int
	low, high uint64
	node      model.NodeID
}

// Frame is a source-symbolized instruction address.
type Frame struct {
	Function string
	File     string
	Line     int64
}

// Symbols maps instruction addresses and names back to source locations.
type Symbols struct {
	functions []function
}

// Symbolize resolves an instruction address, tolerating ASLR by falling back to its function name.
func (s *Symbols) Symbolize(address uint64, name string) Frame {
	for _, function := range s.functions {
		if address >= function.low && address < function.high || name != "" && (name == function.name || strings.HasSuffix(function.name, "::"+name)) {
			return Frame{Function: function.name, File: function.file, Line: int64(function.startLine)}
		}
	}
	return Frame{Function: name}
}

// Analyze reads the executable and its platform debug companion.
func Analyze(binary, workspace, packageName string) (*model.StaticGraph, *Symbols, error) {
	executable, err := openExecutable(binary)
	if err != nil {
		return nil, nil, err
	}
	functions, err := readFunctions(executable.dwarf, workspace)
	if err != nil {
		return nil, nil, err
	}
	if len(functions) == 0 {
		return nil, nil, errors.New("Rust executable contains no workspace DWARF functions")
	}
	graph := buildGraph(functions, workspace, packageName)
	resolveCalls(executable, functions, graph)
	return graph, &Symbols{functions: functions}, nil
}

func openExecutable(path string) (*executable, error) {
	if file, err := macho.Open(path); err == nil {
		data, debugErr := file.DWARF()
		if debugErr != nil || data == nil {
			data, debugErr = openDSYM(path)
		}
		if debugErr != nil {
			return nil, debugErr
		}
		section := file.Section("__text")
		if section == nil {
			return nil, errors.New("Rust Mach-O executable has no __text section")
		}
		text, err := section.Data()
		if err != nil {
			return nil, err
		}
		return &executable{dwarf: data, text: text, base: section.Addr, arm64: file.Cpu == macho.CpuArm64, x86: file.Cpu == macho.CpuAmd64}, nil
	}
	file, err := elf.Open(path)
	if err != nil {
		return nil, fmt.Errorf("open Rust executable: %w", err)
	}
	data, err := file.DWARF()
	if err != nil {
		return nil, err
	}
	section := file.Section(".text")
	if section == nil {
		return nil, errors.New("Rust ELF executable has no .text section")
	}
	text, err := section.Data()
	if err != nil {
		return nil, err
	}
	return &executable{dwarf: data, text: text, base: section.Addr, arm64: file.Machine == elf.EM_AARCH64, x86: file.Machine == elf.EM_X86_64}, nil
}

func openDSYM(binary string) (*dwarf.Data, error) {
	link := binary + ".dSYM"
	entries, err := os.ReadDir(filepath.Join(link, "Contents", "Resources", "DWARF"))
	if err != nil || len(entries) == 0 {
		return nil, errors.New("Rust macOS build did not produce a readable dSYM")
	}
	file, err := macho.Open(filepath.Join(link, "Contents", "Resources", "DWARF", entries[0].Name()))
	if err != nil {
		return nil, err
	}
	return file.DWARF()
}

func readFunctions(data *dwarf.Data, workspace string) ([]function, error) {
	reader := data.Reader()
	var result []function
	for {
		entry, err := reader.Next()
		if err != nil {
			return nil, err
		}
		if entry == nil {
			break
		}
		if entry.Tag != dwarf.TagCompileUnit {
			if entry.Children {
				reader.SkipChildren()
			}
			continue
		}
		compDir, _ := entry.Val(dwarf.AttrCompDir).(string)
		lineReader, _ := data.LineReader(entry)
		files := []*dwarf.LineFile(nil)
		if lineReader != nil {
			files = lineReader.Files()
		}
		first := len(result)
		if err := walk(reader, nil, compDir, files, workspace, &result); err != nil {
			return nil, err
		}
		if lineReader != nil {
			fillFunctionLines(lineReader, compDir, result[first:])
		}
	}
	slices.SortFunc(result, func(left, right function) int {
		if left.low < right.low {
			return -1
		}
		if left.low > right.low {
			return 1
		}
		return strings.Compare(left.name, right.name)
	})
	return result, nil
}

func fillFunctionLines(reader *dwarf.LineReader, compDir string, functions []function) {
	var line dwarf.LineEntry
	for reader.Next(&line) == nil {
		if line.EndSequence || line.File == nil || line.Line <= 0 {
			continue
		}
		path := sourcePath(line.File.Name, compDir)
		for index := range functions {
			function := &functions[index]
			if line.Address >= function.low && line.Address < function.high && path == function.file {
				function.endLine = max(function.endLine, line.Line)
			}
		}
	}
}

func walk(reader *dwarf.Reader, scope []string, compDir string, files []*dwarf.LineFile, workspace string, functions *[]function) error {
	for {
		entry, err := reader.Next()
		if err != nil {
			return err
		}
		if entry == nil {
			return nil
		}
		if entry.Tag == 0 {
			return nil
		}
		name, _ := entry.Val(dwarf.AttrName).(string)
		nextScope := scope
		if name != "" && (entry.Tag == dwarf.TagNamespace || entry.Tag == dwarf.TagStructType || entry.Tag == dwarf.TagClassType) {
			nextScope = append(append([]string(nil), scope...), cleanScope(name))
		}
		if entry.Tag == dwarf.TagSubprogram {
			if function, ok := dwarfFunction(entry, scope, compDir, files, workspace); ok {
				*functions = append(*functions, function)
			}
			if name != "" {
				nextScope = append(append([]string(nil), scope...), cleanScope(name))
			}
		}
		if entry.Children {
			if err := walk(reader, nextScope, compDir, files, workspace, functions); err != nil {
				return err
			}
		}
	}
}

func dwarfFunction(entry *dwarf.Entry, scope []string, compDir string, files []*dwarf.LineFile, workspace string) (function, bool) {
	name, _ := entry.Val(dwarf.AttrName).(string)
	low, ok := entry.Val(dwarf.AttrLowpc).(uint64)
	if !ok || name == "" {
		return function{}, false
	}
	high := highPC(entry.Val(dwarf.AttrHighpc), low)
	if high <= low {
		return function{}, false
	}
	file := dwarfFile(entry.Val(dwarf.AttrDeclFile), files, compDir)
	if file == "" || !within(file, workspace) {
		return function{}, false
	}
	line := integer(entry.Val(dwarf.AttrDeclLine))
	if line <= 0 {
		line = 1
	}
	qualified := append(append([]string(nil), scope...), cleanScope(name))
	return function{name: strings.Join(qualified, "::"), file: file, startLine: line, endLine: line, low: low, high: high}, true
}

func buildGraph(functions []function, workspace, packageName string) *model.StaticGraph {
	graph := &model.StaticGraph{Nodes: make(map[model.NodeID]*model.StaticNode), Files: make(map[model.FileID]*model.StaticFile), Packages: map[model.PackageID]*model.Package{1: {Path: workspace, Name: packageName}}}
	fileIDs := make(map[string]model.FileID)
	for index := range functions {
		entry := &functions[index]
		fileID, ok := fileIDs[entry.file]
		if !ok {
			fileID = model.FileID(len(fileIDs) + 1)
			fileIDs[entry.file] = fileID
			graph.Files[fileID] = &model.StaticFile{ID: fileID, Path: entry.file, Package: 1}
		}
		entry.node = model.NodeID(index + 1)
		node := &model.StaticNode{ID: entry.node, Name: entry.name, Pkg: 1, Syntax: model.Syntax{Kind: model.SyntaxFuncDecl, File: fileID, Start: model.Position{Line: entry.startLine}, End: model.Position{Line: entry.endLine}}}
		graph.Nodes[node.ID] = node
		graph.Files[fileID].Functions = append(graph.Files[fileID].Functions, node.ID)
		if graph.Root == 0 && (entry.name == "main" || strings.HasSuffix(entry.name, "::main")) {
			graph.Root = node.ID
		}
	}
	if graph.Root == 0 {
		graph.Root = 1
	}
	return graph
}

func resolveCalls(executable *executable, functions []function, graph *model.StaticGraph) {
	for _, caller := range functions {
		start, end := int64(caller.low-executable.base), int64(caller.high-executable.base)
		if start < 0 || end <= start || end > int64(len(executable.text)) {
			continue
		}
		code := executable.text[start:end]
		var targets []uint64
		if executable.arm64 {
			targets = armCalls(code, caller.low)
		}
		if executable.x86 {
			targets = x86Calls(code, caller.low)
		}
		for _, address := range targets {
			index := slices.IndexFunc(functions, func(candidate function) bool { return address >= candidate.low && address < candidate.high })
			if index < 0 || functions[index].node == caller.node {
				continue
			}
			callee := functions[index]
			node := graph.Nodes[caller.node]
			if !slices.Contains(node.Out, callee.node) {
				node.Out = append(node.Out, callee.node)
			}
			in := graph.Nodes[callee.node]
			if !slices.Contains(in.In, caller.node) {
				in.In = append(in.In, caller.node)
			}
		}
	}
}

func armCalls(code []byte, base uint64) []uint64 {
	var result []uint64
	for offset := 0; offset+4 <= len(code); offset += 4 {
		instruction, err := arm64asm.Decode(code[offset:])
		if err != nil || instruction.Op != arm64asm.BL {
			continue
		}
		if relative, ok := instruction.Args[0].(arm64asm.PCRel); ok {
			result = append(result, uint64(int64(base)+int64(offset)+int64(relative)))
		}
	}
	return result
}

func x86Calls(code []byte, base uint64) []uint64 {
	var result []uint64
	for offset := 0; offset < len(code); {
		instruction, err := x86asm.Decode(code[offset:], 64)
		if err != nil || instruction.Len == 0 {
			offset++
			continue
		}
		if instruction.Op == x86asm.CALL {
			if relative, ok := instruction.Args[0].(x86asm.Rel); ok {
				result = append(result, uint64(int64(base)+int64(offset+instruction.Len)+int64(relative)))
			}
		}
		offset += instruction.Len
	}
	return result
}

func dwarfFile(value any, files []*dwarf.LineFile, compDir string) string {
	index := integer(value)
	if index < 0 || index >= len(files) || files[index] == nil {
		return ""
	}
	name := files[index].Name
	return sourcePath(name, compDir)
}

func sourcePath(name, compDir string) string {
	if !filepath.IsAbs(name) {
		name = filepath.Join(compDir, name)
	}
	name, _ = filepath.Abs(name)
	return filepath.Clean(name)
}

func highPC(value any, low uint64) uint64 {
	switch value := value.(type) {
	case uint64:
		return value
	case int64:
		return low + uint64(value)
	case int:
		return low + uint64(value)
	default:
		return 0
	}
}

func integer(value any) int {
	switch value := value.(type) {
	case int64:
		return int(value)
	case uint64:
		return int(value)
	case int:
		return value
	default:
		return 0
	}
}

func cleanScope(value string) string {
	value = strings.Trim(strings.TrimSpace(value), "{}")
	value = strings.ReplaceAll(value, "{closure#", "closure#")
	return value
}

func within(path, root string) bool {
	relative, err := filepath.Rel(root, path)
	return err == nil && relative != ".." && !strings.HasPrefix(relative, ".."+string(filepath.Separator))
}
