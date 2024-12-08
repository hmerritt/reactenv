//go:build mage

package main

import (
	"fmt"
	"hmerritt/reactenv/version"
	"os"
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
