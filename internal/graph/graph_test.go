package graph

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

type runner interface { Run(int) int }
type worker struct{}

func (worker) Run(value int) int { return helper(value) }

func main() {
	helper(1)
	helper(2)
	var r runner = worker{}
	r.Run(3)
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

	pkgs, err := packages.Load(&packages.Config{Mode: packages.LoadAllSyntax, Dir: dir}, ".")
	require.NoError(t, err)
	require.Len(t, pkgs, 1)
	require.Empty(t, pkgs[0].Errors)

	parsed, err := GetGraph(pkgs)
	require.NoError(t, err)
	require.NotEmpty(t, parsed.Nodes)
	require.Len(t, parsed.Packages, 1)
	require.Len(t, parsed.Files, 2)

	byName := make(map[string]*Node)
	var runNode *Node
	for _, node := range parsed.Nodes {
		byName[node.Name] = node
		if node.Name == "Run" && node.Signature.Receiver == "graphfixture.worker" {
			runNode = node
		}
	}
	mainNode := byName["main"]
	helperNode := byName["helper"]
	require.NotNil(t, mainNode)
	require.NotNil(t, helperNode)
	require.NotNil(t, runNode)
	require.Equal(t, mainNode.ID, parsed.Root)
	require.Equal(t, "graphfixture.worker", runNode.Signature.Receiver)
	require.Equal(t, "int", runNode.Signature.Params[0].Type)
	require.NotEmpty(t, helperNode.Function.Blocks)
	require.NotEmpty(t, helperNode.Function.Blocks[0].Instructions)
	require.Positive(t, helperNode.Syntax.Start.Line)
	require.NotEmpty(t, helperNode.GoVersion)

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
	require.Equal(t, []FileID{helperFile.ID}, mainFile.Calls)
	require.Equal(t, []FileID{mainFile.ID}, helperFile.CalledBy)

	methodFound := false
	for _, methodSet := range parsed.Program.MethodSets {
		if methodSet.Type == "graphfixture.worker" {
			require.Contains(t, methodSet.Methods, runNode.ID)
			methodFound = true
		}
	}
	require.True(t, methodFound)
}
