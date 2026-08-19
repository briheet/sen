package analysis

import (
	"errors"
	"go/ast"
	"go/token"
	"go/types"
	"slices"
	"strconv"
	"strings"

	"github.com/briheet/sen/internal/model"
	"golang.org/x/tools/go/packages"
)

type (
	Graph     = model.StaticGraph
	Node      = model.StaticNode
	File      = model.StaticFile
	PackageID = model.PackageID
	NodeID    = model.NodeID
	FileID    = model.FileID
)

const (
	SyntaxFuncDecl = model.SyntaxFuncDecl
	SyntaxFuncLit  = model.SyntaxFuncLit
)

var (
	ErrNoMainPackage  = errors.New("no main package found")
	ErrNoMainFunction = errors.New("no main function was found")
	ErrInvalidPackage = errors.New("cannot analyze an invalid package")
)

type sourceFunction struct {
	pkg    *packages.Package
	body   *ast.BlockStmt
	object *types.Func
	parent *sourceFunction
	node   *Node
	file   FileID
	pos    token.Pos
	end    token.Pos
	kind   model.SyntaxKind
	name   string
}

// GetGraph builds the source graph for project packages. Runtime traces add
// relationships that static type information cannot resolve.
func GetGraph(pkgs []*packages.Package) (*Graph, error) {
	if len(pkgs) == 0 || pkgs[0] == nil || pkgs[0].Types == nil {
		return nil, ErrInvalidPackage
	}
	if pkgs[0].Name != "main" {
		return nil, ErrNoMainPackage
	}

	builder := graphBuilder{
		graph: &Graph{
			Nodes:    make(map[NodeID]*Node),
			Files:    make(map[FileID]*File),
			Packages: make(map[PackageID]*model.Package),
		},
		packageIDs:  make(map[*types.Package]PackageID),
		objectIDs:   make(map[*types.Func]NodeID),
		literalIDs:  make(map[*ast.FuncLit]NodeID),
		rootPackage: pkgs[0].PkgPath,
	}
	if err := builder.collect(pkgs); err != nil {
		return nil, err
	}
	builder.connect()
	if builder.graph.Root == 0 {
		return nil, ErrNoMainFunction
	}
	return builder.graph, nil
}

type graphBuilder struct {
	graph       *Graph
	functions   []*sourceFunction
	packageIDs  map[*types.Package]PackageID
	objectIDs   map[*types.Func]NodeID
	literalIDs  map[*ast.FuncLit]NodeID
	rootPackage string
	nextPackage PackageID
	nextNode    NodeID
}

func (b *graphBuilder) collect(pkgs []*packages.Package) error {
	b.nextPackage = 1
	b.nextNode = 1
	ordered := append([]*packages.Package(nil), pkgs...)
	slices.SortFunc(ordered, func(left, right *packages.Package) int {
		return strings.Compare(left.PkgPath, right.PkgPath)
	})
	for _, pkg := range ordered {
		if pkg == nil || pkg.Types == nil || pkg.TypesInfo == nil || pkg.Fset == nil {
			return ErrInvalidPackage
		}
		b.packageID(pkg.Types)
	}

	type sourceFile struct {
		pkg  *packages.Package
		path string
		ast  *ast.File
	}
	files := make([]sourceFile, 0)
	for _, pkg := range ordered {
		for index, syntax := range pkg.Syntax {
			if index >= len(pkg.CompiledGoFiles) {
				continue
			}
			files = append(files, sourceFile{pkg: pkg, path: pkg.CompiledGoFiles[index], ast: syntax})
		}
	}
	slices.SortFunc(files, func(left, right sourceFile) int { return strings.Compare(left.path, right.path) })

	for _, file := range files {
		fileID := FileID(len(b.graph.Files) + 1)
		pkgID := b.packageID(file.pkg.Types)
		b.graph.Files[fileID] = &File{ID: fileID, Path: file.path, Package: pkgID}
		for _, declaration := range file.ast.Decls {
			function, ok := declaration.(*ast.FuncDecl)
			if !ok || function.Body == nil {
				continue
			}
			current := &sourceFunction{
				pkg:  file.pkg,
				body: function.Body,
				file: fileID,
				pos:  function.Pos(),
				end:  function.End(),
				kind: SyntaxFuncDecl,
				name: function.Name.Name,
			}
			current.object, _ = file.pkg.TypesInfo.Defs[function.Name].(*types.Func)
			b.addFunction(current, nil)
			b.collectLiterals(file.pkg, fileID, current, function.Body)
		}
		// Package-level function literals are not descendants of a declaration.
		for _, declaration := range file.ast.Decls {
			if _, ok := declaration.(*ast.FuncDecl); !ok {
				b.collectLiterals(file.pkg, fileID, nil, declaration)
			}
		}
	}
	return nil
}

func (b *graphBuilder) collectLiterals(pkg *packages.Package, file FileID, parent *sourceFunction, syntax ast.Node) {
	count := 0
	ast.Inspect(syntax, func(current ast.Node) bool {
		literal, ok := current.(*ast.FuncLit)
		if !ok {
			return true
		}
		count++
		name := "func$" + strconv.Itoa(count)
		if parent != nil {
			name = parent.name + "." + name
		}
		function := &sourceFunction{
			pkg: pkg, body: literal.Body, parent: parent,
			file: file, pos: literal.Pos(), end: literal.End(),
			kind: SyntaxFuncLit, name: name,
		}
		b.addFunction(function, literal)
		b.collectLiterals(pkg, file, function, literal.Body)
		return false
	})
}

func (b *graphBuilder) addFunction(function *sourceFunction, literal *ast.FuncLit) {
	id := b.nextNode
	b.nextNode++
	function.node = &Node{
		ID:   id,
		Name: function.name,
		Pkg:  b.packageID(function.pkg.Types),
		Syntax: model.Syntax{
			Kind:  function.kind,
			File:  function.file,
			Start: position(function.pkg.Fset, function.pos),
			End:   position(function.pkg.Fset, function.end),
		},
	}
	b.graph.Nodes[id] = function.node
	b.graph.Files[function.file].Functions = append(b.graph.Files[function.file].Functions, id)
	b.functions = append(b.functions, function)
	if function.object != nil {
		b.objectIDs[function.object] = id
	}
	if literal != nil {
		b.literalIDs[literal] = id
	}
	if function.parent != nil {
		parentID := function.parent.node.ID
		function.node.Parent = &parentID
		function.parent.node.Function.AnonFuncs = append(function.parent.node.Function.AnonFuncs, id)
	}
}

func (b *graphBuilder) connect() {
	for _, function := range b.functions {
		if function.pkg.PkgPath == b.rootPackage && function.name == "main" {
			b.graph.Root = function.node.ID
		}
		out := make(map[NodeID]struct{})
		references := make(map[NodeID]struct{})
		ast.Inspect(function.body, func(current ast.Node) bool {
			if literal, ok := current.(*ast.FuncLit); ok {
				if id, exists := b.literalIDs[literal]; exists {
					references[id] = struct{}{}
				}
				return false
			}
			switch current := current.(type) {
			case *ast.CallExpr:
				if id := b.callTarget(function.pkg.TypesInfo, current.Fun); id != 0 && id != function.node.ID {
					out[id] = struct{}{}
				}
			case *ast.Ident:
				if object, ok := function.pkg.TypesInfo.Uses[current].(*types.Func); ok {
					if id := b.functionID(object); id != 0 && id != function.node.ID {
						references[id] = struct{}{}
					}
				}
			case *ast.SelectorExpr:
				if selection := function.pkg.TypesInfo.Selections[current]; selection != nil {
					if object, ok := selection.Obj().(*types.Func); ok {
						if id := b.functionID(object); id != 0 && id != function.node.ID {
							references[id] = struct{}{}
						}
					}
				}
			}
			return true
		})
		function.node.Out = sortedIDs(out)
		function.node.Function.References = sortedIDs(references)
	}

	for _, node := range b.graph.Nodes {
		for _, callee := range node.Out {
			b.graph.Nodes[callee].In = append(b.graph.Nodes[callee].In, node.ID)
		}
	}
	for _, node := range b.graph.Nodes {
		slices.Sort(node.In)
		slices.Sort(node.Function.AnonFuncs)
	}
	b.connectFiles()
}

func (b *graphBuilder) callTarget(info *types.Info, expression ast.Expr) NodeID {
	switch expression := expression.(type) {
	case *ast.FuncLit:
		return b.literalIDs[expression]
	case *ast.Ident:
		function, _ := info.Uses[expression].(*types.Func)
		return b.functionID(function)
	case *ast.SelectorExpr:
		if selection := info.Selections[expression]; selection != nil {
			function, _ := selection.Obj().(*types.Func)
			return b.functionID(function)
		}
		function, _ := info.Uses[expression.Sel].(*types.Func)
		return b.functionID(function)
	case *ast.IndexExpr:
		return b.callTarget(info, expression.X)
	case *ast.IndexListExpr:
		return b.callTarget(info, expression.X)
	case *ast.ParenExpr:
		return b.callTarget(info, expression.X)
	}
	return 0
}

func (b *graphBuilder) functionID(function *types.Func) NodeID {
	if function == nil {
		return 0
	}
	origin := function.Origin()
	if id := b.objectIDs[origin]; id != 0 {
		b.objectIDs[function] = id
		return id
	}

	id := b.nextNode
	b.nextNode++
	pkgID := b.packageID(function.Pkg())
	b.graph.Nodes[id] = &Node{ID: id, Name: function.Name(), Pkg: pkgID}
	b.objectIDs[origin] = id
	b.objectIDs[function] = id
	return id
}

func (b *graphBuilder) packageID(pkg *types.Package) PackageID {
	if pkg == nil {
		return 0
	}
	if id := b.packageIDs[pkg]; id != 0 {
		return id
	}
	id := b.nextPackage
	b.nextPackage++
	b.packageIDs[pkg] = id
	b.graph.Packages[id] = &model.Package{Path: pkg.Path(), Name: pkg.Name()}
	return id
}

func (b *graphBuilder) connectFiles() {
	calls := make(map[FileID]map[FileID]struct{})
	calledBy := make(map[FileID]map[FileID]struct{})
	for _, node := range b.graph.Nodes {
		from := node.Syntax.File
		if from == 0 {
			continue
		}
		for _, callee := range node.Out {
			to := b.graph.Nodes[callee].Syntax.File
			if to == 0 || to == from {
				continue
			}
			if calls[from] == nil {
				calls[from] = make(map[FileID]struct{})
			}
			if calledBy[to] == nil {
				calledBy[to] = make(map[FileID]struct{})
			}
			calls[from][to] = struct{}{}
			calledBy[to][from] = struct{}{}
		}
	}
	for id, file := range b.graph.Files {
		file.Calls = sortedFileIDs(calls[id])
		file.CalledBy = sortedFileIDs(calledBy[id])
		slices.Sort(file.Functions)
	}
}

func position(files *token.FileSet, pos token.Pos) model.Position {
	if pos == token.NoPos {
		return model.Position{}
	}
	value := files.Position(pos)
	return model.Position{Line: value.Line, Column: value.Column}
}

func sortedIDs(values map[NodeID]struct{}) []NodeID {
	result := make([]NodeID, 0, len(values))
	for id := range values {
		result = append(result, id)
	}
	slices.Sort(result)
	return result
}

func sortedFileIDs(values map[FileID]struct{}) []FileID {
	result := make([]FileID, 0, len(values))
	for id := range values {
		result = append(result, id)
	}
	slices.Sort(result)
	return result
}
