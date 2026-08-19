package model

// PackageID uniquely identifies a package within a static graph.
type PackageID uint64

// NodeID uniquely identifies a function node within a static graph.
type NodeID uint64

// FileID uniquely identifies a source file within a static graph.
type FileID uint64

// SyntaxKind identifies the source syntax that defines a function node.
type SyntaxKind uint8

const (
	// SyntaxFuncDecl represents a declared function or method.
	SyntaxFuncDecl SyntaxKind = iota
	// SyntaxFuncLit represents an anonymous function literal.
	SyntaxFuncLit
)

// StaticGraph is sen's language-neutral representation of analyzed code.
type StaticGraph struct {
	Root     NodeID
	Nodes    map[NodeID]*StaticNode
	Files    map[FileID]*StaticFile
	Packages map[PackageID]*Package
}

// StaticNode contains the source identity and relationships used at runtime.
type StaticNode struct {
	Name   string
	ID     NodeID
	Syntax Syntax
	Parent *NodeID
	Pkg    PackageID

	Function FunctionBody
	In       []NodeID
	Out      []NodeID
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
}

// Package describes an analyzed source package.
type Package struct {
	Path string
	Name string
}

// FunctionBody contains relationships not represented by direct calls.
type FunctionBody struct {
	AnonFuncs  []NodeID
	References []NodeID
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
