// Package analysis converts an OCaml project into Senbon's normalized graph.
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

	"github.com/briheet/senbon/internal/model"
)

//go:embed analyze.ml
var helperSource []byte

// Project holds the analyzed graph and the target's entry source file.
type Project struct {
	Graph *model.StaticGraph
	Entry string
}

// helperGraph mirrors the JSON emitted by analyze.ml.
type helperGraph struct {
	Entry     string       `json:"entry"`
	Functions []string     `json:"functions"`
	Edges     []helperEdge `json:"edges"`
}

// helperEdge is one resolved call edge.
type helperEdge struct {
	From uint64 `json:"from"`
	To   uint64 `json:"to"`
}

// Analyze runs the embedded compiler-libs analyzer for the entry source.
func Analyze(ctx context.Context, entryPath string) (*Project, error) {
	helperPath, err := writeHelper()
	if err != nil {
		return nil, err
	}
	defer func() { _ = os.Remove(helperPath) }()

	binary := filepath.Join(os.TempDir(), fmt.Sprintf("senbon-ocaml-analyze-%d", os.Getpid()))
	cmd := exec.CommandContext(ctx, "ocamlc", "-I", ocamlCompileInclude(), ocamlCommon(), helperPath, "-o", binary)
	if output, err := cmd.CombinedOutput(); err != nil {
		return nil, fmt.Errorf("compiling analyzer failed: %w: %s", err, output)
	}
	defer func() { _ = os.Remove(binary) }()

	outputPath := filepath.Join(os.TempDir(), fmt.Sprintf("senbon-ocaml-%d.json", os.Getpid()))
	defer func() { _ = os.Remove(outputPath) }()

	cmd = exec.CommandContext(ctx, binary, entryPath, outputPath)
	if output, err := cmd.CombinedOutput(); err != nil {
		return nil, fmt.Errorf("analysis helper failed: %w: %s", err, output)
	}
	data, err := os.ReadFile(outputPath)
	if err != nil {
		return nil, err
	}
	return parse(data, entryPath)
}

// parse builds a project from helper JSON output.
func parse(data []byte, entryPath string) (*Project, error) {
	var helper helperGraph
	if err := json.Unmarshal(data, &helper); err != nil {
		return nil, err
	}

	graph := &model.StaticGraph{
		Root:     model.NodeID(1), // first top-level binding is the entry
		Nodes:    make(map[model.NodeID]*model.StaticNode),
		Files:    make(map[model.FileID]*model.StaticFile),
		Packages: make(map[model.PackageID]*model.Package),
	}
	project := &Project{Graph: graph, Entry: helper.Entry}
	dir := filepath.Dir(entryPath)
	graph.Packages[1] = &model.Package{Path: dir, Name: filepath.Base(dir)}

	fileID := model.FileID(1)
	graph.Files[fileID] = &model.StaticFile{ID: fileID, Path: entryPath, Package: 1}
	graph.Program.Files = append(graph.Program.Files, fileID)

	// functions: id -> node; keep the ids from the analyzer
	for id, name := range helper.Functions {
		node := &model.StaticNode{
			Name:   name,
			ID:     model.NodeID(id) + 1, // model reserves 0 as sentinel
			Pkg:    1,
			Syntax: model.Syntax{File: fileID},
		}
		graph.Nodes[node.ID] = node
		graph.Files[fileID].Functions = append(graph.Files[fileID].Functions, node.ID)
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
	return project, nil
}

// writeHelper writes the embedded analyzer source to a temporary file.
func writeHelper() (string, error) {
	file, err := os.CreateTemp("", "senbon-analyze-*.ml")
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

// ocamlCompileInclude returns the compiler-libs include path.
func ocamlCompileInclude() string {
	return ocamlStdlib() + "/compiler-libs"
}

var ocamlStdlibPath = ""

func ocamlStdlib() string {
	if ocamlStdlibPath != "" {
		return ocamlStdlibPath
	}
	if output, err := exec.Command("ocamlc", "-where").Output(); err == nil {
		ocamlStdlibPath = string(output[:len(output)-1])
	}
	return ocamlStdlibPath
}

func ocamlCommon() string {
	return ocamlStdlib() + "/compiler-libs/ocamlcommon.cma"
}
