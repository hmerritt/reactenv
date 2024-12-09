//go:build mage

package main

import (
	"os"
	"path"

	"github.com/magefile/mage/mg"
	"github.com/magefile/mage/sh"
)

type Package mg.Namespace

// ----------------------------------------------------------------------------
// Packages
// ----------------------------------------------------------------------------

func (Package) Cli(target string) error {
	log := NewLogger()

	currentDir, _ := os.Getwd()
	cliDir := path.Join(currentDir, "packages", "cli")

	if err := RunStream([]string{"mage", "-v", target}, cliDir, true); err != nil {
		return log.Error(err)
	}

	log.End()
	return nil
}

func (Package) Webpack(target string) error {
	log := NewLogger()

	currentDir, _ := os.Getwd()
	cliDir := path.Join(currentDir, "packages", "plugin-webpack")

	if err := RunStream([]string{"yarn", target}, cliDir, true); err != nil {
		return log.Error(err)
	}

	log.End()
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
		{"go", "mod", "tidy"},
	})
}
