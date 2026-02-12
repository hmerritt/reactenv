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
	runCommandHelperPath = "REACTENV_RUN_PATH"
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
	if os.Getenv(runCommandHelperEnv) == "1" {
		helperPath := os.Getenv(runCommandHelperPath)
		cmd := &RunCommand{FileMatchPattern: defaultFileMatchPattern}
		cmd.Run([]string{helperPath})
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

	cmd := exec.Command(os.Args[0], "-test.run=TestRunCommandRunMissingEnvExits")
	env := append([]string{}, os.Environ()...)
	env = filterOutEnv(env, "MISSING=")
	env = append(env, runCommandHelperEnv+"=1", runCommandHelperPath+"="+tempDir, "PRESENT=present")
	cmd.Env = env

	err := cmd.Run()

	var exitErr *exec.ExitError
	require.ErrorAs(t, err, &exitErr)
	require.Equal(t, 1, exitErr.ExitCode())

	require.Equal(t, originalRoot, readFile(t, rootJS))
	require.Equal(t, originalNested, readFile(t, nestedJS))
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
