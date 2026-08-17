package analysis

import (
	"go/ast"
	"go/token"
	"go/types"
	"slices"
	"sort"

	"golang.org/x/tools/go/callgraph"
	"golang.org/x/tools/go/callgraph/rta"
	"golang.org/x/tools/go/packages"
	"golang.org/x/tools/go/ssa"
)

// PackageID uniquely identifies a package within a Graph.
type PackageID uint64

// NodeID uniquely identifies a function node within a Graph.
type NodeID uint64

// FileID uniquely identifies a source file within a Graph.
type FileID uint64

// BlockID identifies an SSA basic block within its function.
type BlockID uint64

// SyntaxKind identifies the source syntax that defines a function node.
type SyntaxKind uint8

const (
	// SyntaxFuncDecl represents a declared function or method.
	SyntaxFuncDecl SyntaxKind = iota
	// SyntaxFuncLit represents an anonymous function literal.
	SyntaxFuncLit
	// SyntaxRange represents a synthetic function created for a range statement.
	SyntaxRange
)

// Graph is Senbon's internal representation of an analyzed program.
type Graph struct {
	Root     NodeID                 // Root Node of the graph
	Nodes    map[NodeID]*Node       // Functions to functions based Node mapping
	Files    map[FileID]*File       // Files having function metadata and files to files mapping
	Packages map[PackageID]*Package // Program packages
	Program  Program                // enclosing program
}

// Node represents a function or method and its call and control-flow metadata.
type Node struct {
	Name      string      // Function based infos
	ID        NodeID      // Node ID
	Object    *ObjectFunc // Base objects func
	Signature Signature   // Function sig
	Synthetic string      // SOurce info if synthetic fuinction

	Syntax    Syntax   // Ast stuff
	Info      TypeInfo // type annotations
	GoVersion string

	Parent *NodeID   // enclosing function if so
	Pkg    PackageID // enclosing package

	Function FunctionBody // Functionbody fields

	In  []NodeID // incoming edges
	Out []NodeID // outgoing edges
}

// Program describes the files, packages, and method sets in the graph.
type Program struct {
	Files      []FileID
	Packages   []PackageID
	BuildMode  string
	MethodSets []MethodSet
}

// MethodSet associates a runtime type with its reachable methods.
type MethodSet struct {
	Type    string
	Methods []NodeID
}

// Syntax describes the source syntax that produced a function node.
type Syntax struct {
	Kind  SyntaxKind
	File  FileID
	Start Position
	End   Position
}

// Position identifies a location in a source file.
type Position struct {
	Line   int
	Column int
	Offset int
}

// ObjectFunc describes the type-checker object behind a function.
type ObjectFunc struct {
	Name   string      // FullName of the package or reciever
	Origin *ObjectFunc // Canonical func for its reciever
	Pkg    PackageID   // Package which this function belongs to
}

// Signature describes a function or method signature.
type Signature struct {
	Receiver   string   // If method or nil
	Params     []Param  // Parameters of a signature
	Results    []Param  // results of a signatire
	Variadic   bool     // Vardic or not
	TypeParams []string // Type params

}

// Param describes a named parameter or result.
type Param struct {
	Name string
	Type string
}

// Package describes an analyzed Go package.
type Package struct {
	Path string
	Name string
}

// TypeInfo contains type-checker metadata associated with a node.
type TypeInfo struct {
	Type      string
	Object    string
	Package   string
	Selection *Selection
}

// Selection describes a field or method selection.
type Selection struct {
	Kind     string
	Receiver string
	Object   string
	Indirect bool
	Index    []int
}

// FunctionBody contains the SSA body and generic metadata of a function.
type FunctionBody struct {
	FreeVars  []Variable // closure supplied vcars
	Locals    []Variable // frame allocated vars
	Blocks    []Block    // basic blocks of function
	Recover   *BlockID
	AnonFuncs []NodeID // as the name suggests

	TypeArgs       []Type
	RecvTypeParams []TypeParam
	RecvTypeArgs   []Type

	Origin *NodeID
}

// Variable describes a named SSA variable and its type.
type Variable struct {
	Name string
	Type string
}

// TypeParam describes a generic type parameter and its constraint.
type TypeParam struct {
	Name       string
	Constraint string
}

// Type contains the string representation of a Go type.
type Type struct {
	Name string
}

// Block represents an SSA basic block and its control-flow edges.
type Block struct {
	ID           BlockID
	Index        int
	Comment      string
	Instructions []Instruction
	Pred         []BlockID
	Succ         []BlockID
}

// Instruction describes an SSA instruction and its source position.
type Instruction struct {
	Op       string
	Position Position
}

// File describes a source file, its functions, and cross-file call relationships.
type File struct {
	ID        FileID    // fileID
	Path      string    // path to file
	Package   PackageID // packages in the file
	Functions []NodeID  // functions and methods present in the file

	Calls    []FileID // which file has called
	CalledBy []FileID // who is this file calling
}

// Node returns the node with the given ID, or nil if it does not exist.
func (g *Graph) Node(id NodeID) *Node {
	return g.Nodes[id]
}

// Children returns the nodes called by the given node.
func (g *Graph) Children(id NodeID) []*Node {
	node := g.Node(id)
	if node == nil {
		return nil
	}

	children := make([]*Node, 0, len(node.Out))
	for _, childID := range node.Out {
		if child := g.Node(childID); child != nil {
			children = append(children, child)
		}
	}
	return children
}

// Parents returns the nodes that call the given node.
func (g *Graph) Parents(id NodeID) []*Node {
	node := g.Node(id)
	if node == nil {
		return nil
	}

	parents := make([]*Node, 0, len(node.In))
	for _, parentID := range node.In {
		if parent := g.Node(parentID); parent != nil {
			parents = append(parents, parent)
		}
	}
	return parents
}

// FileFunctions returns the function nodes declared in the given file.
func (g *Graph) FileFunctions(id FileID) []*Node {
	file := g.Files[id]
	if file == nil {
		return nil
	}

	functions := make([]*Node, 0, len(file.Functions))
	for _, functionID := range file.Functions {
		if function := g.Node(functionID); function != nil {
			functions = append(functions, function)
		}
	}
	return functions
}

func GetGraph(pkgs []*packages.Package) (*Graph, error) {
	// build ssa package and get main
	mainFunc, err := BuildPackagesAndReturnMain(pkgs)
	if err != nil {
		return nil, err
	}

	// Get rta result. No reflection
	// TODO: Get back to here
	result := BuildCallgraph(mainFunc)

	// Build graph
	graph := buildGraph(result)
	return graph, nil
}

// Take in result and build graph
func buildGraph(result *rta.Result) *Graph {
	graph := &Graph{
		Nodes:    make(map[NodeID]*Node),
		Files:    make(map[FileID]*File),
		Packages: make(map[PackageID]*Package),
	}

	reachable := make([]*ssa.Function, 0, len(result.Reachable))
	for fn := range result.Reachable {
		reachable = append(reachable, fn)
	}
	sort.Slice(reachable, func(i, j int) bool {
		left := reachable[i].Prog.Fset.Position(reachable[i].Pos())
		right := reachable[j].Prog.Fset.Position(reachable[j].Pos())
		if left.Filename != right.Filename {
			return left.Filename < right.Filename
		}
		if left.Offset != right.Offset {
			return left.Offset < right.Offset
		}
		return reachable[i].String() < reachable[j].String()
	})
	for _, fn := range reachable {
		result.CallGraph.CreateNode(fn)
	}

	nodes := make([]*callgraph.Node, 0, len(result.CallGraph.Nodes))
	for _, node := range result.CallGraph.Nodes {
		if node.Func != nil {
			nodes = append(nodes, node)
		}
	}
	sort.Slice(nodes, func(i, j int) bool { return nodes[i].ID < nodes[j].ID })

	if result.CallGraph.Root != nil {
		graph.Root = NodeID(result.CallGraph.Root.ID)
	}

	qualifier := func(pkg *types.Package) string {
		if pkg == nil {
			return ""
		}
		return pkg.Path()
	}
	typeString := func(typ types.Type) string {
		if typ == nil {
			return ""
		}
		return types.TypeString(typ, qualifier)
	}
	position := func(fset *token.FileSet, pos token.Pos) Position {
		if pos == token.NoPos {
			return Position{}
		}
		p := fset.Position(pos)
		return Position{Line: p.Line, Column: p.Column, Offset: p.Offset}
	}

	packageIDs := make(map[*types.Package]PackageID)
	fileIDs := make(map[string]FileID)
	nextPackageID := PackageID(1)
	nextFileID := FileID(1)
	packageID := func(pkg *types.Package) PackageID {
		if pkg == nil {
			return 0
		}
		if id, ok := packageIDs[pkg]; ok {
			return id
		}
		id := nextPackageID
		nextPackageID++
		packageIDs[pkg] = id
		graph.Packages[id] = &Package{Path: pkg.Path(), Name: pkg.Name()}
		graph.Program.Packages = append(graph.Program.Packages, id)
		return id
	}
	fileID := func(path string, pkgID PackageID) FileID {
		if path == "" {
			return 0
		}
		if id, ok := fileIDs[path]; ok {
			return id
		}
		id := nextFileID
		nextFileID++
		fileIDs[path] = id
		graph.Files[id] = &File{ID: id, Path: path, Package: pkgID}
		graph.Program.Files = append(graph.Program.Files, id)
		return id
	}
	functionPackage := func(fn *ssa.Function) *types.Package {
		if fn.Pkg != nil {
			return fn.Pkg.Pkg
		}
		if object := fn.Object(); object != nil {
			return object.Pkg()
		}
		if origin := fn.Origin(); origin != nil && origin.Pkg != nil {
			return origin.Pkg.Pkg
		}
		return nil
	}

	functionIDs := make(map[*ssa.Function]NodeID, len(nodes))
	functionFiles := make(map[*ssa.Function]FileID, len(nodes))
	for _, callNode := range nodes {
		fn := callNode.Func
		id := NodeID(callNode.ID)
		functionIDs[fn] = id

		pkg := functionPackage(fn)
		pkgID := packageID(pkg)
		node := &Node{
			Name:      fn.Name(),
			ID:        id,
			Synthetic: fn.Synthetic,
			Pkg:       pkgID,
			Info: TypeInfo{
				Type: typeString(fn.Type()),
			},
		}
		if pkg != nil {
			node.GoVersion = pkg.GoVersion()
			node.Info.Package = pkg.Path()
		}
		if object, ok := fn.Object().(*types.Func); ok {
			node.Info.Object = types.ObjectString(object, qualifier)
			node.Object = &ObjectFunc{Name: object.FullName(), Pkg: packageID(object.Pkg())}
			if origin := object.Origin(); origin != object {
				node.Object.Origin = &ObjectFunc{Name: origin.FullName(), Pkg: packageID(origin.Pkg())}
			}
		}

		sig := fn.Signature
		if recv := sig.Recv(); recv != nil {
			node.Signature.Receiver = typeString(recv.Type())
		}
		for i := range sig.Params().Len() {
			param := sig.Params().At(i)
			node.Signature.Params = append(node.Signature.Params, Param{Name: param.Name(), Type: typeString(param.Type())})
		}
		for i := range sig.Results().Len() {
			result := sig.Results().At(i)
			node.Signature.Results = append(node.Signature.Results, Param{Name: result.Name(), Type: typeString(result.Type())})
		}
		node.Signature.Variadic = sig.Variadic()
		for i := range sig.TypeParams().Len() {
			node.Signature.TypeParams = append(node.Signature.TypeParams, sig.TypeParams().At(i).Obj().Name())
		}

		for _, freeVar := range fn.FreeVars {
			node.Function.FreeVars = append(node.Function.FreeVars, Variable{Name: freeVar.Name(), Type: typeString(freeVar.Type())})
		}
		for _, local := range fn.Locals {
			localType := local.Type()
			if pointer, ok := localType.(*types.Pointer); ok {
				localType = pointer.Elem()
			}
			name := local.Comment
			if name == "" {
				name = local.Name()
			}
			node.Function.Locals = append(node.Function.Locals, Variable{Name: name, Type: typeString(localType)})
		}
		for _, block := range fn.Blocks {
			parsed := Block{ID: BlockID(block.Index), Index: block.Index, Comment: block.Comment}
			for _, instruction := range block.Instrs {
				parsed.Instructions = append(parsed.Instructions, Instruction{
					Op:       instruction.String(),
					Position: position(fn.Prog.Fset, instruction.Pos()),
				})
			}
			for _, pred := range block.Preds {
				parsed.Pred = append(parsed.Pred, BlockID(pred.Index))
			}
			for _, succ := range block.Succs {
				parsed.Succ = append(parsed.Succ, BlockID(succ.Index))
			}
			node.Function.Blocks = append(node.Function.Blocks, parsed)
		}
		if fn.Recover != nil {
			recoverID := BlockID(fn.Recover.Index)
			node.Function.Recover = &recoverID
		}
		for i := range sig.RecvTypeParams().Len() {
			param := sig.RecvTypeParams().At(i)
			node.Function.RecvTypeParams = append(node.Function.RecvTypeParams, TypeParam{
				Name:       param.Obj().Name(),
				Constraint: typeString(param.Constraint()),
			})
		}
		typeArgs := fn.TypeArgs()
		recvArgs := sig.RecvTypeParams().Len()
		if recvArgs > len(typeArgs) {
			recvArgs = len(typeArgs)
		}
		for i, arg := range typeArgs {
			parsed := Type{Name: typeString(arg)}
			if i < recvArgs {
				node.Function.RecvTypeArgs = append(node.Function.RecvTypeArgs, parsed)
			} else {
				node.Function.TypeArgs = append(node.Function.TypeArgs, parsed)
			}
		}

		syntax := fn.Syntax()
		pos := fn.Pos()
		if syntax != nil {
			pos = syntax.Pos()
			node.Syntax.Start = position(fn.Prog.Fset, syntax.Pos())
			node.Syntax.End = position(fn.Prog.Fset, syntax.End())
			switch syntax.(type) {
			case *ast.FuncDecl:
				node.Syntax.Kind = SyntaxFuncDecl
			case *ast.FuncLit:
				node.Syntax.Kind = SyntaxFuncLit
			case *ast.RangeStmt:
				node.Syntax.Kind = SyntaxRange
			}
		}
		file := fn.Prog.Fset.Position(pos).Filename
		node.Syntax.File = fileID(file, pkgID)
		functionFiles[fn] = node.Syntax.File
		if node.Syntax.File != 0 {
			graph.Files[node.Syntax.File].Functions = append(graph.Files[node.Syntax.File].Functions, id)
		}
		graph.Nodes[id] = node
	}

	for _, callNode := range nodes {
		node := graph.Nodes[functionIDs[callNode.Func]]
		if parentID, ok := functionIDs[callNode.Func.Parent()]; ok {
			id := parentID
			node.Parent = &id
		}
		if originID, ok := functionIDs[callNode.Func.Origin()]; ok {
			id := originID
			node.Function.Origin = &id
		}
		for _, anon := range callNode.Func.AnonFuncs {
			if id, ok := functionIDs[anon]; ok {
				node.Function.AnonFuncs = append(node.Function.AnonFuncs, id)
			}
		}
		slices.Sort(node.Function.AnonFuncs)
	}

	in := make(map[NodeID]map[NodeID]struct{}, len(nodes))
	out := make(map[NodeID]map[NodeID]struct{}, len(nodes))
	fileCalls := make(map[FileID]map[FileID]struct{})
	fileCalledBy := make(map[FileID]map[FileID]struct{})
	for _, callNode := range nodes {
		callerID := functionIDs[callNode.Func]
		for _, edge := range callNode.Out {
			calleeID, ok := functionIDs[edge.Callee.Func]
			if !ok {
				continue
			}
			if out[callerID] == nil {
				out[callerID] = make(map[NodeID]struct{})
			}
			if in[calleeID] == nil {
				in[calleeID] = make(map[NodeID]struct{})
			}
			out[callerID][calleeID] = struct{}{}
			in[calleeID][callerID] = struct{}{}

			callerFile := functionFiles[callNode.Func]
			calleeFile := functionFiles[edge.Callee.Func]
			if callerFile == 0 || calleeFile == 0 || callerFile == calleeFile {
				continue
			}
			if fileCalls[callerFile] == nil {
				fileCalls[callerFile] = make(map[FileID]struct{})
			}
			if fileCalledBy[calleeFile] == nil {
				fileCalledBy[calleeFile] = make(map[FileID]struct{})
			}
			fileCalls[callerFile][calleeFile] = struct{}{}
			fileCalledBy[calleeFile][callerFile] = struct{}{}
		}
	}
	for id, node := range graph.Nodes {
		for caller := range in[id] {
			node.In = append(node.In, caller)
		}
		for callee := range out[id] {
			node.Out = append(node.Out, callee)
		}
		slices.Sort(node.In)
		slices.Sort(node.Out)
	}
	for id, file := range graph.Files {
		for called := range fileCalls[id] {
			file.Calls = append(file.Calls, called)
		}
		for caller := range fileCalledBy[id] {
			file.CalledBy = append(file.CalledBy, caller)
		}
		slices.Sort(file.Functions)
		slices.Sort(file.Calls)
		slices.Sort(file.CalledBy)
	}

	runtimeTypes := result.RuntimeTypes.Keys()
	sort.Slice(runtimeTypes, func(i, j int) bool { return typeString(runtimeTypes[i]) < typeString(runtimeTypes[j]) })
	if result.CallGraph.Root != nil && result.CallGraph.Root.Func != nil {
		prog := result.CallGraph.Root.Func.Prog
		for _, runtimeType := range runtimeTypes {
			parsed := MethodSet{Type: typeString(runtimeType)}
			seen := make(map[NodeID]struct{})
			for selection := range prog.MethodSets.MethodSet(runtimeType).Methods() {
				method := prog.MethodValue(selection)
				if id, ok := functionIDs[method]; ok {
					seen[id] = struct{}{}
				}
			}
			for id := range seen {
				parsed.Methods = append(parsed.Methods, id)
			}
			slices.Sort(parsed.Methods)
			graph.Program.MethodSets = append(graph.Program.MethodSets, parsed)
		}
	}

	return graph
}
