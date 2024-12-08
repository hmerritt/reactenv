//go:build mage

package main

import (
	"fmt"
	"hmerritt/reactenv/version"
	"os"
	"regexp"
	"strconv"
	"strings"

	"github.com/magefile/mage/mg"
	"github.com/magefile/mage/sh"
)

type Build mg.Namespace

var Aliases = map[string]interface{}{
	"build": Build.Release,
}

// ----------------------------------------------------------------------------
// Test
// ----------------------------------------------------------------------------

// Runs Go tests
func Test() error {
	log := NewLogger()
	defer log.End()
	return RunSync([][]string{
		{"gotestsum", "--format", "pkgname", "--", "--cover", "./..."},
	})
}

func Bench() error {
	log := NewLogger()
	defer log.End()
	return RunSync([][]string{
		{"gotestsum", "--format", "pkgname", "--", "--cover", "-bench", ".", "-benchmem", "./..."},
	})
}

// ----------------------------------------------------------------------------
// Build
// ----------------------------------------------------------------------------

func (Build) Debug() error {
	log := NewLogger()
	defer log.End()
	log.Info("compiling debug binary")
	return RunSync([][]string{
		{"go", "build", "-ldflags", "-s -w", "."},
	})
}

func (Build) Release() error {
	log := NewLogger()
	defer log.End()
	log.Info("compiling release binary")
	return RunSync([][]string{
		{"gox",
			"-osarch",
			"darwin/amd64 linux/amd64 linux/arm64 windows/amd64",
			"-gocmd",
			"go",
			"-ldflags",
			LdFlagString(),
			"-tags",
			"reactenv",
			"-output",
			"bin/{{.OS}}_{{.Arch}}/reactenv",
			"."},
	})
}

// ----------------------------------------------------------------------------
// Release
// ----------------------------------------------------------------------------

func Release() error {
	log := NewLogger()
	defer log.End()

	releaseVersion := GetEnv("RELEASE_VERSION", version.Version)
	log.Info("release version: ", releaseVersion)

	releaseArchs := []string{"darwin_amd64", "linux_amd64", "linux_arm64", "windows_amd64"}

	for _, arch := range releaseArchs {
		binDirPath := fmt.Sprintf("bin/%s", arch)
		binFilePath := ""

		// Search each binary path for the release binary,
		// ensure binary file exists and check it's file size.
		log.Info("checking release binary for:", arch)
		binPathfiles, err := os.ReadDir(binDirPath)
		if err != nil {
			return log.Error("error reading directory:", arch, err)
		}
		for _, file := range binPathfiles {
			if file.IsDir() {
				continue
			}
			if strings.Contains(file.Name(), "reactenv") {
				fileInfo, err := file.Info()
				if err != nil {
					return log.Error("error getting file info:", arch, err)
				}
				if fileInfo.Size() < 1000000 {
					return log.Error("release binary is too small:", arch, err)
				}
				binFilePath = fmt.Sprintf("%s/%s", binDirPath, file.Name())
				break
			}
		}
		if binFilePath == "" {
			return log.Error("failed to find release binary", arch)
		}

		log.Info("zip for release")
		zipFileName := fmt.Sprintf("reactenv_%s_%s.zip", releaseVersion, arch)
		zipFilePath := fmt.Sprintf("bin/%s", zipFileName)

		err = ZipFiles(zipFilePath, binFilePath)
		if err != nil {
			return log.Error("failed to zip binary", err)
		}
	}

	return nil
}

// ----------------------------------------------------------------------------
// Housekeeping
// ----------------------------------------------------------------------------

// Bootstraps required packages (installs required linux/macOS packages if needed)
func Bootstrap() error {
	log := NewLogger()
	defer log.End()

	// Install mage bootstrap (the recommended, as seen in https://magefile.org)
	if !ExecExists("mage") && ExecExists("git") {
		log.Info("installing mage")
		tmpDir := "__tmp_mage"

		if err := sh.Run("git", "clone", "https://github.com/magefile/mage", tmpDir); err != nil {
			return log.Error("error: installing mage: ", err)
		}

		if err := os.Chdir(tmpDir); err != nil {
			return log.Error("error: installing mage: ", err)
		}

		if err := sh.Run("go", "run", "bootstrap.go"); err != nil {
			return log.Error("error: installing mage: ", err)
		}

		if err := os.Chdir("../"); err != nil {
			return log.Error("error: installing mage: ", err)
		}

		os.RemoveAll(tmpDir)
	}

	// Install Go dependencies
	log.Info("installing go dependencies")
	return RunSync([][]string{
		{"go", "mod", "vendor"},
		{"go", "mod", "tidy"},
		{"go", "generate", "-tags", "tools", "tools/tools.go"},
	})
}

// Update all Go dependencies
func UpdateDeps() error {
	log := NewLogger()
	defer log.End()
	return RunSync([][]string{
		{"go", "get", "-u", "all"},
		{"go", "mod", "vendor"},
		{"go", "mod", "tidy"},
	})
}

// Bumps (syncs) patch version to the commit count (see `version/version_base.go`)
func BumpVersion() error {
	log := NewLogger()
	defer log.End()

	// Get the total commit count
	commitCountString, err := sh.Output("git", "rev-list", "--count", "HEAD")

	if err != nil {
		return log.Error("failed to get commit count:", err)
	}

	commitCount, err := strconv.Atoi(commitCountString)

	if err != nil {
		return log.Error("failed to parse commit count:", commitCountString, err)
	}

	versionFile := "version/version_base.go"
	versionFileContent, err := os.ReadFile(versionFile)

	if err != nil {
		return log.Error("failed open version file:", err)
	}

	// Extract current patch version
	versionMatch := regexp.MustCompile(`= "(\d+).(\d+).(\d+)"`).FindStringSubmatch(string(versionFileContent))

	if len(versionMatch) < 4 {
		return log.Error("failed to parse version:", versionFile, versionMatch)
	}

	majorVersion := versionMatch[1]
	minorVersion := versionMatch[2]
	patchVersionCurrent := versionMatch[3]
	patchVersion := commitCount

	versionCurrent := fmt.Sprintf("%s.%s.%s", majorVersion, minorVersion, patchVersionCurrent)
	versionNew := fmt.Sprintf("%s.%s.%d", majorVersion, minorVersion, patchVersion)
	log.Info("bumping version", versionCurrent, "->", versionNew)

	// Update version file
	versionFileContent = regexp.MustCompile(`= "\d+.\d+.\d+"`).ReplaceAll(versionFileContent, []byte(fmt.Sprintf(`= "%s.%s.%d"`, majorVersion, minorVersion, patchVersion)))

	if err := os.WriteFile(versionFile, versionFileContent, 0644); err != nil {
		return log.Error("failed to write version file:", err)
	}

	return nil
}
