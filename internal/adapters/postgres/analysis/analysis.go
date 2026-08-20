// Package analysis builds a synthetic graph of a running PostgreSQL database.
// Statements and tables share one attribution graph; the TUI presents them as
// separate views.
package analysis

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"strconv"
	"strings"

	"github.com/briheet/sen/internal/model"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

const (
	// ModulePath is the synthetic package path returned as the graph namespace.
	ModulePath = "sen/postgres"
	moduleName = "postgres"
	maxLabel   = 60
)

// Statement identifies one normalized query from pg_stat_statements.
type Statement struct {
	QueryID int64
	Query   string
}

// Table identities one relation from pg_stat_user_tables.
type Table struct {
	Name string
}

// Analyze enumerates the current database's statements and user tables. The
// connection is closed before returning; the runtime opens its own.
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

// queryStatements tolerates pg_stat_statements being absent so table
// topology and database metrics remain usable without the extension.
func queryStatements(ctx context.Context, conn *pgx.Conn) ([]Statement, error) {
	rows, err := conn.Query(ctx, `
		SELECT DISTINCT queryid, query
		FROM pg_stat_statements
		WHERE dbid = (SELECT oid FROM pg_database WHERE datname = current_database())
		  AND queryid IS NOT NULL AND query IS NOT NULL`)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "42P01" {
			return nil, nil
		}
		return nil, err
	}
	defer rows.Close()

	var out []Statement
	for rows.Next() {
		var statement Statement
		if err := rows.Scan(&statement.QueryID, &statement.Query); err != nil {
			return nil, err
		}
		out = append(out, statement)
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

// BuildGraph constructs the attribution graph from the given identities.
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
	sort.Slice(statements, func(i, j int) bool { return statements[i].QueryID < statements[j].QueryID })
	for _, statement := range statements {
		addNode(graph, rootPkg, rootID, StatementLabel(statement.Query), StmtPath(statement.QueryID))
	}
	sort.Slice(tables, func(i, j int) bool { return tables[i].Name < tables[j].Name })
	for _, table := range tables {
		addNode(graph, rootPkg, rootID, table.Name, TablePath(table.Name))
	}
	return graph
}

// StatementLabel turns SQL into a compact, single-line graph label.
func StatementLabel(query string) string {
	query = strings.Join(strings.Fields(query), " ")
	runes := []rune(query)
	if len(runes) <= maxLabel {
		return query
	}
	return string(runes[:maxLabel]) + "…"
}

// StmtPath returns the synthetic file path for a statement query id.
func StmtPath(queryID int64) string {
	return ModulePath + "/stmt/" + strconv.FormatInt(queryID, 10)
}

// addNode adds one database entity as a synthetic child of the server root.
func addNode(graph *model.StaticGraph, pkg model.PackageID, root model.NodeID, name, path string) {
	id := model.NodeID(len(graph.Nodes) + 1)
	fileID := model.FileID(len(graph.Files) + 1)
	graph.Nodes[id] = &model.StaticNode{
		Name: name, ID: id, Pkg: pkg,
		Syntax: model.Syntax{Kind: model.SyntaxFuncDecl, File: fileID, Start: model.Position{Line: 1}, End: model.Position{Line: 1}},
		In:     []model.NodeID{root},
	}
	graph.Nodes[root].Out = append(graph.Nodes[root].Out, id)
	graph.Files[fileID] = &model.StaticFile{
		ID: fileID, Path: path, Package: pkg, Functions: []model.NodeID{id},
	}
}

// TablePath returns the synthetic file path for a table name.
func TablePath(name string) string {
	return ModulePath + "/table/" + name
}
