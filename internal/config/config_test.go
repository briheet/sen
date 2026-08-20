package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestLoad(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	path := filepath.Join(dir, "sen.toml")
	data := []byte(`[project]
name = "my-backend"

[[services]]
name = "api"
type = "server"
lang = "go"
path = "./cmd/server"
build_args = ["-tags", "production"]
run_args = ["--port", "8080"]

[[services]]
name = "worker"
type = "server"
lang = "node"
path = "./cmd/worker"
run_args = ["--queue", "events"]

[[services]]
name = "cache"
type = "kv"
provider = "redis"
address = "localhost:6379"

[[services]]
name = "database"
type = "db"
provider = "postgres"
address = " postgres://sen:sen@localhost:5432/sen "
`)
	require.NoError(t, os.WriteFile(path, data, 0o600))

	result, err := Load(path)
	require.NoError(t, err)
	require.Equal(t, "my-backend", result.Project.Name)
	require.Equal(t, []Service{
		{
			Name:      "api",
			Type:      ServiceTypeServer,
			Lang:      ServiceLangGo,
			Path:      filepath.Join(dir, "cmd", "server"),
			BuildArgs: []string{"-tags", "production"},
			RunArgs:   []string{"--port", "8080"},
		},
		{
			Name:    "worker",
			Type:    ServiceTypeServer,
			Lang:    ServiceLangNode,
			Path:    filepath.Join(dir, "cmd", "worker"),
			RunArgs: []string{"--queue", "events"},
		},
		{Name: "cache", Type: ServiceTypeKV, Provider: ServiceProviderRedis, Address: "localhost:6379"},
		{Name: "database", Type: ServiceTypeDB, Provider: ServiceProviderPostgres, Address: "postgres://sen:sen@localhost:5432/sen"},
	}, result.Services)
}

func TestLoadErrors(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		content string
		want    string
	}{
		{name: "malformed TOML", content: `[project`, want: "read config"},
		{name: "unknown field", content: validConfig(`extra = "value"`), want: "invalid keys"},
		{name: "missing project name", content: "[project]\n\n[[services]]\nname = \"api\"\ntype = \"server\"\nlang = \"go\"\npath = \".\"", want: "Project.Name"},
		{name: "project name is a path", content: "[project]\nname = \"../project\"\n\n[[services]]\nname = \"api\"\ntype = \"server\"\nlang = \"go\"\npath = \".\"", want: "project name must not be a path"},
		{name: "missing services", content: "[project]\nname = \"project\"", want: "Config.Services"},
		{name: "missing service name", content: validConfig("type = \"server\"\nlang = \"go\"\npath = \".\""), want: "Services[0].Name"},
		{name: "duplicate service name", content: validConfig("name = \"api\"\ntype = \"server\"\nlang = \"go\"\npath = \".\"\n\n[[services]]\nname = \"api\"\ntype = \"server\"\nlang = \"node\"\npath = \".\""), want: "duplicate name"},
		{name: "unsupported type", content: validConfig("name = \"api\"\ntype = \"python\"\npath = \".\""), want: "Services[0].Type"},
		{name: "unsupported lang", content: validConfig("name = \"api\"\ntype = \"server\"\nlang = \"python\"\npath = \".\""), want: "Services[0].Lang"},
		{name: "unsupported provider", content: validConfig("name = \"cache\"\ntype = \"kv\"\nprovider = \"memcached\"\naddress = \"localhost:11211\""), want: "Services[0].Provider"},
		{name: "server missing lang", content: validConfig("name = \"api\"\ntype = \"server\"\npath = \".\""), want: "requires lang"},
		{name: "server missing path", content: validConfig("name = \"api\"\ntype = \"server\"\nlang = \"go\""), want: "requires path"},
		{name: "server with provider", content: validConfig("name = \"api\"\ntype = \"server\"\nlang = \"go\"\nprovider = \"redis\"\npath = \".\""), want: "cannot define provider"},
		{name: "server with address", content: validConfig("name = \"api\"\ntype = \"server\"\nlang = \"node\"\npath = \".\"\naddress = \"localhost:3000\""), want: "cannot define address"},
		{name: "kv missing provider", content: validConfig("name = \"cache\"\ntype = \"kv\"\naddress = \"localhost:6379\""), want: "requires provider"},
		{name: "kv missing address", content: validConfig("name = \"cache\"\ntype = \"kv\"\nprovider = \"redis\""), want: "requires address"},
		{name: "kv with lang", content: validConfig("name = \"cache\"\ntype = \"kv\"\nprovider = \"redis\"\nlang = \"go\"\naddress = \"localhost:6379\""), want: "cannot define lang"},
		{name: "kv with path", content: validConfig("name = \"cache\"\ntype = \"kv\"\nprovider = \"redis\"\naddress = \"localhost:6379\"\npath = \".\""), want: "cannot define path"},
		{name: "kv with arguments", content: validConfig("name = \"cache\"\ntype = \"kv\"\nprovider = \"redis\"\naddress = \"localhost:6379\"\nrun_args = [\"--flag\"]"), want: "cannot define build_args or run_args"},
		{name: "kv with database provider", content: validConfig("name = \"cache\"\ntype = \"kv\"\nprovider = \"postgres\"\naddress = \"localhost:5432\""), want: "unsupported provider"},
		{name: "db missing provider", content: validConfig("name = \"database\"\ntype = \"db\"\naddress = \"localhost:5432\""), want: "requires provider"},
		{name: "db missing address", content: validConfig("name = \"database\"\ntype = \"db\"\nprovider = \"postgres\""), want: "requires address"},
		{name: "db with kv provider", content: validConfig("name = \"database\"\ntype = \"db\"\nprovider = \"redis\"\naddress = \"localhost:5432\""), want: "unsupported provider"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			path := filepath.Join(t.TempDir(), "sen.toml")
			require.NoError(t, os.WriteFile(path, []byte(test.content), 0o600))

			_, err := Load(path)
			require.ErrorContains(t, err, test.want)
		})
	}
}

func TestLoadMissingFile(t *testing.T) {
	t.Parallel()

	_, err := Load(filepath.Join(t.TempDir(), "sen.toml"))
	require.ErrorContains(t, err, "read config")
}

func TestResolvePath(t *testing.T) {
	dir := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(dir, DefaultPath), nil, 0o600))

	path, err := ResolvePath(DefaultPath, false, []string{dir})
	require.NoError(t, err)
	require.Equal(t, filepath.Join(dir, DefaultPath), path)

	path, err = ResolvePath(dir, true, nil)
	require.NoError(t, err)
	require.Equal(t, filepath.Join(dir, DefaultPath), path)

	_, err = ResolvePath("config.toml", true, []string{dir})
	require.ErrorIs(t, err, errConfigPathConflict)
}

func validConfig(service string) string {
	return `[project]
name = "project"

[[services]]
` + service
}
