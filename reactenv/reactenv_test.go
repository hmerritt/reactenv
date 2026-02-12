package reactenv

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strings"
	"testing"

	"github.com/hmerritt/reactenv/ui"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func writeTestFile(t *testing.T, dir string, name string) {
	t.Helper()
	filePath := filepath.Join(dir, name)
	require.NoError(t, os.WriteFile(filePath, []byte("test"), 0644), "write test file %s", filePath)
}

func TestReactenvFindFilesMatchesFilesAndIgnoresDirs(t *testing.T) {
	tempDir := t.TempDir()

	writeTestFile(t, tempDir, "alpha.js")
	writeTestFile(t, tempDir, "beta.js")
	writeTestFile(t, tempDir, "gamma.css")

	matchingDir := filepath.Join(tempDir, "dir.js")
	require.NoError(t, os.Mkdir(matchingDir, 0755), "create matching dir")

	renv := NewReactenv(nil)

	require.NoError(t, renv.FindFiles(tempDir, `.*\.js$`), "FindFiles returned error")

	require.Equal(t, tempDir, renv.Dir, "Dir")

	require.Equal(t, 2, renv.FilesMatchTotal, "FilesMatchTotal")

	require.Len(t, renv.Files, 2, "Files length")
	require.Len(t, renv.FileRelPaths, 2, "FileRelPaths length")

	found := map[string]bool{}
	for _, file := range renv.Files {
		found[(*file).Name()] = true
	}

	expected := map[string]bool{
		"alpha.js": true,
		"beta.js":  true,
	}

	require.Equal(t, expected, found, "matched files")

	relFound := map[string]bool{}
	for _, relPath := range renv.FileRelPaths {
		relFound[relPath] = true
	}

	require.Equal(t, expected, relFound, "matched relative paths")
}

func TestReactenvFindFilesMatchesFilesRecursively(t *testing.T) {
	tempDir := t.TempDir()

	writeTestFile(t, tempDir, "root.js")

	nestedDir := filepath.Join(tempDir, "nested")
	require.NoError(t, os.Mkdir(nestedDir, 0755), "create nested dir")
	writeTestFile(t, nestedDir, "nested.js")

	deepDir := filepath.Join(nestedDir, "deep")
	require.NoError(t, os.Mkdir(deepDir, 0755), "create deep dir")
	writeTestFile(t, deepDir, "deep.js")

	renv := NewReactenv(nil)

	require.NoError(t, renv.FindFiles(tempDir, `.*\.js$`), "FindFiles returned error")
	require.Equal(t, 3, renv.FilesMatchTotal, "FilesMatchTotal")
	require.Len(t, renv.Files, 3, "Files length")
	require.Len(t, renv.FileRelPaths, 3, "FileRelPaths length")

	expected := map[string]bool{
		"root.js":             true,
		"nested/nested.js":    true,
		"nested/deep/deep.js": true,
	}

	found := map[string]bool{}
	for _, relPath := range renv.FileRelPaths {
		found[relPath] = true
	}

	require.Equal(t, expected, found, "matched relative paths")
}

func TestReactenvFindFilesSkipsNodeModules(t *testing.T) {
	tempDir := t.TempDir()

	writeTestFile(t, tempDir, "root.js")

	nodeModulesDir := filepath.Join(tempDir, "node_modules")
	require.NoError(t, os.Mkdir(nodeModulesDir, 0755), "create node_modules dir")
	writeTestFile(t, nodeModulesDir, "ignored.js")

	nestedDir := filepath.Join(tempDir, "nested")
	require.NoError(t, os.Mkdir(nestedDir, 0755), "create nested dir")
	writeTestFile(t, nestedDir, "nested.js")

	renv := NewReactenv(nil)

	require.NoError(t, renv.FindFiles(tempDir, `.*\.js$`), "FindFiles returned error")

	expected := map[string]bool{
		"root.js":          true,
		"nested/nested.js": true,
	}

	found := map[string]bool{}
	for _, relPath := range renv.FileRelPaths {
		found[relPath] = true
	}

	require.Equal(t, expected, found, "matched relative paths")
}

func TestReactenvFindFilesReturnsErrorForMissingDir(t *testing.T) {
	renv := NewReactenv(nil)
	missingDir := filepath.Join(t.TempDir(), "missing")

	require.Error(t, renv.FindFiles(missingDir, `.*`), "expected error for missing directory")
}

func TestReactenvFindFilesReturnsErrorForBadRegex(t *testing.T) {
	renv := NewReactenv(nil)
	tempDir := t.TempDir()

	require.Error(t, renv.FindFiles(tempDir, `[`), "expected error for invalid regex")
}

func TestReactenvFilesWalkCallsCallbackInOrder(t *testing.T) {
	tempDir := t.TempDir()

	writeTestFile(t, tempDir, "b.js")
	writeTestFile(t, tempDir, "a.js")
	writeTestFile(t, tempDir, "c.css")

	renv := NewReactenv(nil)
	require.NoError(t, renv.FindFiles(tempDir, `.*\.js$`))

	type walkCall struct {
		index    int
		name     string
		filePath string
	}
	calls := make([]walkCall, 0)

	err := renv.FilesWalk(func(fileIndex int, file os.DirEntry, filePath string) error {
		calls = append(calls, walkCall{
			index:    fileIndex,
			name:     file.Name(),
			filePath: filePath,
		})
		return nil
	})
	require.NoError(t, err)

	expected := []walkCall{
		{index: 0, name: "a.js", filePath: path.Join(tempDir, "a.js")},
		{index: 1, name: "b.js", filePath: path.Join(tempDir, "b.js")},
	}

	require.Equal(t, expected, calls)
}

func TestReactenvFilesWalkStopsOnError(t *testing.T) {
	tempDir := t.TempDir()

	writeTestFile(t, tempDir, "first.js")
	writeTestFile(t, tempDir, "second.js")

	renv := NewReactenv(nil)
	require.NoError(t, renv.FindFiles(tempDir, `.*\.js$`))

	callCount := 0
	expectedErr := os.ErrInvalid

	err := renv.FilesWalk(func(fileIndex int, file os.DirEntry, filePath string) error {
		callCount++
		return expectedErr
	})

	require.ErrorIs(t, err, expectedErr)
	require.Equal(t, 1, callCount)
}

func TestReactenvFilesWalkContentsCallsCallbackInOrderWithContents(t *testing.T) {
	tempDir := t.TempDir()

	require.NoError(t, os.WriteFile(filepath.Join(tempDir, "b.js"), []byte("beta"), 0644))
	require.NoError(t, os.WriteFile(filepath.Join(tempDir, "a.js"), []byte("alpha"), 0644))
	writeTestFile(t, tempDir, "c.css")

	renv := NewReactenv(nil)
	require.NoError(t, renv.FindFiles(tempDir, `.*\.js$`))

	type walkCall struct {
		index    int
		name     string
		filePath string
		contents string
	}
	calls := make([]walkCall, 0)

	err := renv.FilesWalkContents(func(fileIndex int, file os.DirEntry, filePath string, fileContents []byte) error {
		calls = append(calls, walkCall{
			index:    fileIndex,
			name:     file.Name(),
			filePath: filePath,
			contents: string(fileContents),
		})
		return nil
	})
	require.NoError(t, err)

	expected := []walkCall{
		{index: 0, name: "a.js", filePath: path.Join(tempDir, "a.js"), contents: "alpha"},
		{index: 1, name: "b.js", filePath: path.Join(tempDir, "b.js"), contents: "beta"},
	}

	require.Equal(t, expected, calls)
}

func TestReactenvFilesWalkContentsStopsOnError(t *testing.T) {
	tempDir := t.TempDir()

	writeTestFile(t, tempDir, "first.js")
	writeTestFile(t, tempDir, "second.js")

	renv := NewReactenv(nil)
	require.NoError(t, renv.FindFiles(tempDir, `.*\.js$`))

	callCount := 0
	expectedErr := os.ErrInvalid

	err := renv.FilesWalkContents(func(fileIndex int, file os.DirEntry, filePath string, fileContents []byte) error {
		callCount++
		return expectedErr
	})

	require.ErrorIs(t, err, expectedErr)
	require.Equal(t, 1, callCount)
}

func TestReactenvFilesWalkContentsReadErrorExits(t *testing.T) {
	cmd := exec.Command(os.Args[0], "-test.run=TestReactenvFilesWalkContentsReadErrorHelper")
	cmd.Env = append(os.Environ(), "REACTENV_HELPER_PROCESS=1")
	err := cmd.Run()

	var exitErr *exec.ExitError
	require.ErrorAs(t, err, &exitErr)
	require.Equal(t, 1, exitErr.ExitCode())
}

func TestReactenvFilesWalkContentsReadErrorHelper(t *testing.T) {
	if os.Getenv("REACTENV_HELPER_PROCESS") != "1" {
		t.Skip("helper process only")
	}

	tempDir := t.TempDir()
	writeTestFile(t, tempDir, "missing.js")

	renv := NewReactenv(ui.GetUi())
	require.NoError(t, renv.FindFiles(tempDir, `.*\.js$`))
	require.NoError(t, os.Remove(filepath.Join(tempDir, "missing.js")))

	renv.FilesWalkContents(func(fileIndex int, file os.DirEntry, filePath string, fileContents []byte) error {
		return nil
	})
}

func TestFindAllOccurrenceBytePositions(t *testing.T) {
	prefix := fmt.Appendf([]byte(""), "%s.", REACTENV_PREFIX)

	tests := []struct {
		name     string
		input    string
		expected [][]int
	}{
		{
			name:     "Single isolated variable",
			input:    "__reactenv.MY_VAR",
			expected: [][]int{{0, 17}},
		},
		{
			name:     "Contiguous variables (the primary edge case)",
			input:    "__reactenv.FIRST_VAR__reactenv.SECOND_VAR",
			expected: [][]int{{0, 20}, {20, 41}},
		},
		{
			name:     "Rejection of leading numeral",
			input:    "__reactenv.1INVALID",
			expected: nil, // The function should bypass this entirely.
		},
		{
			name:     "Rejection of leading dollar sign",
			input:    "__reactenv.$INVALID",
			expected: nil,
		},
		{
			name:     "Rejection of leading percent sign",
			input:    "__reactenv.%INVALID",
			expected: nil,
		},
		{
			name:     "Rejection of leading exclamation mark",
			input:    "__reactenv.!INVALID",
			expected: nil,
		},
		{
			name:     "Rejection of leading ampersand",
			input:    "__reactenv.&INVALID",
			expected: nil,
		},
		{
			name:     "Rejection of leading asterisk",
			input:    "__reactenv.*INVALID",
			expected: nil,
		},
		{
			name:     "Rejection of leading open parenthesis",
			input:    "__reactenv.(INVALID",
			expected: nil,
		},
		{
			name:     "Rejection of leading close parenthesis",
			input:    "__reactenv.)INVALID",
			expected: nil,
		},
		{
			name:     "Rejection of leading open square bracket",
			input:    "__reactenv.[INVALID",
			expected: nil,
		},
		{
			name:     "Rejection of leading close square bracket",
			input:    "__reactenv.]INVALID",
			expected: nil,
		},
		{
			name:     "Rejection of leading open curly brace",
			input:    "__reactenv.{INVALID",
			expected: nil,
		},
		{
			name:     "Rejection of leading close curly brace",
			input:    "__reactenv.}INVALID",
			expected: nil,
		},
		{
			name:     "Acceptance of leading underscore",
			input:    "__reactenv._VALID",
			expected: [][]int{{0, 17}},
		},
		{
			name:     "Early termination upon encountering invalid characters",
			input:    "__reactenv.VALID-INVALID",
			expected: [][]int{{0, 16}}, // Halts exactly before the hyphen.
		},
		{
			name:     "Premature string termination",
			input:    "Incomplete text __reactenv.",
			expected: nil, // Must not panic when the string ends at the prefix.
		},
		{
			name:     "Complete absence of occurrences",
			input:    "Standard text devoid of any environment variables.",
			expected: nil,
		},
		{
			name:     "Multiple spaced variables",
			input:    "Start __reactenv.ONE middle __reactenv.TWO end",
			expected: [][]int{{6, 20}, {28, 42}},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			actual := FindAllOccurrenceBytePositions([]byte(tc.input), prefix)

			// testify/assert will recursively compare the multidimensional slices
			// and output a detailed diff if an index mismatch occurs.
			assert.Equal(t, tc.expected, actual, "Mismatch in calculated indices.")
		})
	}
}

func FuzzFindAllOccurrenceBytePositions(f *testing.F) {
	// Establish the Seed Corpus
	f.Add([]byte("__reactenv.MY_VAR"), []byte("__reactenv."))
	f.Add([]byte("__reactenv.FIRST__reactenv.SECOND"), []byte("__reactenv."))
	f.Add([]byte("__reactenv.1INVALID"), []byte("__reactenv."))
	f.Add([]byte("Standard text devoid of variables"), []byte("__reactenv."))
	f.Add([]byte(""), []byte("__reactenv."))
	f.Add([]byte("__reactenv.VALID"), []byte("")) // Tests our new empty-prefix guard clause.

	f.Fuzz(func(t *testing.T, data []byte, prefix []byte) {
		// Simply executing the function proves it does not panic.
		indices := FindAllOccurrenceBytePositions(data, prefix)

		// Validate Invariants on the returned data.
		for _, match := range indices {
			start, end := match[0], match[1]

			// Invariant A: Bounds Safety
			// The indices must never exceed the slice capacity, and the
			// start must logically precede the end.
			if start < 0 || end > len(data) || start >= end {
				t.Errorf("Bounds violation: generated indices [%d, %d] invalid for data length %d", start, end, len(data))
			}

			// Invariant B: Prefix Integrity
			// If a valid match was claimed, the resulting string MUST
			// physically begin with the requested prefix.
			extracted := data[start:end]
			if !bytes.HasPrefix(extracted, prefix) {
				t.Errorf("Prefix violation: extracted %q does not begin with %q", extracted, prefix)
			}

			// Invariant C: Content Capture
			// The function must have captured at least one valid character
			// beyond the length of the prefix itself.
			if len(extracted) <= len(prefix) {
				t.Errorf("Length violation: extracted %q contains no variable identifier", extracted)
			}
		}
	})
}

func TestReactenvFindOccurrencesPopulatesFields(t *testing.T) {
	tempDir := t.TempDir()
	content := "a=__reactenv.FIRST;b=__reactenv.SECOND;c=__reactenv.FIRST;d=__reactenv._THIRD1;e=__reactenv.1BAD;"
	filePath := filepath.Join(tempDir, "app.js")

	require.NoError(t, os.WriteFile(filePath, []byte(content), 0644))

	t.Setenv("FIRST", "one")
	t.Setenv("_THIRD1", "three")

	renv := NewReactenv(nil)
	require.NoError(t, renv.FindFiles(tempDir, `.*\.js$`))

	require.NoError(t, renv.FindOccurrences())

	expectedTokens := []string{
		"__reactenv.FIRST",
		"__reactenv.SECOND",
		"__reactenv.FIRST",
		"__reactenv._THIRD1",
	}
	expectedOccurrences := occurrencesFromTokens(t, content, expectedTokens)

	require.Equal(t, 4, renv.OccurrencesTotal)
	require.Len(t, renv.Files, 1)
	require.Equal(t, "app.js", (*renv.Files[0]).Name())
	require.Len(t, renv.OccurrencesByFile, 1)
	require.Equal(t, expectedOccurrences, renv.OccurrencesByFile[0].Occurrences)

	expectedKeys := map[string]bool{
		"FIRST":   true,
		"SECOND":  true,
		"_THIRD1": true,
	}
	require.Equal(t, expectedKeys, renv.OccurrenceKeys)

	expectedReplacement := map[string]string{
		"FIRST":   "one",
		"_THIRD1": "three",
	}
	require.Equal(t, expectedReplacement, renv.OccurrenceKeysReplacement)
}

func TestReactenvFindOccurrencesFiltersFilesWithoutMatches(t *testing.T) {
	tempDir := t.TempDir()

	require.NoError(t, os.WriteFile(filepath.Join(tempDir, "a.js"), []byte("__reactenv.A"), 0644))
	require.NoError(t, os.WriteFile(filepath.Join(tempDir, "b.js"), []byte("no matches"), 0644))
	require.NoError(t, os.WriteFile(filepath.Join(tempDir, "c.js"), []byte("__reactenv.B__reactenv.A"), 0644))

	t.Setenv("A", "value-a")

	renv := NewReactenv(nil)
	require.NoError(t, renv.FindFiles(tempDir, `.*\.js$`))

	require.NoError(t, renv.FindOccurrences())

	require.Equal(t, 3, renv.OccurrencesTotal)
	require.Len(t, renv.Files, 2)
	require.Len(t, renv.FileRelPaths, 2)
	require.Len(t, renv.OccurrencesByFile, 2)

	counts := map[string]int{}
	for i, file := range renv.Files {
		counts[(*file).Name()] = len(renv.OccurrencesByFile[i].Occurrences)
	}

	require.Equal(t, map[string]int{
		"a.js": 1,
		"c.js": 2,
	}, counts)

	paths := map[string]bool{}
	for _, relPath := range renv.FileRelPaths {
		paths[relPath] = true
	}

	require.Equal(t, map[string]bool{"a.js": true, "c.js": true}, paths)

	expectedKeys := map[string]bool{
		"A": true,
		"B": true,
	}
	require.Equal(t, expectedKeys, renv.OccurrenceKeys)

	expectedReplacement := map[string]string{
		"A": "value-a",
	}
	require.Equal(t, expectedReplacement, renv.OccurrenceKeysReplacement)
}

func TestReactenvFindOccurrencesNoMatchesClearsFiles(t *testing.T) {
	tempDir := t.TempDir()

	writeTestFile(t, tempDir, "a.js")
	writeTestFile(t, tempDir, "b.js")

	renv := NewReactenv(nil)
	require.NoError(t, renv.FindFiles(tempDir, `.*\.js$`))

	require.NoError(t, renv.FindOccurrences())

	require.Equal(t, 0, renv.OccurrencesTotal)
	require.Empty(t, renv.Files)
	require.Empty(t, renv.FileRelPaths)
	require.Empty(t, renv.OccurrencesByFile)
	require.Empty(t, renv.OccurrenceKeys)
	require.Empty(t, renv.OccurrenceKeysReplacement)
}

func TestReactenvFindOccurrencesResetsStateOnRepeat(t *testing.T) {
	tempDir := t.TempDir()
	filePath := filepath.Join(tempDir, "app.js")

	require.NoError(t, os.WriteFile(filePath, []byte("__reactenv.ONE"), 0644))
	t.Setenv("ONE", "1")

	renv := NewReactenv(nil)
	require.NoError(t, renv.FindFiles(tempDir, `.*\.js$`))

	require.NoError(t, renv.FindOccurrences())

	require.Equal(t, 1, renv.OccurrencesTotal)
	require.Equal(t, map[string]bool{"ONE": true}, renv.OccurrenceKeys)
	require.Equal(t, map[string]string{"ONE": "1"}, renv.OccurrenceKeysReplacement)
	require.Len(t, renv.Files, 1)
	require.Len(t, renv.FileRelPaths, 1)
	require.Len(t, renv.OccurrencesByFile, 1)

	require.NoError(t, os.WriteFile(filePath, []byte("no occurrences"), 0644))

	require.NoError(t, renv.FindOccurrences())

	require.Equal(t, 0, renv.OccurrencesTotal)
	require.Empty(t, renv.Files)
	require.Empty(t, renv.FileRelPaths)
	require.Empty(t, renv.OccurrencesByFile)
	require.Empty(t, renv.OccurrenceKeys)
	require.Empty(t, renv.OccurrenceKeysReplacement)
}

func FuzzReactenvFindOccurrences(f *testing.F) {
	f.Add("")
	f.Add("no matches here")
	f.Add("__reactenv.A")
	f.Add("__reactenv.A__reactenv.B")
	f.Add("__reactenv.1BAD")

	f.Fuzz(func(t *testing.T, content string) {
		tempDir := t.TempDir()
		filePath := filepath.Join(tempDir, "app.js")
		require.NoError(t, os.WriteFile(filePath, []byte(content), 0644))

		renv := NewReactenv(nil)
		require.NoError(t, renv.FindFiles(tempDir, `.*\.js$`))

		require.NoError(t, renv.FindOccurrences())

		prefix := fmt.Appendf([]byte(""), "%s.", REACTENV_PREFIX)
		matches := FindAllOccurrenceBytePositions([]byte(content), prefix)
		expectedTotal := len(matches)

		require.Equal(t, expectedTotal, renv.OccurrencesTotal)
		require.Equal(t, len(renv.Files), len(renv.OccurrencesByFile))

		if expectedTotal == 0 {
			require.Empty(t, renv.Files)
			require.Empty(t, renv.OccurrencesByFile)
			require.Empty(t, renv.OccurrenceKeys)
			return
		}

		require.Len(t, renv.Files, 1)
		require.Len(t, renv.OccurrencesByFile, 1)
		require.Len(t, renv.OccurrencesByFile[0].Occurrences, expectedTotal)

		expectedKeys := map[string]bool{}
		for i, match := range matches {
			occurrence := renv.OccurrencesByFile[0].Occurrences[i]
			require.Equal(t, match, occurrence.StartEnd)

			occurrenceText := content[match[0]:match[1]]
			key := strings.TrimPrefix(occurrenceText, "__reactenv.")
			expectedKeys[key] = true
			require.Equal(t, key, occurrence.Key)
		}

		require.Equal(t, expectedKeys, renv.OccurrenceKeys)

		for key, value := range renv.OccurrenceKeysReplacement {
			envValue, ok := os.LookupEnv(key)
			require.True(t, ok)
			require.Equal(t, envValue, value)
		}
	})
}

func TestReactenvReplaceOccurrencesReplacesMatches(t *testing.T) {
	tempDir := t.TempDir()
	content := "start __reactenv.ONE mid __reactenv.TWO end __reactenv.ONE!"
	filePath := filepath.Join(tempDir, "app.js")

	require.NoError(t, os.WriteFile(filePath, []byte(content), 0644))
	t.Setenv("ONE", "1")
	t.Setenv("TWO", "two")

	renv := NewReactenv(nil)
	require.NoError(t, renv.FindFiles(tempDir, `.*\.js$`))
	require.NoError(t, renv.FindOccurrences())

	renv.ReplaceOccurrences()

	updated, err := os.ReadFile(filePath)
	require.NoError(t, err)
	require.Equal(t, "start 1 mid two end 1!", string(updated))
}

func TestReactenvReplaceOccurrencesHandlesAdjacentMatches(t *testing.T) {
	tempDir := t.TempDir()
	content := "__reactenv.ONE__reactenv.TWO"
	filePath := filepath.Join(tempDir, "app.js")

	require.NoError(t, os.WriteFile(filePath, []byte(content), 0644))
	t.Setenv("ONE", "first")
	t.Setenv("TWO", "second")

	renv := NewReactenv(nil)
	require.NoError(t, renv.FindFiles(tempDir, `.*\.js$`))
	require.NoError(t, renv.FindOccurrences())

	renv.ReplaceOccurrences()

	updated, err := os.ReadFile(filePath)
	require.NoError(t, err)
	require.Equal(t, "firstsecond", string(updated))
}

func TestReactenvReplaceOccurrencesUsesEmptyStringForMissingValues(t *testing.T) {
	tempDir := t.TempDir()
	content := "x__reactenv.MISSINGy"
	filePath := filepath.Join(tempDir, "app.js")

	require.NoError(t, os.WriteFile(filePath, []byte(content), 0644))

	renv := NewReactenv(nil)
	require.NoError(t, renv.FindFiles(tempDir, `.*\.js$`))
	require.NoError(t, renv.FindOccurrences())

	renv.ReplaceOccurrences()

	updated, err := os.ReadFile(filePath)
	require.NoError(t, err)
	require.Equal(t, "x", string(updated))
}

func TestReactenvReplaceOccurrencesNoMatchesIsNoop(t *testing.T) {
	tempDir := t.TempDir()
	content := "no matches"
	filePath := filepath.Join(tempDir, "app.js")

	require.NoError(t, os.WriteFile(filePath, []byte(content), 0644))

	renv := NewReactenv(nil)
	require.NoError(t, renv.FindFiles(tempDir, `.*\.js$`))
	require.NoError(t, renv.FindOccurrences())

	renv.ReplaceOccurrences()

	updated, err := os.ReadFile(filePath)
	require.NoError(t, err)
	require.Equal(t, content, string(updated))
}

func TestReactenvReplaceOccurrencesMultipleFiles(t *testing.T) {
	tempDir := t.TempDir()
	fileA := filepath.Join(tempDir, "a.js")
	fileB := filepath.Join(tempDir, "b.js")

	require.NoError(t, os.WriteFile(fileA, []byte("a=__reactenv.A"), 0644))
	require.NoError(t, os.WriteFile(fileB, []byte("b=__reactenv.B"), 0644))

	t.Setenv("A", "alpha")
	t.Setenv("B", "beta")

	renv := NewReactenv(nil)
	require.NoError(t, renv.FindFiles(tempDir, `.*\.js$`))
	require.NoError(t, renv.FindOccurrences())

	renv.ReplaceOccurrences()

	updatedA, err := os.ReadFile(fileA)
	require.NoError(t, err)
	require.Equal(t, "a=alpha", string(updatedA))

	updatedB, err := os.ReadFile(fileB)
	require.NoError(t, err)
	require.Equal(t, "b=beta", string(updatedB))
}

func FuzzReactenvReplaceOccurrences(f *testing.F) {
	f.Add("")
	f.Add("plain text")
	f.Add("__reactenv.A")
	f.Add("__reactenv.A__reactenv.B")
	f.Add("x__reactenv._A1y__reactenv.$B2z")
	f.Add("__reactenv.1BAD__reactenv.GOOD")

	f.Fuzz(func(t *testing.T, content string) {
		tempDir := t.TempDir()
		filePath := filepath.Join(tempDir, "app.js")
		require.NoError(t, os.WriteFile(filePath, []byte(content), 0644))

		for _, key := range keysFromContent(content) {
			t.Setenv(key, "value-"+key)
		}

		renv := NewReactenv(nil)
		require.NoError(t, renv.FindFiles(tempDir, `.*\.js$`))
		require.NoError(t, renv.FindOccurrences())

		renv.ReplaceOccurrences()

		updated, err := os.ReadFile(filePath)
		require.NoError(t, err)

		expected := applyExpectedReplacements(content, renv.OccurrencesByFile, renv.OccurrenceKeysReplacement)
		require.Equal(t, expected, string(updated))

		remaining := FindAllOccurrenceBytePositions(updated, []byte(REACTENV_PREFIX+"."))
		require.Empty(t, remaining)
	})
}

func occurrencesFromTokens(t *testing.T, content string, tokens []string) []Occurrence {
	t.Helper()

	occurrences := make([]Occurrence, 0, len(tokens))
	searchStart := 0

	for _, token := range tokens {
		index := strings.Index(content[searchStart:], token)
		require.NotEqual(t, -1, index, "token %s not found", token)

		start := searchStart + index
		end := start + len(token)
		key := strings.TrimPrefix(token, "__reactenv.")

		occurrences = append(occurrences, Occurrence{
			Key:      key,
			StartEnd: []int{start, end},
		})

		searchStart = end
	}

	return occurrences
}

func applyExpectedReplacements(content string, occurrencesByFile []*FileOccurrences, replacements OccurrenceKeysReplacement) string {
	if len(occurrencesByFile) == 0 {
		return content
	}

	occurrences := occurrencesByFile[0].Occurrences
	if len(occurrences) == 0 {
		return content
	}

	contentBytes := []byte(content)
	result := make([]byte, 0, len(contentBytes))
	lastIndex := 0
	for _, occurrence := range occurrences {
		start, end := occurrence.StartEnd[0], occurrence.StartEnd[1]
		result = append(result, contentBytes[lastIndex:start]...)
		result = append(result, replacements[occurrence.Key]...)
		lastIndex = end
	}
	result = append(result, contentBytes[lastIndex:]...)
	return string(result)
}

func keysFromContent(content string) []string {
	positions := FindAllOccurrenceBytePositions([]byte(content), []byte(REACTENV_PREFIX+"."))
	keys := make([]string, 0, len(positions))
	for _, pos := range positions {
		token := content[pos[0]:pos[1]]
		key := strings.TrimPrefix(token, REACTENV_PREFIX+".")
		keys = append(keys, key)
	}
	return keys
}
