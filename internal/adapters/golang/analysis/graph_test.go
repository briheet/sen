package analysis

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
	"golang.org/x/tools/go/packages"
)

func TestGraphTraversal(t *testing.T) {
	parent := &Node{ID: 1, Out: []NodeID{2}}
	child := &Node{ID: 2, In: []NodeID{1}}
	graph := &Graph{
		Nodes: map[NodeID]*Node{1: parent, 2: child},
		Files: map[FileID]*File{1: {ID: 1, Functions: []NodeID{1, 2}}},
	}

	require.Same(t, parent, graph.Node(1))
	require.Equal(t, []*Node{child}, graph.Children(1))
	require.Equal(t, []*Node{parent}, graph.Parents(2))
	require.Equal(t, []*Node{parent, child}, graph.FileFunctions(1))
	require.Nil(t, graph.Node(99))
	require.Nil(t, graph.Children(99))
	require.Nil(t, graph.Parents(99))
	require.Nil(t, graph.FileFunctions(99))
}

func TestBuildGraph(t *testing.T) {
	dir := t.TempDir()
	files := map[string]string{
		"go.mod": "module graphfixture\n\ngo 1.25\n",
		"main.go": `package main

import "reflect"

type runner interface { Run(int) int }
type worker struct{}
type reflected struct{}

func (worker) Run(value int) int { return helper(value) }
func (reflected) Reflected() {}
func dead() {}
func callback() {}
func use(callback func()) { callback() }

func main() {
	helper(1)
	helper(2)
	use(callback)
	var r runner = worker{}
	r.Run(3)
	reflect.ValueOf(reflected{}).MethodByName("Reflected").Call(nil)
	captured := 4
	func() { helper(captured) }()
}
`,
		"helper.go": `package main

func helper(value int) int {
	if value > 0 {
		return value + 1
	}
	return value
}
`,
	}
	for name, contents := range files {
		require.NoError(t, os.WriteFile(filepath.Join(dir, name), []byte(contents), 0o600))
	}

	pkgs, err := packages.Load(&packages.Config{Mode: packages.LoadSyntax, Dir: dir}, ".")
	require.NoError(t, err)
	require.Len(t, pkgs, 1)
	require.Empty(t, pkgs[0].Errors)

	parsed, err := GetGraph(pkgs)
	require.NoError(t, err)
	require.NotEmpty(t, parsed.Nodes)
	require.Contains(t, parsed.Packages, parsed.Nodes[parsed.Root].Pkg)

	byName := make(map[string]*Node)
	var runNode *Node
	var reflectedNode *Node
	var deadNode *Node
	var externalNode *Node
	for _, node := range parsed.Nodes {
		pkg := parsed.Packages[node.Pkg]
		if node.Syntax.File != 0 {
			byName[node.Name] = node
		}
		if node.Name == "Run" && node.Syntax.File != 0 {
			runNode = node
		}
		if node.Name == "Reflected" && node.Syntax.File != 0 {
			reflectedNode = node
		}
		if node.Name == "dead" && node.Syntax.File != 0 {
			deadNode = node
		}
		if pkg != nil && pkg.Path == "reflect" {
			externalNode = node
		}
	}
	mainNode := byName["main"]
	helperNode := byName["helper"]
	callbackNode := byName["callback"]
	require.NotNil(t, mainNode)
	require.NotNil(t, helperNode)
	require.NotNil(t, callbackNode)
	require.NotNil(t, runNode)
	require.NotNil(t, reflectedNode)
	require.NotNil(t, deadNode)
	require.NotNil(t, externalNode)
	require.Equal(t, mainNode.ID, parsed.Root)
	require.Positive(t, helperNode.Syntax.Start.Line)

	helperCalls := 0
	for _, id := range mainNode.Out {
		if id == helperNode.ID {
			helperCalls++
		}
	}
	require.Equal(t, 1, helperCalls)

	require.Len(t, mainNode.Function.AnonFuncs, 1)
	closure := parsed.Nodes[mainNode.Function.AnonFuncs[0]]
	require.NotNil(t, closure)
	require.NotNil(t, closure.Parent)
	require.Equal(t, mainNode.ID, *closure.Parent)
	require.Equal(t, SyntaxFuncLit, closure.Syntax.Kind)
	require.Contains(t, mainNode.Function.References, callbackNode.ID)
	require.NotContains(t, mainNode.Out, runNode.ID)
	require.NotContains(t, mainNode.Out, reflectedNode.ID)

	var mainFile, helperFile *File
	for _, file := range parsed.Files {
		switch filepath.Base(file.Path) {
		case "main.go":
			mainFile = file
		case "helper.go":
			helperFile = file
		}
	}
	require.NotNil(t, mainFile)
	require.NotNil(t, helperFile)
	require.Contains(t, mainFile.Calls, helperFile.ID)
	require.Contains(t, helperFile.CalledBy, mainFile.ID)

}
