// Package analysis builds TigerBeetle's synthetic attribution topology.
package analysis

import (
	"fmt"
	"strings"

	"github.com/briheet/sen/internal/model"
)

const (
	// ModulePath is the namespace used by TigerBeetle profiles.
	ModulePath = "sen/tigerbeetle"
	moduleName = "tigerbeetle"
)

// Operations are the stable client request operations exposed by TigerBeetle.
var Operations = []string{
	"create_accounts",
	"create_transfers",
	"lookup_accounts",
	"lookup_transfers",
	"get_account_transfers",
	"get_account_balances",
	"query_accounts",
	"query_transfers",
}

var operationSet = func() map[string]struct{} {
	result := make(map[string]struct{}, len(Operations))
	for _, operation := range Operations {
		result[operation] = struct{}{}
	}
	return result
}()

// NormalizeOperation accepts TigerBeetle's enum tag form and rejects internal
// operations, keeping metric cardinality bounded to the public API.
func NormalizeOperation(operation string) (string, bool) {
	operation = strings.ToLower(strings.TrimPrefix(strings.TrimSpace(operation), "Operation."))
	_, ok := operationSet[operation]
	return operation, ok
}

// BuildGraph creates one cluster root with operation and configured replica
// children. Replica order is significant and matches --addresses.
func BuildGraph(addresses []string) *model.StaticGraph {
	graph := &model.StaticGraph{
		Nodes: make(map[model.NodeID]*model.StaticNode), Files: make(map[model.FileID]*model.StaticFile),
		Packages: make(map[model.PackageID]*model.Package),
	}
	const pkg = model.PackageID(1)
	graph.Packages[pkg] = &model.Package{Path: ModulePath, Name: moduleName}
	addRoot(graph, pkg)
	for _, operation := range Operations {
		addNode(graph, pkg, operation, OperationPath(operation))
	}
	for index := range addresses {
		addNode(graph, pkg, fmt.Sprintf("replica %d", index), ReplicaPath(uint32(index)))
	}
	return graph
}

func addRoot(graph *model.StaticGraph, pkg model.PackageID) {
	const id, file = model.NodeID(1), model.FileID(1)
	graph.Root = id
	graph.Nodes[id] = &model.StaticNode{
		Name: "tigerbeetle-cluster", ID: id, Pkg: pkg,
		Syntax: model.Syntax{Kind: model.SyntaxFuncDecl, File: file, Start: model.Position{Line: 1}, End: model.Position{Line: 1}},
	}
	graph.Files[file] = &model.StaticFile{ID: file, Path: ModulePath + "/cluster", Package: pkg, Functions: []model.NodeID{id}}
}

func addNode(graph *model.StaticGraph, pkg model.PackageID, name, path string) {
	id := model.NodeID(len(graph.Nodes) + 1)
	file := model.FileID(len(graph.Files) + 1)
	graph.Nodes[id] = &model.StaticNode{
		Name: name, ID: id, Pkg: pkg,
		Syntax: model.Syntax{Kind: model.SyntaxFuncDecl, File: file, Start: model.Position{Line: 1}, End: model.Position{Line: 1}},
		In:     []model.NodeID{graph.Root},
	}
	graph.Nodes[graph.Root].Out = append(graph.Nodes[graph.Root].Out, id)
	graph.Files[file] = &model.StaticFile{ID: file, Path: path, Package: pkg, Functions: []model.NodeID{id}}
}

// OperationPath returns the synthetic source path for an operation.
func OperationPath(operation string) string { return ModulePath + "/operation/" + operation }

// ReplicaPath returns the synthetic source path for a replica index.
func ReplicaPath(replica uint32) string { return fmt.Sprintf("%s/replica/%d", ModulePath, replica) }
