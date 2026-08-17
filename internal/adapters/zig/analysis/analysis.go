// Package analysis converts a Zig project into Senbon's normalized graph.
package analysis

import (
	"context"
	_ "embed"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strings"

	"github.com/briheet/senbon/internal/model"
)

//go:embed analyze.zig
var helperSource []byte

// Project holds the analyzed graph and the module map needed to build the target.
type Project struct {
	Graph   *model.StaticGraph
	Entry   string
	Modules map[string]string   // import name -> source file
	Imports map[string][]string // source file -> its local import names
}

// helperGraph mirrors the JSON emitted by analyze.zig.
type helperGraph struct {
	Root  uint64       `json:"root"`
	Entry string       `json:"entry"`
	Files []helperFile `json:"files"`
	Edges []helperEdge `json:"edges"`
}

// helperFile is one analyzed source file.
type helperFile struct {
	Path      string           `json:"path"`
	Imports   []string         `json:"imports"`
	Functions []helperFunction `json:"functions"`
}

// helperFunction is one function declaration.
type helperFunction struct {
	ID        uint64 `json:"id"`
	Name      string `json:"name"`
	StartLine int    `json:"startLine"`
	StartCol  int    `json:"startCol"`
	EndLine   int    `json:"endLine"`
	EndCol    int    `json:"endCol"`
}

// helperEdge is one resolved call edge.
type helperEdge struct {
	From uint64 `json:"from"`
	To   uint64 `json:"to"`
}

// Analyze runs the embedded helper and builds the project model.
func Analyze(ctx context.Context, projectDir string) (*Project, error) {
	if _, err := exec.LookPath("zig"); err != nil {
		return nil, fmt.Errorf("zig is required for Zig analysis: %w", err)
	}
	helperPath, err := writeHelper()
	if err != nil {
		return nil, err
	}
	defer func() { _ = os.Remove(helperPath) }()

	outputPath := filepath.Join(os.TempDir(), fmt.Sprintf("senbon-zig-%d.json", os.Getpid()))
	defer func() { _ = os.Remove(outputPath) }()

	command := exec.CommandContext(ctx, "zig", "run", helperPath, "-lc", "--", projectDir, outputPath)
	if output, err := command.CombinedOutput(); err != nil {
		return nil, fmt.Errorf("analysis helper failed: %w: %s", err, output)
	}
	data, err := os.ReadFile(outputPath)
	if err != nil {
		return nil, err
	}
	return parse(data, projectDir)
}

// parse builds a project from helper JSON output.
func parse(data []byte, projectDir string) (*Project, error) {
	var helper helperGraph
	if err := json.Unmarshal(data, &helper); err != nil {
		return nil, err
	}

	project := &Project{
		Modules: make(map[string]string),
		Imports: make(map[string][]string),
	}
	graph := &model.StaticGraph{
		Root:     model.NodeID(helper.Root + 1),
		Nodes:    make(map[model.NodeID]*model.StaticNode),
		Files:    make(map[model.FileID]*model.StaticFile),
		Packages: make(map[model.PackageID]*model.Package),
	}
	project.Graph = graph
	graph.Packages[1] = &model.Package{Path: projectDir, Name: filepath.Base(projectDir)}

	fileIDs := make(map[string]model.FileID, len(helper.Files))
	for index, file := range helper.Files {
		id := model.FileID(index + 1)
		fileIDs[file.Path] = id
		graph.Files[id] = &model.StaticFile{ID: id, Path: file.Path, Package: 1}
		graph.Program.Files = append(graph.Program.Files, id)
	}
	for _, file := range helper.Files {
		fileID := fileIDs[file.Path]
		for _, fn := range file.Functions {
			node := &model.StaticNode{
				Name: fn.Name,
				ID:   model.NodeID(fn.ID + 1),
				Pkg:  1,
				Syntax: model.Syntax{
					File:  fileID,
					Start: model.Position{Line: fn.StartLine + 1, Column: fn.StartCol + 1},
					End:   model.Position{Line: fn.EndLine + 1, Column: fn.EndCol + 1},
				},
			}
			graph.Nodes[node.ID] = node
			graph.Files[fileID].Functions = append(graph.Files[fileID].Functions, node.ID)
		}
	}
	for _, edge := range helper.Edges {
		from := model.NodeID(edge.From + 1)
		to := model.NodeID(edge.To + 1)
		if _, ok := graph.Nodes[from]; !ok {
			continue
		}
		if _, ok := graph.Nodes[to]; !ok {
			continue
		}
		if slices.Contains(graph.Nodes[from].Out, to) {
			continue
		}
		graph.Nodes[from].Out = append(graph.Nodes[from].Out, to)
		graph.Nodes[to].In = append(graph.Nodes[to].In, from)
	}
	for _, node := range graph.Nodes {
		slices.Sort(node.In)
		slices.Sort(node.Out)
	}

	project.Entry = helper.Entry
	for _, file := range helper.Files {
		for _, name := range file.Imports {
			if isBuiltinImport(name) {
				continue
			}
			if _, ok := project.Modules[name]; !ok {
				if module, ok := resolveModule(helper.Files, name); ok {
					project.Modules[name] = module
				}
			}
			if _, ok := project.Modules[name]; ok {
				project.Imports[file.Path] = append(project.Imports[file.Path], name)
			}
		}
	}
	return project, nil
}

func isBuiltinImport(name string) bool {
	return name == "std" || name == "builtin" || name == "root"
}

// resolveModule finds the project file for an import name.
func resolveModule(files []helperFile, name string) (string, bool) {
	var candidate string
	if strings.HasSuffix(name, ".zig") {
		candidate = filepath.Base(name)
	} else {
		candidate = filepath.Base(name) + ".zig"
	}
	for _, file := range files {
		if filepath.Base(file.Path) == candidate {
			return file.Path, true
		}
	}
	return "", false
}

// writeHelper writes the embedded analyzer to a temporary file.
func writeHelper() (string, error) {
	file, err := os.CreateTemp("", "senbon-analyze-*.zig")
	if err != nil {
		return "", err
	}
	if _, err := file.Write(helperSource); err != nil {
		_ = file.Close()
		_ = os.Remove(file.Name())
		return "", err
	}
	if err := file.Close(); err != nil {
		_ = os.Remove(file.Name())
		return "", err
	}
	return file.Name(), nil
}
