package command

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

const (
	runCommandHelperEnv  = "REACTENV_RUN_HELPER"
	runCommandHelperMode = "REACTENV_RUN_MODE"
	runCommandHelperPath = "REACTENV_RUN_PATH"
)

const (
	runHelperModeMissingArgs  = "missing-args"
	runHelperModeInvalidMatch = "invalid-match"
	runHelperModeMissingEnv   = "missing-env"
	runHelperModeNoFiles      = "no-files"
	runHelperModePartialWrite = "partial-write"
)

func TestRunCommandRunSuccessDefaultMatcher(t *testing.T) {
	tempDir := t.TempDir()

	nestedDir := filepath.Join(tempDir, "nested")
	require.NoError(t, os.MkdirAll(nestedDir, 0755), "create nested dir")

	rootJS := filepath.Join(tempDir, "root.js")
	nestedJS := filepath.Join(nestedDir, "app.js")
	ignoredTxt := filepath.Join(nestedDir, "ignored.txt")
	ignoredCSS := filepath.Join(tempDir, "style.css")
	ignoredJS := filepath.Join(tempDir, "misc.js")

	writeFile(t, rootJS, `const api="__reactenv.API_URL";`)
	writeFile(t, nestedJS, `const name="__reactenv.NAME";`)
	writeFile(t, ignoredTxt, `__reactenv.IGNORED`)
	writeFile(t, ignoredCSS, "body { color: red; }")
	writeFile(t, ignoredJS, "no matches")

	t.Setenv("API_URL", "https://example.com")
	t.Setenv("NAME", "app-name")

	cmd := &RunCommand{FileMatchPattern: defaultFileMatchPattern}
	exitCode := cmd.Run([]string{tempDir})
	require.Equal(t, 0, exitCode)

	require.Equal(t, `const api="https://example.com";`, readFile(t, rootJS))
	require.Equal(t, `const name="app-name";`, readFile(t, nestedJS))
	require.Equal(t, "__reactenv.IGNORED", readFile(t, ignoredTxt))
	require.Equal(t, "body { color: red; }", readFile(t, ignoredCSS))
	require.Equal(t, "no matches", readFile(t, ignoredJS))
}

func TestRunCommandRunSuccessCustomMatcher(t *testing.T) {
	tempDir := t.TempDir()

	nestedDir := filepath.Join(tempDir, "nested")
	require.NoError(t, os.MkdirAll(nestedDir, 0755), "create nested dir")

	rootJS := filepath.Join(tempDir, "root.js")
	targetJS := filepath.Join(nestedDir, "only.js")
	skipJS := filepath.Join(nestedDir, "skip.js")

	writeFile(t, rootJS, `const root="__reactenv.ROOT";`)
	writeFile(t, targetJS, `const nested="__reactenv.NESTED";`)
	writeFile(t, skipJS, `const skip="__reactenv.SKIP";`)

	t.Setenv("ROOT", "root")
	t.Setenv("NESTED", "nested")
	t.Setenv("SKIP", "skip")

	cmd := &RunCommand{FileMatchPattern: "glob:nested/only.js"}
	exitCode := cmd.Run([]string{tempDir})
	require.Equal(t, 0, exitCode)

	require.Equal(t, `const nested="nested";`, readFile(t, targetJS))
	require.Equal(t, `const root="__reactenv.ROOT";`, readFile(t, rootJS))
	require.Equal(t, `const skip="__reactenv.SKIP";`, readFile(t, skipJS))
}

func TestRunCommandRunMissingEnvExits(t *testing.T) {
	if handleRunCommandHelper(t) {
		return
	}

	tempDir := t.TempDir()

	nestedDir := filepath.Join(tempDir, "nested")
	require.NoError(t, os.MkdirAll(nestedDir, 0755), "create nested dir")

	rootJS := filepath.Join(tempDir, "root.js")
	nestedJS := filepath.Join(nestedDir, "app.js")

	originalRoot := `const missing="__reactenv.MISSING";`
	originalNested := `const present="__reactenv.PRESENT";`

	writeFile(t, rootJS, originalRoot)
	writeFile(t, nestedJS, originalNested)

	err := runCommandHelperProcess("TestRunCommandRunMissingEnvExits", runHelperModeMissingEnv, tempDir, "PRESENT=present", "MISSING=")

	var exitErr *exec.ExitError
	require.ErrorAs(t, err, &exitErr)
	require.Equal(t, 1, exitErr.ExitCode())

	require.Equal(t, originalRoot, readFile(t, rootJS))
	require.Equal(t, originalNested, readFile(t, nestedJS))
}

func TestRunCommandRunMissingArgsExits(t *testing.T) {
	if handleRunCommandHelper(t) {
		return
	}

	err := runCommandHelperProcess("TestRunCommandRunMissingArgsExits", runHelperModeMissingArgs, "", "")

	var exitErr *exec.ExitError
	require.ErrorAs(t, err, &exitErr)
	require.Equal(t, 1, exitErr.ExitCode())
}

func TestRunCommandRunInvalidMatchExits(t *testing.T) {
	if handleRunCommandHelper(t) {
		return
	}

	tempDir := t.TempDir()
	writeFile(t, filepath.Join(tempDir, "root.js"), `const api="__reactenv.API_URL";`)

	err := runCommandHelperProcess("TestRunCommandRunInvalidMatchExits", runHelperModeInvalidMatch, tempDir, "API_URL=https://example.com")

	var exitErr *exec.ExitError
	require.ErrorAs(t, err, &exitErr)
	require.Equal(t, 1, exitErr.ExitCode())
}

func TestRunCommandRunNoFilesFoundExits(t *testing.T) {
	if handleRunCommandHelper(t) {
		return
	}

	tempDir := t.TempDir()
	writeFile(t, filepath.Join(tempDir, "style.css"), "body { color: red; }")

	err := runCommandHelperProcess("TestRunCommandRunNoFilesFoundExits", runHelperModeNoFiles, tempDir)

	var exitErr *exec.ExitError
	require.ErrorAs(t, err, &exitErr)
	require.Equal(t, 1, exitErr.ExitCode())
}

func TestRunCommandRunNoOccurrencesReturnsOne(t *testing.T) {
	tempDir := t.TempDir()

	nestedDir := filepath.Join(tempDir, "nested")
	require.NoError(t, os.MkdirAll(nestedDir, 0755), "create nested dir")

	rootJS := filepath.Join(tempDir, "root.js")
	nestedJS := filepath.Join(nestedDir, "app.js")

	writeFile(t, rootJS, "const api='no matches';")
	writeFile(t, nestedJS, "const name='still none';")

	cmd := &RunCommand{FileMatchPattern: defaultFileMatchPattern}
	exitCode := cmd.Run([]string{tempDir})
	require.Equal(t, 1, exitCode)

	require.Equal(t, "const api='no matches';", readFile(t, rootJS))
	require.Equal(t, "const name='still none';", readFile(t, nestedJS))
}

func TestRunCommandRunMissingEnvNoPartialWrites(t *testing.T) {
	if handleRunCommandHelper(t) {
		return
	}

	tempDir := t.TempDir()

	nestedDir := filepath.Join(tempDir, "nested")
	require.NoError(t, os.MkdirAll(nestedDir, 0755), "create nested dir")

	rootJS := filepath.Join(tempDir, "root.js")
	nestedJS := filepath.Join(nestedDir, "app.js")
	otherJS := filepath.Join(nestedDir, "other.js")

	originalRoot := `const missing="__reactenv.MISSING";`
	originalNested := `const present="__reactenv.PRESENT";`
	originalOther := `const another="__reactenv.ANOTHER";`

	writeFile(t, rootJS, originalRoot)
	writeFile(t, nestedJS, originalNested)
	writeFile(t, otherJS, originalOther)

	err := runCommandHelperProcess("TestRunCommandRunMissingEnvNoPartialWrites", runHelperModePartialWrite, tempDir, "PRESENT=present", "ANOTHER=another", "MISSING=")

	var exitErr *exec.ExitError
	require.ErrorAs(t, err, &exitErr)
	require.Equal(t, 1, exitErr.ExitCode())

	require.Equal(t, originalRoot, readFile(t, rootJS))
	require.Equal(t, originalNested, readFile(t, nestedJS))
	require.Equal(t, originalOther, readFile(t, otherJS))
}

func handleRunCommandHelper(t *testing.T) bool {
	t.Helper()

	if os.Getenv(runCommandHelperEnv) != "1" {
		return false
	}

	mode := os.Getenv(runCommandHelperMode)
	helperPath := os.Getenv(runCommandHelperPath)

	cmd := &RunCommand{FileMatchPattern: defaultFileMatchPattern}

	switch mode {
	case runHelperModeMissingArgs:
		cmd.Run([]string{})
	case runHelperModeInvalidMatch:
		cmd.FileMatchPattern = "["
		cmd.Run([]string{helperPath})
	case runHelperModeMissingEnv:
		cmd.Run([]string{helperPath})
	case runHelperModeNoFiles:
		cmd.Run([]string{helperPath})
	case runHelperModePartialWrite:
		cmd.Run([]string{helperPath})
	default:
		t.Fatalf("unknown helper mode: %s", mode)
	}

	return true
}

func runCommandHelperProcess(testName string, mode string, dir string, envOverrides ...string) error {
	cmd := exec.Command(os.Args[0], "-test.run="+testName)
	env := append([]string{}, os.Environ()...)
	env = append(env, runCommandHelperEnv+"=1", runCommandHelperMode+"="+mode)
	if dir != "" {
		env = append(env, runCommandHelperPath+"="+dir)
	}

	if len(envOverrides) > 0 {
		for _, override := range envOverrides {
			if strings.HasSuffix(override, "=") {
				env = filterOutEnv(env, override)
				continue
			}
			env = append(env, override)
		}
	}

	cmd.Env = env
	return cmd.Run()
}

func writeFile(t *testing.T, path string, content string) {
	t.Helper()
	require.NoError(t, os.WriteFile(path, []byte(content), 0644), "write file %s", path)
}

func readFile(t *testing.T, path string) string {
	t.Helper()
	content, err := os.ReadFile(path)
	require.NoError(t, err, "read file %s", path)
	return string(content)
}

func filterOutEnv(env []string, prefixes ...string) []string {
	filtered := make([]string, 0, len(env))
	for _, entry := range env {
		skip := false
		for _, prefix := range prefixes {
			if strings.HasPrefix(entry, prefix) {
				skip = true
				break
			}
		}
		if !skip {
			filtered = append(filtered, entry)
		}
	}
	return filtered
}
