// Package build prepares Cargo applications for analysis and execution.
package build

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"slices"
	"strings"

	"github.com/briheet/sen/internal/adapters"
	"github.com/briheet/sen/internal/config"
)

const (
	tempPattern      = "sen-rust-*"
	WrapperEnv       = "SEN_RUSTC_WRAPPER"
	wrapperSourceEnv = "SEN_RUST_SOURCE"
	wrapperShadowEnv = "SEN_RUST_SHADOW"
	wrapperNextEnv   = "SEN_RUST_NEXT_WRAPPER"
	consolePackage   = "console-subscriber"
	consoleVersion   = "0.5"
)

// Prepared is one isolated Cargo build owned by a Rust runtime.
type Prepared struct {
	Binary      string
	SourceRoot  string
	Workspace   string
	Package     string
	Target      string
	TempDir     string
	ConsoleMode config.TokioConsoleMode
}

// Cleanup removes this build's isolated Cargo output and source shadow.
func (p *Prepared) Cleanup() error {
	if p == nil || p.TempDir == "" {
		return nil
	}
	return os.RemoveAll(p.TempDir)
}

type metadata struct {
	WorkspaceRoot string            `json:"workspace_root"`
	Packages      []metadataPackage `json:"packages"`
	Resolve       *metadataResolve  `json:"resolve"`
}

type metadataPackage struct {
	ID           string               `json:"id"`
	Name         string               `json:"name"`
	ManifestPath string               `json:"manifest_path"`
	Targets      []metadataTarget     `json:"targets"`
	Dependencies []metadataDependency `json:"dependencies"`
}

type metadataTarget struct {
	Name    string   `json:"name"`
	Kind    []string `json:"kind"`
	SrcPath string   `json:"src_path"`
}

type metadataDependency struct {
	Name     string   `json:"name"`
	Rename   string   `json:"rename"`
	Req      string   `json:"req"`
	Features []string `json:"features"`
}

type metadataResolve struct {
	Nodes []metadataNode `json:"nodes"`
}

type metadataNode struct {
	ID           string   `json:"id"`
	Dependencies []string `json:"dependencies"`
	Features     []string `json:"features"`
}

type selection struct {
	pkg    metadataPackage
	target metadataTarget
}

type cargoMessage struct {
	Reason     string         `json:"reason"`
	PackageID  string         `json:"package_id"`
	Target     metadataTarget `json:"target"`
	Executable string         `json:"executable"`
	Message    struct {
		Rendered string `json:"rendered"`
	} `json:"message"`
}

// Prepare discovers, instruments, and builds one executable Cargo target.
func Prepare(ctx context.Context, sourcePath string, buildArgs []string, mode config.TokioConsoleMode, output adapters.Output) (_ *Prepared, err error) {
	sourcePath, err = filepath.Abs(sourcePath)
	if err != nil {
		return nil, err
	}
	sourcePath, err = filepath.EvalSymlinks(sourcePath)
	if err != nil {
		return nil, err
	}

	meta, err := loadMetadata(ctx, sourcePath, output.Stderr)
	if err != nil {
		return nil, err
	}
	selected, err := selectTarget(meta, sourcePath, buildArgs)
	if err != nil {
		return nil, err
	}
	if mode == config.TokioConsoleInject {
		if err := validateConsoleDependency(meta, selected.pkg); err != nil {
			return nil, err
		}
	}

	tempDir, err := os.MkdirTemp("", tempPattern)
	if err != nil {
		return nil, err
	}
	defer func() {
		if err != nil {
			_ = os.RemoveAll(tempDir)
		}
	}()

	env := append([]string(nil), os.Environ()...)
	env = appendRustFlags(env, mode)
	env = setEnv(env, "CARGO_TARGET_DIR", filepath.Join(tempDir, "target"))
	if mode == config.TokioConsoleInject {
		shadow, shadowErr := createShadow(selected.pkg, selected.target, tempDir)
		if shadowErr != nil {
			return nil, shadowErr
		}
		executable, executableErr := os.Executable()
		if executableErr != nil {
			return nil, executableErr
		}
		env = appendEncodedRustFlag(env, "--remap-path-prefix="+filepath.Join(tempDir, "shadow")+"="+filepath.Dir(selected.pkg.ManifestPath))
		env = setEnv(env, wrapperNextEnv, os.Getenv("RUSTC_WORKSPACE_WRAPPER"))
		env = setEnv(env, "RUSTC_WORKSPACE_WRAPPER", executable)
		env = setEnv(env, WrapperEnv, "1")
		env = setEnv(env, wrapperSourceEnv, filepath.Clean(selected.target.SrcPath))
		env = setEnv(env, wrapperShadowEnv, shadow)
	}

	binary, err := cargoBuild(ctx, sourcePath, buildArgs, selected, env, output)
	if err != nil {
		return nil, err
	}
	return &Prepared{
		Binary:      binary,
		SourceRoot:  filepath.Dir(selected.pkg.ManifestPath),
		Workspace:   meta.WorkspaceRoot,
		Package:     selected.pkg.Name,
		Target:      selected.target.Name,
		TempDir:     tempDir,
		ConsoleMode: mode,
	}, nil
}

func appendEncodedRustFlag(env []string, flag string) []string {
	flags := envValue(env, "CARGO_ENCODED_RUSTFLAGS")
	if flags != "" {
		flags += "\x1f"
	}
	return setEnv(env, "CARGO_ENCODED_RUSTFLAGS", flags+flag)
}

func loadMetadata(ctx context.Context, dir string, stderr io.Writer) (metadata, error) {
	var captured bytes.Buffer
	if stderr == nil {
		stderr = &captured
	}
	cmd := exec.CommandContext(ctx, "cargo", "metadata", "--format-version=1")
	cmd.Dir = dir
	cmd.Stderr = stderr
	data, err := cmd.Output()
	if err != nil {
		return metadata{}, commandError("cargo metadata", err, captured.String())
	}
	var result metadata
	if err := json.Unmarshal(data, &result); err != nil {
		return metadata{}, fmt.Errorf("decode cargo metadata: %w", err)
	}
	return result, nil
}

func selectTarget(meta metadata, sourcePath string, buildArgs []string) (selection, error) {
	requested := requestedBinary(buildArgs)
	var candidates []selection
	for _, pkg := range meta.Packages {
		manifestDir := filepath.Dir(pkg.ManifestPath)
		if !within(sourcePath, manifestDir) && !within(manifestDir, sourcePath) {
			continue
		}
		for _, target := range pkg.Targets {
			if !slices.Contains(target.Kind, "bin") || requested != "" && target.Name != requested {
				continue
			}
			candidates = append(candidates, selection{pkg: pkg, target: target})
		}
	}
	if len(candidates) == 0 {
		if requested != "" {
			return selection{}, fmt.Errorf("cargo binary %q was not found below %s", requested, sourcePath)
		}
		return selection{}, errors.New("no Cargo binary target found")
	}
	if len(candidates) != 1 {
		names := make([]string, 0, len(candidates))
		for _, candidate := range candidates {
			names = append(names, candidate.target.Name)
		}
		slices.Sort(names)
		return selection{}, fmt.Errorf("multiple Cargo binaries found (%s); select one with build_args = [\"--bin\", \"name\"]", strings.Join(names, ", "))
	}
	return candidates[0], nil
}

func requestedBinary(args []string) string {
	for index, arg := range args {
		if arg == "--bin" && index+1 < len(args) {
			return args[index+1]
		}
		if value, ok := strings.CutPrefix(arg, "--bin="); ok {
			return value
		}
	}
	return ""
}

func validateConsoleDependency(meta metadata, pkg metadataPackage) error {
	found := false
	for _, dependency := range pkg.Dependencies {
		if dependency.Name == consolePackage {
			found = true
			if !strings.Contains(dependency.Req, consoleVersion) && dependency.Req != "*" {
				return fmt.Errorf("tokio_console=inject requires console-subscriber 0.5.x; Cargo.toml declares %s", dependency.Req)
			}
		}
	}
	if !found {
		return errors.New("tokio_console=inject requires the official dependency console-subscriber = \"0.5\"")
	}
	packages := make(map[string]metadataPackage, len(meta.Packages))
	nodes := make(map[string]metadataNode, len(meta.Resolve.Nodes))
	for _, candidate := range meta.Packages {
		packages[candidate.ID] = candidate
	}
	for _, node := range meta.Resolve.Nodes {
		nodes[node.ID] = node
	}
	for _, dependency := range nodes[pkg.ID].Dependencies {
		if packages[dependency].Name == "tokio" && slices.Contains(nodes[dependency].Features, "tracing") {
			return nil
		}
	}
	return errors.New("tokio_console=inject requires Tokio's official tracing feature")
}

func cargoBuild(ctx context.Context, dir string, buildArgs []string, selected selection, env []string, output adapters.Output) (string, error) {
	var capturedStderr, diagnostics bytes.Buffer
	stderr := output.Stderr
	if stderr == nil {
		stderr = &capturedStderr
	}
	args := []string{"build"}
	args = append(args, buildArgs...)
	args = append(args, "--message-format=json-render-diagnostics")
	cmd := exec.CommandContext(ctx, "cargo", args...)
	cmd.Dir = dir
	cmd.Env = env
	cmd.Stdin = os.Stdin
	cmd.Stderr = stderr
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return "", err
	}
	if err := cmd.Start(); err != nil {
		return "", err
	}
	var binary string
	scanner := bufio.NewScanner(stdout)
	scanner.Buffer(make([]byte, 64*1024), 4*1024*1024)
	for scanner.Scan() {
		line := scanner.Bytes()
		var message cargoMessage
		if json.Unmarshal(line, &message) == nil && message.Reason == "compiler-artifact" && message.PackageID == selected.pkg.ID && message.Target.Name == selected.target.Name && message.Executable != "" {
			binary = message.Executable
			continue
		}
		if message.Message.Rendered != "" {
			diagnostics.WriteString(message.Message.Rendered)
		}
		if output.Stdout != nil {
			_, _ = output.Stdout.Write(append(append([]byte(nil), line...), '\n'))
		}
	}
	err = errors.Join(scanner.Err(), cmd.Wait())
	if err != nil {
		return "", commandError("cargo build", err, diagnostics.String()+capturedStderr.String())
	}
	if binary == "" {
		return "", errors.New("Cargo did not report the selected executable artifact")
	}
	return binary, nil
}

func commandError(operation string, err error, details string) error {
	details = strings.TrimSpace(details)
	if details == "" {
		return fmt.Errorf("%s: %w", operation, err)
	}
	return fmt.Errorf("%s: %w\n%s", operation, err, details)
}

func createShadow(pkg metadataPackage, target metadataTarget, tempDir string) (string, error) {
	root := filepath.Dir(pkg.ManifestPath)
	shadowRoot := filepath.Join(tempDir, "shadow")
	if err := mirror(root, shadowRoot); err != nil {
		return "", err
	}
	relative, err := filepath.Rel(root, target.SrcPath)
	if err != nil || strings.HasPrefix(relative, "..") {
		return "", errors.New("Cargo binary entrypoint is outside its package")
	}
	shadow := filepath.Join(shadowRoot, relative)
	data, err := os.ReadFile(target.SrcPath)
	if err != nil {
		return "", err
	}
	injected, err := injectMain(data)
	if err != nil {
		return "", err
	}
	if err := os.Remove(shadow); err != nil {
		return "", err
	}
	if err := os.WriteFile(shadow, injected, 0o600); err != nil {
		return "", err
	}
	return shadow, nil
}

func mirror(root, target string) error {
	return filepath.WalkDir(root, func(path string, entry os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		relative, err := filepath.Rel(root, path)
		if err != nil {
			return err
		}
		if relative == ".git" || relative == "target" || strings.HasPrefix(relative, ".git"+string(filepath.Separator)) || strings.HasPrefix(relative, "target"+string(filepath.Separator)) {
			if entry.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}
		destination := filepath.Join(target, relative)
		if entry.IsDir() {
			return os.MkdirAll(destination, 0o700)
		}
		return os.Symlink(path, destination)
	})
}

func injectMain(source []byte) ([]byte, error) {
	brace := mainBrace(source)
	if brace < 0 {
		return nil, errors.New("tokio_console=inject requires a source-defined fn main entrypoint")
	}
	const statement = "\n    console_subscriber::init();"
	result := make([]byte, 0, len(source)+len(statement))
	result = append(result, source[:brace+1]...)
	result = append(result, statement...)
	result = append(result, source[brace+1:]...)
	return result, nil
}

func mainBrace(source []byte) int {
	tokens := rustTokens(source)
	for index := 0; index+1 < len(tokens); index++ {
		if string(source[tokens[index][0]:tokens[index][1]]) != "fn" || string(source[tokens[index+1][0]:tokens[index+1][1]]) != "main" {
			continue
		}
		for position := tokens[index+1][1]; position < len(source); position++ {
			if source[position] == '{' {
				return position
			}
		}
	}
	return -1
}

// rustTokens returns identifier bounds while ignoring comments and string literals.
func rustTokens(source []byte) [][2]int {
	var tokens [][2]int
	for index := 0; index < len(source); {
		switch {
		case index+1 < len(source) && source[index] == '/' && source[index+1] == '/':
			index += 2
			for index < len(source) && source[index] != '\n' {
				index++
			}
		case index+1 < len(source) && source[index] == '/' && source[index+1] == '*':
			index += 2
			depth := 1
			for index < len(source) && depth > 0 {
				if index+1 < len(source) && source[index] == '/' && source[index+1] == '*' {
					depth++
					index += 2
					continue
				}
				if index+1 < len(source) && source[index] == '*' && source[index+1] == '/' {
					depth--
					index += 2
					continue
				}
				index++
			}
		case source[index] == '"' || source[index] == '\'':
			quote := source[index]
			index++
			for index < len(source) {
				if source[index] == '\\' {
					index += min(2, len(source)-index)
					continue
				}
				value := source[index]
				index++
				if value == quote {
					break
				}
			}
		case isIdentifierStart(source[index]):
			start := index
			index++
			for index < len(source) && isIdentifierContinue(source[index]) {
				index++
			}
			tokens = append(tokens, [2]int{start, index})
		default:
			index++
		}
	}
	return tokens
}

func isIdentifierStart(value byte) bool {
	return value == '_' || value >= 'a' && value <= 'z' || value >= 'A' && value <= 'Z'
}
func isIdentifierContinue(value byte) bool {
	return isIdentifierStart(value) || value >= '0' && value <= '9'
}

func appendRustFlags(env []string, mode config.TokioConsoleMode) []string {
	flags := []string{"-Cforce-frame-pointers=yes", "-Cdebuginfo=1", "-Cstrip=none"}
	if runtime.GOOS == "darwin" {
		flags = append(flags, "-Csplit-debuginfo=packed")
	} else {
		flags = append(flags, "-Csplit-debuginfo=off")
	}
	if mode != config.TokioConsoleOff {
		flags = append(flags, "--cfg", "tokio_unstable")
	}
	// rustup toolchains inside a Nix shell do not use Nix's compiler wrapper,
	// so forward its library search flags to the native linker explicitly.
	for _, flag := range strings.Fields(envValue(env, "NIX_LDFLAGS")) {
		flags = append(flags, "-Clink-arg="+flag)
	}
	encoded := envValue(env, "CARGO_ENCODED_RUSTFLAGS")
	if encoded != "" {
		flags = append(strings.Split(encoded, "\x1f"), flags...)
	} else if plain := envValue(env, "RUSTFLAGS"); plain != "" {
		flags = append(strings.Fields(plain), flags...)
	}
	return setEnv(env, "CARGO_ENCODED_RUSTFLAGS", strings.Join(flags, "\x1f"))
}

func setEnv(env []string, key, value string) []string {
	prefix := key + "="
	for index, item := range env {
		if strings.HasPrefix(item, prefix) {
			env[index] = prefix + value
			return env
		}
	}
	return append(env, prefix+value)
}

func envValue(env []string, key string) string {
	prefix := key + "="
	for index := len(env) - 1; index >= 0; index-- {
		if value, ok := strings.CutPrefix(env[index], prefix); ok {
			return value
		}
	}
	return ""
}

func within(path, root string) bool {
	relative, err := filepath.Rel(root, path)
	return err == nil && relative != ".." && !strings.HasPrefix(relative, ".."+string(filepath.Separator))
}

// RunWrapper replaces the selected crate root and delegates to rustc or the user's wrapper.
func RunWrapper(args []string) int {
	if len(args) == 0 {
		_, _ = fmt.Fprintln(os.Stderr, "sen rustc wrapper: missing rustc executable")
		return 2
	}
	executable := args[0]
	compilerArgs := append([]string(nil), args[1:]...)
	from, to := filepath.Clean(os.Getenv(wrapperSourceEnv)), os.Getenv(wrapperShadowEnv)
	for index, arg := range compilerArgs {
		candidate := filepath.Clean(arg)
		if !filepath.IsAbs(candidate) {
			if absolute, err := filepath.Abs(candidate); err == nil {
				candidate = absolute
			}
		}
		if candidate == from {
			compilerArgs[index] = to
		}
	}
	if next := os.Getenv(wrapperNextEnv); next != "" {
		compilerArgs = append([]string{executable}, compilerArgs...)
		executable = next
	}
	cmd := exec.Command(executable, compilerArgs...)
	cmd.Stdin, cmd.Stdout, cmd.Stderr = os.Stdin, os.Stdout, os.Stderr
	cmd.Env = os.Environ()
	if err := cmd.Run(); err != nil {
		var exit *exec.ExitError
		if errors.As(err, &exit) {
			return exit.ExitCode()
		}
		_, _ = fmt.Fprintln(os.Stderr, err)
		return 1
	}
	return 0
}
