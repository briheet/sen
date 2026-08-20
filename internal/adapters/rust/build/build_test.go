package build

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestInjectMainIgnoresCommentsAndStrings(t *testing.T) {
	t.Parallel()
	source := []byte("// fn main() {}\nconst TEXT: &str = \"fn main() {}\";\n#[tokio::main]\nasync fn main() {\n    println!(\"ready\");\n}")

	result, err := injectMain(source)
	require.NoError(t, err)
	require.Contains(t, string(result), "async fn main() {\n    console_subscriber::init();")
	require.Equal(t, 1, strings.Count(string(result), "console_subscriber::init();"))
}

func TestInjectMainRequiresSourceEntrypoint(t *testing.T) {
	t.Parallel()
	_, err := injectMain([]byte("pub fn run() {}"))
	require.ErrorContains(t, err, "fn main")
}
