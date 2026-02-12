package reactenv

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestBuildFileMatcherModes(t *testing.T) {
	t.Run("AutoRegex", func(t *testing.T) {
		matcher, mode, err := buildFileMatcher(`^nested/.*\.js$`)
		require.NoError(t, err)
		require.Equal(t, fileMatchModeRegex, mode)
		require.True(t, matcher("nested/app.js"))
		require.False(t, matcher("root.js"))
	})

	t.Run("AutoGlob", func(t *testing.T) {
		matcher, mode, err := buildFileMatcher("**/*.js")
		require.NoError(t, err)
		require.Equal(t, fileMatchModeGlob, mode)
		require.True(t, matcher("nested/app.js"))
		require.True(t, matcher("app.js"))
		require.False(t, matcher("nested/app.css"))
	})

	t.Run("RegexPrefix", func(t *testing.T) {
		matcher, mode, err := buildFileMatcher("regex:^assets/.*\\.js$")
		require.NoError(t, err)
		require.Equal(t, fileMatchModeRegex, mode)
		require.True(t, matcher("assets/app.js"))
		require.False(t, matcher("app.js"))
	})

	t.Run("GlobPrefix", func(t *testing.T) {
		matcher, mode, err := buildFileMatcher("glob:*.js")
		require.NoError(t, err)
		require.Equal(t, fileMatchModeGlob, mode)
		require.True(t, matcher("app.js"))
		require.False(t, matcher("nested/app.js"))
	})

	t.Run("AutoPrefersRegexWhenBothValid", func(t *testing.T) {
		matcher, mode, err := buildFileMatcher("app.js")
		require.NoError(t, err)
		require.Equal(t, fileMatchModeRegex, mode)
		require.True(t, matcher("nested/app.js"))
	})
}

func TestBuildFileMatcherAutoInvalid(t *testing.T) {
	_, mode, err := buildFileMatcher("[")
	require.Error(t, err)

	var matchErr *FileMatchError
	require.ErrorAs(t, err, &matchErr)
	require.Equal(t, fileMatchModeAuto, mode)
	require.Equal(t, fileMatchModeAuto, matchErr.Mode)
	require.NotNil(t, matchErr.AutoRegexErr)
	require.NotNil(t, matchErr.Err)
}

func TestBuildFileMatcherPrefixInvalid(t *testing.T) {
	t.Run("RegexPrefix", func(t *testing.T) {
		_, mode, err := buildFileMatcher("regex:[")
		require.Error(t, err)

		var matchErr *FileMatchError
		require.ErrorAs(t, err, &matchErr)
		require.Equal(t, fileMatchModeRegex, mode)
		require.Equal(t, fileMatchModeRegex, matchErr.Mode)
		require.NotNil(t, matchErr.Err)
	})

	t.Run("GlobPrefix", func(t *testing.T) {
		_, mode, err := buildFileMatcher("glob:[")
		require.Error(t, err)

		var matchErr *FileMatchError
		require.ErrorAs(t, err, &matchErr)
		require.Equal(t, fileMatchModeGlob, mode)
		require.Equal(t, fileMatchModeGlob, matchErr.Mode)
		require.NotNil(t, matchErr.Err)
	})
}

func TestBuildRegexMatcherMatches(t *testing.T) {
	matcher, err := buildRegexMatcher("regex:^assets/.*\\.js$", `^assets/.*\.js$`)
	require.NoError(t, err)
	require.True(t, matcher("assets/app.js"))
	require.False(t, matcher("app.js"))
}

func TestBuildRegexMatcherInvalid(t *testing.T) {
	_, err := buildRegexMatcher("[", "[")
	require.Error(t, err)

	var matchErr *FileMatchError
	require.ErrorAs(t, err, &matchErr)
	require.Equal(t, fileMatchModeRegex, matchErr.Mode)
	require.NotNil(t, matchErr.Err)
}

func TestBuildGlobMatcherMatches(t *testing.T) {
	matcher, err := buildGlobMatcher("glob:**/*.js", "**/*.js")
	require.NoError(t, err)
	require.True(t, matcher("nested/app.js"))
	require.True(t, matcher("app.js"))
	require.False(t, matcher("nested/app.css"))
}

func TestBuildGlobMatcherInvalid(t *testing.T) {
	_, err := buildGlobMatcher("[", "[")
	require.Error(t, err)

	var matchErr *FileMatchError
	require.ErrorAs(t, err, &matchErr)
	require.Equal(t, fileMatchModeGlob, matchErr.Mode)
	require.NotNil(t, matchErr.Err)
}
