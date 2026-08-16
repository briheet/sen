package graph

import (
	"github.com/briheet/senbon/internal/analysis"
	"golang.org/x/tools/go/callgraph/rta"
	"golang.org/x/tools/go/packages"
)

type PackageID uint64

type NodeID uint64
type FileID uint64

type BlockID uint64

type SyntaxKind uint8

const (
	SyntaxFuncDecl SyntaxKind = iota
	SyntaxFuncLit
	SyntaxRange
)

// This is the Internal Representation of Graph
type Graph struct {
	Root     NodeID                 // Root Node of the graph
	Nodes    map[NodeID]*Node       // Functions to functions based Node mapping
	Files    map[FileID]*File       // Files having function metadata and files to files mapping
	Packages map[PackageID]*Package // Program packages
	Program  Program                // enclosing program
}

// A Node is a function in a particular file
// It has all metadata of files and other files originating
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

// main go program
type Program struct {
	Files      []FileID
	Packages   []PackageID
	BuildMode  string
	MethodSets []MethodSet
}

type MethodSet struct {
	Type    string
	Methods []NodeID
}

// ast.Node
type Syntax struct {
	Kind  SyntaxKind
	File  FileID
	Start Position
	End   Position
}

// tok.pos
type Position struct {
	Line   int
	Column int
	Offset int
}

type ObjectFunc struct {
	Name   string      // FullName of the package or reciever
	Origin *ObjectFunc // Canonical func for its reciever
	Pkg    PackageID   // Package which this function belongs to
}

// Object signature type
type Signature struct {
	Receiver   string   // If method or nil
	Params     []Param  // Parameters of a signature
	Results    []Param  // results of a signatire
	Variadic   bool     // Vardic or not
	TypeParams []string // Type params

}

// Base params
type Param struct {
	Name string
	Type string
}

// Metadata for go's package
type Package struct {
	Path string
	Name string
}

// type annotations
type TypeInfo struct {
	Type      string
	Object    string
	Package   string
	Selection *Selection
}

type Selection struct {
	Kind     string
	Receiver string
	Object   string
	Indirect bool
	Index    []int
}

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

type Variable struct {
	Name string
	Type string
}

type TypeParam struct {
	Name       string
	Constraint string
}

type Type struct {
	Name string
}

type Block struct {
	ID           BlockID
	Index        int
	Comment      string
	Instructions []Instruction
	Pred         []BlockID
	Succ         []BlockID
}

type Instruction struct {
	Op       string
	Position Position
}

// A file is a normal file in the project
// It is mapped via calle function based mapping
type File struct {
	ID        FileID    // fileID
	Path      string    // path to file
	Package   PackageID // packages in the file
	Functions []NodeID  // functions and methods present in the file

	Calls    []FileID // which file has called
	CalledBy []FileID // who is this file calling
}

func GetGraph(pkgs []*packages.Package) (*Graph, error) {
	// build ssa package and get main
	mainFunc, err := analysis.BuildPackagesAndReturnMain(pkgs)
	if err != nil {
		return nil, err
	}

	// Get rta result. No reflection
	// TODO: Get back to here
	result := analysis.BuildCallgraph(mainFunc)

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

	nodeIDs := make(map[*callgraph.Node]NodeID)

	return graph
}
