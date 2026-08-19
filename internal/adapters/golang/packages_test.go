package golang

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/briheet/sen/internal/adapters/golang/analysis"
	"github.com/stretchr/testify/require"
)

func TestLoadHTTPExample(t *testing.T) {
	pkgs, err := LoadPackages(context.Background(), "../../../examples/go/http", nil)
	require.NoError(t, err)
	require.Len(t, pkgs, 1)
	require.False(t, pkgs[0].IllTyped)
}

func TestLoadPackagesParsesOnlyLocalImportClosure(t *testing.T) {
	root := t.TempDir()
	require.NoError(t, os.MkdirAll(filepath.Join(root, "helper"), 0o700))
	require.NoError(t, os.WriteFile(filepath.Join(root, "go.mod"), []byte("module example.com/app\n\ngo 1.25\n"), 0o600))
	require.NoError(t, os.WriteFile(filepath.Join(root, "main.go"), []byte(`package main

import (
	"fmt"
	"example.com/app/helper"
)

func main() { fmt.Println(helper.Value()) }
`), 0o600))
	require.NoError(t, os.WriteFile(filepath.Join(root, "helper", "helper.go"), []byte(`package helper

func Value() int { return 1 }
`), 0o600))

	pkgs, err := LoadPackages(context.Background(), root, nil)
	require.NoError(t, err)
	require.Len(t, pkgs, 2)
	require.Equal(t, "example.com/app", pkgs[0].PkgPath)
	for _, pkg := range pkgs {
		require.NotEmpty(t, pkg.Syntax)
		require.Equal(t, "example.com/app", pkg.Module.Path)
	}

	graph, err := analysis.GetGraph(pkgs)
	require.NoError(t, err)
	var mainID, helperID analysis.NodeID
	for id, node := range graph.Nodes {
		switch node.Name {
		case "main":
			mainID = id
		case "Value":
			if node.Syntax.File != 0 {
				helperID = id
			}
		}
	}
	require.NotZero(t, mainID)
	require.NotZero(t, helperID)
	require.Contains(t, graph.Nodes[mainID].Out, helperID)
}
