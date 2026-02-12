//go:build mage

package main

import (
	"fmt"
	"io/fs"
	"os"
	"path"
	"regexp"
	"strconv"
	"strings"

	"github.com/hmerritt/reactenv/version"

	"github.com/magefile/mage/mg"
	"github.com/magefile/mage/sh"
)

type Build mg.Namespace
type Release mg.Namespace
type Npm mg.Namespace

var Aliases = map[string]interface{}{
	"build": Build.Release,
}

// ----------------------------------------------------------------------------
// Npm packages - run (yarn) commands within each npm directory
// ----------------------------------------------------------------------------

func (Npm) Webpack(target string) error {
	log := NewLogger()

	currentDir, _ := os.Getwd()
	cliDir := path.Join(currentDir, "npm", "plugin-webpack")

	if err := RunStream([]string{"yarn", target}, cliDir, true); err != nil {
		return log.Error(err)
	}

	log.End()
	return nil
}

// ----------------------------------------------------------------------------
// Test
// ----------------------------------------------------------------------------

// Runs Go tests
func Test() error {
	log := NewLogger()
	defer log.End()
	return RunSync([][]string{
		{"gotestsum", "--format", "pkgname", "--", "-coverprofile", "cover.out", "./..."},
	})
}

func TestWatch() error {
	log := NewLogger()
	defer log.End()
	return RunSync([][]string{
		{"gotestsum", "--watch", "--format", "pkgname", "--", "./..."},
	})
}

// Prints coverage report
func Coverage() error {
	log := NewLogger()
	defer log.End()
	return RunSync([][]string{
		{"go-ignore-cov", "--file", "cover.out"},
		{"go", "tool", "cover", "-func", "cover.out"},
	})
}

func Bench() error {
	log := NewLogger()
	defer log.End()
	return RunSync([][]string{
		{"gotestsum", "--format", "pkgname", "--", "-bench", ".", "-benchmem", "./..."},
	})
}

// ----------------------------------------------------------------------------
// Build
// ----------------------------------------------------------------------------

func (Build) Clean() error {
	log := NewLogger()
	defer log.End()
	log.Info("cleaning bin and dist directories")

	if err := os.RemoveAll("bin"); err != nil {
		return log.Error(err)
	}

	if err := os.RemoveAll("dist"); err != nil {
		return log.Error(err)
	}

	log.Info("cleaning npm directories")

	return fs.WalkDir(os.DirFS("npm"), ".", func(filePath string, d fs.DirEntry, err error) error {
		if d.IsDir() ||
			filePath == "reactenv/bin/reactenv" ||
			!strings.HasPrefix(filePath, "reactenv") ||
			!strings.HasPrefix(d.Name(), "reactenv") {
			return nil
		}
		return os.Remove(path.Join("npm", filePath))
	})
}

func (Build) Debug() error {
	log := NewLogger()
	defer log.End()
	log.Info("compiling debug binary")
	return RunSync([][]string{
		{"go", "build", "-ldflags", "-s -w", "."},
	})
}

func (Build) Release() error {
	mg.Deps(Build.Clean)

	log := NewLogger()
	defer log.End()
	log.Info("compiling release binary")
	return RunSync([][]string{
		{"gox",
			"-osarch",
			"darwin/amd64 darwin/arm64 freebsd/amd64 freebsd/arm linux/amd64 linux/arm64 netbsd/amd64 netbsd/arm openbsd/amd64 windows/amd64",
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

// Prep for release
//
// - Copy each binary to `npm` directories
//
// - Zip up binaries, then copy zips to `dist` directory
func (Release) Prep() error {
	log := NewLogger()
	defer log.End()

	releaseVersion := GetEnv("RELEASE_VERSION", version.Version)
	log.Info("release version: ", releaseVersion)

	releaseArchs := []string{
		"darwin_amd64",
		"darwin_arm64",
		"freebsd_amd64",
		"freebsd_arm",
		"linux_amd64",
		"linux_arm64",
		"netbsd_amd64",
		"netbsd_arm",
		"openbsd_amd64",
		"windows_amd64"}

	releaseArchsNpmDirectories := []string{
		"reactenv-darwin-x64",
		"reactenv-darwin-arm64",
		"",
		"",
		"reactenv-linux-x64",
		"reactenv-linux-arm64",
		"",
		"",
		"",
		"reactenv-win32-x64"}

	// Zip
	for archIndex, arch := range releaseArchs {
		log.Info(fmt.Sprintf("%13s", arch))
		archLog := fmt.Sprintf("%13s |", arch)
		binDirPath := fmt.Sprintf("bin/%s", arch)
		binFilePath := ""

		// Search each binary path for the release binary,
		// ensure binary file exists and check it's file size.
		log.Debug(archLog, "checking release binary")
		binPathfiles, err := os.ReadDir(binDirPath)
		if err != nil {
			return log.Error(archLog, "error reading directory:", err)
		}
		for _, file := range binPathfiles {
			if file.IsDir() {
				continue
			}
			if strings.Contains(file.Name(), "reactenv") {
				fileInfo, err := file.Info()
				if err != nil {
					return log.Error(archLog, "error getting file info:", err)
				}
				if fileInfo.Size() < 1000000 {
					return log.Error(archLog, "release binary is too small:", err)
				}
				binFilePath = fmt.Sprintf("%s/%s", binDirPath, file.Name())
				break
			}
		}
		if binFilePath == "" {
			return log.Error(archLog, "failed to find release binary", arch)
		}

		log.Debug(archLog, "zip for release")
		zipFileName := fmt.Sprintf("reactenv_%s_%s.zip", releaseVersion, arch)
		zipFilePath := fmt.Sprintf("bin/%s", zipFileName)

		err = ZipFiles(zipFilePath, binFilePath)
		if err != nil {
			return log.Error(archLog, "failed to zip binary", err)
		}

		// Copy binary into npm directory
		if releaseArchsNpmDirectories[archIndex] != "" {
			log.Debug(archLog, "copy binary into npm directory")
			CopyFile(binFilePath, path.Join("npm", releaseArchsNpmDirectories[archIndex], path.Base(binFilePath)))
		}
	}

	// Copy zips to dist directory
	log.Debug("copying release zips to dist directory")
	if err := os.MkdirAll("dist", 0755); err != nil {
		return log.Error("failed to create dist directory:", err)
	}
	for _, arch := range releaseArchs {
		zipFilePath := fmt.Sprintf("bin/%s_%s_%s.zip", BINARY_FILENAME, releaseVersion, arch)
		CopyFile(zipFilePath, path.Join("dist", path.Base(zipFilePath)))
	}

	return nil
}

// Upload all files in a directory to an FTP server (configurable). Used to upload
// all release zips in `dist`.
//
// A new directory is created on the FTP server for the release version,
// and the release files (everything in `LOCAL_PATH` directory) are uploaded to it.
func (Release) FTP() error {
	log := NewLogger()
	defer log.End()

	ftpHost := os.Getenv("FTP_HOST")
	ftpUsername := os.Getenv("FTP_USERNAME")
	ftpPassword := os.Getenv("FTP_PASSWORD")
	ftpPath := os.Getenv("FTP_PATH")
	localPath := os.Getenv("LOCAL_PATH")
	releaseVersion := os.Getenv("RELEASE_VERSION")
	ftpReleasePath := path.Join(ftpPath, releaseVersion)

	log.Info("ftp upload path: ", ftpReleasePath)

	if ftpHost == "" || ftpUsername == "" || ftpPassword == "" || ftpPath == "" || localPath == "" || releaseVersion == "" {
		return log.Error("required FTP environment variables not set")
	}

	err := FTPUploadDir(ftpHost, ftpUsername, ftpPassword, ftpReleasePath, localPath, releaseVersion)
	if err != nil {
		return log.Error("failed to upload release files to FTP server:", err)
	}

	return nil
}

// Runs `npm publish` for each npm package.
func (Release) NPM() error {
	log := NewLogger()
	defer log.End()

	// Read all npm directories
	npmDirectories, err := fs.ReadDir(os.DirFS("npm"), ".")
	if err != nil {
		return log.Error("failed to read npm directories:", err)
	}

	// Run `npm publish` for each npm package
	for _, dir := range npmDirectories {
		if !dir.IsDir() || !strings.HasPrefix(dir.Name(), "reactenv") {
			continue
		}

		packagePath := path.Join("npm", dir.Name())
		log.Info("publishing npm package: ", packagePath)

		if err := RunStream([]string{"npm", "publish"}, packagePath, false); err != nil {
			return log.Error("failed to publish npm package:", packagePath, err)
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

	filesWithVersion := []string{
		"version/version_base.go",
		"npm/reactenv/package.json",
		"npm/reactenv-darwin-arm64/package.json",
		"npm/reactenv-darwin-x64/package.json",
		"npm/reactenv-linux-arm64/package.json",
		"npm/reactenv-linux-x64/package.json",
		"npm/reactenv-win32-x64/package.json",
	}

	for index, versionFile := range filesWithVersion {
		versionFileContent, err := os.ReadFile(versionFile)

		if err != nil {
			return log.Error("failed open version file:", err)
		}

		// Change regex based on file type
		var versionMatchRegex, versionReplaceRegex, versionReplaceSprintf string
		switch path.Ext(versionFile) {
		case ".go":
			versionMatchRegex = `= "(\d+).(\d+).(\d+)"`
			versionReplaceRegex = `= "\d+.\d+.\d+"`
			versionReplaceSprintf = `= "%s.%s.%d"`
		case ".json":
			versionMatchRegex = `"version": "(\d+).(\d+).(\d+)"`
			versionReplaceRegex = `"version": "\d+.\d+.\d+"`
			versionReplaceSprintf = `"version": "%s.%s.%d"`
		}

		// Extract current patch version
		versionMatch := regexp.MustCompile(versionMatchRegex).FindStringSubmatch(string(versionFileContent))

		if len(versionMatch) < 4 {
			return log.Error("failed to parse version:", versionFile, versionMatch)
		}

		majorVersion := versionMatch[1]
		minorVersion := versionMatch[2]
		patchVersionCurrent := versionMatch[3]
		patchVersion := commitCount

		versionCurrent := fmt.Sprintf("%s.%s.%s", majorVersion, minorVersion, patchVersionCurrent)
		versionNew := fmt.Sprintf("%s.%s.%d", majorVersion, minorVersion, patchVersion)

		if index == 0 {
			log.Info("bumping version", versionCurrent, "->", versionNew)
		}

		// Update version file
		versionFileContent = regexp.MustCompile(versionReplaceRegex).ReplaceAll(versionFileContent, []byte(fmt.Sprintf(versionReplaceSprintf, majorVersion, minorVersion, patchVersion)))

		if err := os.WriteFile(versionFile, versionFileContent, 0644); err != nil {
			return log.Error("failed to write version file:", err)
		}
	}

	return nil
}
