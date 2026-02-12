package ui

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestWrapStringWrapsWithIndent(t *testing.T) {
	input := "alpha beta gamma"

	got := WrapString(input, 10, 2)

	require.Equal(t, "alpha beta\n  gamma", got)
}

func TestWrapStringExactLimitDoesNotWrap(t *testing.T) {
	input := "alpha beta"

	got := WrapString(input, 10, 2)

	require.Equal(t, input, got)
}

func TestWrapStringOneOverLimitWraps(t *testing.T) {
	input := "alpha beta"

	got := WrapString(input, 9, 2)

	require.Equal(t, "alpha\n  beta", got)
}

func TestWrapStringPreservesExplicitNewlinesWithIndent(t *testing.T) {
	input := "alpha\nbeta"

	got := WrapString(input, 10, 3)

	require.Equal(t, "alpha\n   beta", got)
}

func TestWrapStringPreservesConsecutiveNewlinesWithIndent(t *testing.T) {
	input := "alpha\n\nbeta"

	got := WrapString(input, 10, 2)

	require.Equal(t, "alpha\n  \n  beta", got)
}

func TestWrapStringPreservesLeadingAndTrailingSpacesWithinLimit(t *testing.T) {
	input := "  alpha beta  "

	got := WrapString(input, 20, 2)

	require.Equal(t, input, got)
}

func TestWrapStringTreatsTabsAsWhitespace(t *testing.T) {
	input := "alpha\tbeta"

	got := WrapString(input, 9, 2)

	require.Equal(t, "alpha\n  beta", got)
}

func TestWrapStringDropsExtraSpacesOnWrap(t *testing.T) {
	input := "alpha  beta"

	got := WrapString(input, 10, 2)

	require.Equal(t, "alpha\n  beta", got)
}

func TestWrapStringDoesNotBreakOnNBSPOrLongWord(t *testing.T) {
	input := "alpha\u00A0beta"

	got := WrapString(input, 6, 2)

	require.Equal(t, input, got)
}

func TestWrapStringDoesNotBreakLongWord(t *testing.T) {
	input := "supercalifragilisticexpialidocious"

	got := WrapString(input, 8, 2)

	require.Equal(t, input, got)
}

func TestWrapStringIndentZero(t *testing.T) {
	input := "alpha beta"

	got := WrapString(input, 9, 0)

	require.Equal(t, "alpha\nbeta", got)
}

func TestWrapAtLengthDelegatesToWrapString(t *testing.T) {
	input := "one two three"

	got := WrapAtLength(input, 1)
	expected := WrapString(input, MaxLineLength, 1)

	require.Equal(t, expected, got)
}

func TestWrapAtLengthKeepsLinesWithinLimitForWords(t *testing.T) {
	input := strings.TrimSpace(strings.Repeat("alpha beta ", 10))

	got := WrapAtLength(input, 2)
	lines := strings.Split(got, "\n")

	require.Greater(t, len(lines), 1)
	for i, line := range lines {
		if i == 0 {
			require.LessOrEqual(t, len(line), MaxLineLength)
			continue
		}

		require.True(t, strings.HasPrefix(line, "  "))
		require.LessOrEqual(t, len(strings.TrimPrefix(line, "  ")), MaxLineLength)
	}
}

func TestIndentStringAddsIndentAfterNewlines(t *testing.T) {
	input := "a\nb\n"

	got := IndentString(input, 2)

	require.Equal(t, "a\n  b\n  ", got)
}

func TestIndentStringNoNewlines(t *testing.T) {
	input := "abc"

	got := IndentString(input, 2)

	require.Equal(t, input, got)
}

func TestIndentStringEmpty(t *testing.T) {
	got := IndentString("", 3)

	require.Equal(t, "", got)
}

func TestIndentStringMultipleNewlines(t *testing.T) {
	input := "a\nb\nc"

	got := IndentString(input, 2)

	require.Equal(t, "a\n  b\n  c", got)
}

func TestPluralize(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		count    int
		expected string
	}{
		{name: "Singular", input: "item", count: 1, expected: "item"},
		{name: "Zero", input: "item", count: 0, expected: "items"},
		{name: "Negative", input: "item", count: -1, expected: "items"},
		{name: "Many", input: "item", count: 3, expected: "items"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			require.Equal(t, tc.expected, Pluralize(tc.input, tc.count))
		})
	}
}
