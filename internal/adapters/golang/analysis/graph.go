package analysis

import (
	"go/ast"
	"go/token"
	"go/types"
	"slices"
	"sort"

	"github.com/briheet/senbon/internal/model"
	"golang.org/x/tools/go/callgraph"
	"golang.org/x/tools/go/callgraph/rta"
	"golang.org/x/tools/go/packages"
	"golang.org/x/tools/go/ssa"
)

type (
	Graph        = model.StaticGraph
	Node         = model.StaticNode
	File         = model.StaticFile
	PackageID    = model.PackageID
	NodeID       = model.NodeID
	FileID       = model.FileID
	BlockID      = model.BlockID
	SyntaxKind   = model.SyntaxKind
	Position     = model.Position
	ObjectFunc   = model.ObjectFunc
	Signature    = model.Signature
	Param        = model.Param
	Package      = model.Package
	TypeInfo     = model.TypeInfo
	Selection    = model.Selection
	FunctionBody = model.FunctionBody
	Variable     = model.Variable
	TypeParam    = model.TypeParam
	Type         = model.Type
	Block        = model.Block
	Instruction  = model.Instruction
	Syntax       = model.Syntax
	Program      = model.Program
	MethodSet    = model.MethodSet
)

const (
	SyntaxFuncDecl = model.SyntaxFuncDecl
	SyntaxFuncLit  = model.SyntaxFuncLit
	SyntaxRange    = model.SyntaxRange
)

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
