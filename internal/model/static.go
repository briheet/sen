package model

// PackageID uniquely identifies a package within a static graph.
type PackageID uint64

// NodeID uniquely identifies a function node within a static graph.
type NodeID uint64

// FileID uniquely identifies a source file within a static graph.
type FileID uint64

// BlockID identifies a basic block within its function.
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

// StaticGraph is sen's language-neutral representation of analyzed code.
type StaticGraph struct {
	Root     NodeID
	Nodes    map[NodeID]*StaticNode
	Files    map[FileID]*StaticFile
	Packages map[PackageID]*Package
	Program  Program
}

// StaticNode represents a function and its call and control-flow metadata.
type StaticNode struct {
	Name      string
	ID        NodeID
	Object    *ObjectFunc
	Signature Signature
	Synthetic string

	Syntax    Syntax
	Info      TypeInfo
	GoVersion string

	Parent *NodeID
	Pkg    PackageID

	Function FunctionBody
	In       []NodeID
	Out      []NodeID
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

// ObjectFunc describes the source object behind a function.
type ObjectFunc struct {
	Name   string
	Origin *ObjectFunc
	Pkg    PackageID
}

// Signature describes a function or method signature.
type Signature struct {
	Receiver   string
	Params     []Param
	Results    []Param
	Variadic   bool
	TypeParams []string
}

// Param describes a named parameter or result.
type Param struct {
	Name string
	Type string
}

// Package describes an analyzed source package.
type Package struct {
	Path string
	Name string
}

// TypeInfo contains type metadata associated with a node.
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

// FunctionBody contains a function's body and generic metadata.
type FunctionBody struct {
	FreeVars   []Variable
	Locals     []Variable
	Blocks     []Block
	Recover    *BlockID
	AnonFuncs  []NodeID
	References []NodeID

	TypeArgs       []Type
	RecvTypeParams []TypeParam
	RecvTypeArgs   []Type

	Origin *NodeID
}

// Variable describes a named variable and its type.
type Variable struct {
	Name string
	Type string
}

// TypeParam describes a generic type parameter and its constraint.
type TypeParam struct {
	Name       string
	Constraint string
}

// Type contains the string representation of a source type.
type Type struct {
	Name string
}

// Block represents a basic block and its control-flow edges.
type Block struct {
	ID           BlockID
	Index        int
	Comment      string
	Instructions []Instruction
	Pred         []BlockID
	Succ         []BlockID
}

// Instruction describes an instruction and its source position.
type Instruction struct {
	Op       string
	Position Position
}

// StaticFile describes a source file and its call relationships.
type StaticFile struct {
	ID        FileID
	Path      string
	Package   PackageID
	Functions []NodeID
	Calls     []FileID
	CalledBy  []FileID
}

// Node returns the node with the given ID.
func (g *StaticGraph) Node(id NodeID) *StaticNode {
	return g.Nodes[id]
}

// Children returns the nodes called by the given node.
func (g *StaticGraph) Children(id NodeID) []*StaticNode {
	node := g.Node(id)
	if node == nil {
		return nil
	}
	children := make([]*StaticNode, 0, len(node.Out))
	for _, childID := range node.Out {
		if child := g.Node(childID); child != nil {
			children = append(children, child)
		}
	}
	return children
}

// Parents returns the nodes that call the given node.
func (g *StaticGraph) Parents(id NodeID) []*StaticNode {
	node := g.Node(id)
	if node == nil {
		return nil
	}
	parents := make([]*StaticNode, 0, len(node.In))
	for _, parentID := range node.In {
		if parent := g.Node(parentID); parent != nil {
			parents = append(parents, parent)
		}
	}
	return parents
}

// FileFunctions returns the function nodes declared in the given file.
func (g *StaticGraph) FileFunctions(id FileID) []*StaticNode {
	file := g.Files[id]
	if file == nil {
		return nil
	}
	functions := make([]*StaticNode, 0, len(file.Functions))
	for _, functionID := range file.Functions {
		if function := g.Node(functionID); function != nil {
			functions = append(functions, function)
		}
	}
	return functions
}
