// Package analysis builds a synthetic static graph of the running PostgreSQL
// server's current workload. Unlike Go/Node there is no source tree to parse;
// and unlike fixed-command stores, PostgreSQL's objects are open-ended. So we
// enumerate the statements and tables currently present in the server's own
// statistics views and turn each into a node, giving the TUI a stable surface
// onto which per-query and per-table heat is attributed.
package analysis

import (
	"context"
	"fmt"
	"sort"
	"strconv"

	"github.com/briheet/sen/internal/model"
	"github.com/jackc/pgx/v5"
)

const (
	// ModulePath is the synthetic package path returned as the graph namespace.
	ModulePath = "sen/postgres"
	moduleName = "postgres"
)

// Statement identities one normalized query from pg_stat_statements.
type Statement struct {
	QueryID int64
	Query   string
}

// Table identities one relation from pg_stat_user_tables.
type Table struct {
	Name string
}

// Analyze connects to the given PostgreSQL server, enumerates its current
// statements and user tables, and returns a graph with one node per identity.
// The connection is closed before returning; the runtime opens its own.
func Analyze(ctx context.Context, dsn string) (*model.StaticGraph, string, error) {
	conn, err := pgx.Connect(ctx, dsn)
	if err != nil {
		return nil, "", fmt.Errorf("postgres: connect: %w", err)
	}
	defer func() { _ = conn.Close(ctx) }()

	statements, err := queryStatements(ctx, conn)
	if err != nil {
		return nil, "", err
	}
	tables, err := queryTables(ctx, conn)
	if err != nil {
		return nil, "", err
	}
	return BuildGraph(statements, tables), ModulePath, nil
}

// queryStatements reads DISTINCT statement identities. It tolerates the
// pg_stat_statements view being absent (extension not installed): on failure
// it returns an empty set so the graph still shows tables and server metrics.
func queryStatements(ctx context.Context, conn *pgx.Conn) ([]Statement, error) {
	rows, err := conn.Query(ctx, `
		SELECT DISTINCT queryid, query
		FROM pg_stat_statements
		WHERE queryid IS NOT NULL AND query IS NOT NULL`)
	if err != nil {
		return nil, nil
	}
	defer rows.Close()

	var out []Statement
	for rows.Next() {
		var s Statement
		if err := rows.Scan(&s.QueryID, &s.Query); err != nil {
			return nil, err
		}
		out = append(out, s)
	}
	return out, rows.Err()
}

func queryTables(ctx context.Context, conn *pgx.Conn) ([]Table, error) {
	rows, err := conn.Query(ctx, `
		SELECT relname FROM pg_stat_user_tables ORDER BY relname`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []Table
	for rows.Next() {
		var t Table
		if err := rows.Scan(&t.Name); err != nil {
			return nil, err
		}
		out = append(out, t)
	}
	return out, rows.Err()
}

// BuildGraph constructs the synthetic graph from the given identities.
func BuildGraph(statements []Statement, tables []Table) *model.StaticGraph {
	graph := &model.StaticGraph{
		Nodes:    make(map[model.NodeID]*model.StaticNode),
		Files:    make(map[model.FileID]*model.StaticFile),
		Packages: make(map[model.PackageID]*model.Package),
	}
	rootPkg := model.PackageID(1)
	graph.Packages[rootPkg] = &model.Package{Path: ModulePath, Name: moduleName}

	rootID := model.NodeID(1)
	rootFileID := model.FileID(1)
	graph.Root = rootID
	graph.Nodes[rootID] = &model.StaticNode{
		Name:   "postgres-server",
		ID:     rootID,
		Pkg:    rootPkg,
		Syntax: model.Syntax{Kind: model.SyntaxFuncDecl, File: rootFileID, Start: model.Position{Line: 1}, End: model.Position{Line: 1}},
	}
	graph.Files[rootFileID] = &model.StaticFile{
		ID: rootFileID, Path: ModulePath + "/postgres-server", Package: rootPkg,
		Functions: []model.NodeID{rootID},
	}
	addNodes(graph, rootPkg, rootID, "stmt", statementNames(statements))
	addNodes(graph, rootPkg, rootID, "table", tableNames(tables))
	return graph
}

// addNodes adds one node+file per identity, each listed as a child of the root.
func addNodes(graph *model.StaticGraph, pkg model.PackageID, root model.NodeID, kind string, names []string) {
	for _, name := range names {
		id := model.NodeID(len(graph.Nodes) + 1)
		fileID := model.FileID(len(graph.Files) + 1)
		path := ModulePath + "/" + kind + "/" + name
		graph.Nodes[id] = &model.StaticNode{
			Name:   name,
			ID:     id,
			Pkg:    pkg,
			Syntax: model.Syntax{Kind: model.SyntaxFuncDecl, File: fileID, Start: model.Position{Line: 1}, End: model.Position{Line: 1}},
			In:     []model.NodeID{root},
		}
		graph.Nodes[root].Out = append(graph.Nodes[root].Out, id)
		graph.Files[fileID] = &model.StaticFile{
			ID: fileID, Path: path, Package: pkg, Functions: []model.NodeID{id},
		}
	}
}

func statementNames(statements []Statement) []string {
	names := make([]string, 0, len(statements))
	for _, s := range statements {
		names = append(names, strconv.FormatInt(s.QueryID, 10))
	}
	sort.Strings(names)
	return names
}

func tableNames(tables []Table) []string {
	names := make([]string, 0, len(tables))
	for _, t := range tables {
		names = append(names, t.Name)
	}
	sort.Strings(names)
	return names
}

// StmtPath returns the synthetic file path for a statement query id.
func StmtPath(queryID int64) string {
	return ModulePath + "/stmt/" + strconv.FormatInt(queryID, 10)
}

// TablePath returns the synthetic file path for a table name.
func TablePath(name string) string {
	return ModulePath + "/table/" + name
}
