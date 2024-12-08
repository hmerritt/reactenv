package command

import (
	"fmt"
	"io/fs"
	"os"
	"path"
	"regexp"
	"strings"

	"github.com/hmerritt/reactenv/ui"
)

type RunCommand struct {
	*BaseCommand
}

func (c *RunCommand) Synopsis() string {
	return "Inject environment variables into a built react app"
}

func (c *RunCommand) Help() string {
	helpText := `
Usage: reactenv run [options] FILE
  
  Inject environment variables into a built react app.
`

	return strings.TrimSpace(helpText)
}

func (c *RunCommand) Flags() *FlagMap {
	return GetFlagMap(FlagNamesGlobal)
}

func (c *RunCommand) strictExit() {
	if c.Flags().Get("strict").Value == true {
		c.UI.Error("\nAn error occured while using the '--strict' flag.")
		os.Exit(1)
	}
}

func (c *RunCommand) Run(args []string) int {
	duration := ui.InitDuration(c.UI)

	args = c.Flags().Parse(c.UI, args)

	if len(args) == 0 {
		c.UI.Error("No asset path entered.")
		os.Exit(1)
	}

	pathToAssets := args[0]

	if _, err := os.Stat(pathToAssets); os.IsNotExist(err) {
		c.UI.Error("Asset path does not exist.")
		os.Exit(1)
	}

	// Find all .js files
	assetFiles, err := os.ReadDir(pathToAssets)
	jsFiles := make([]fs.DirEntry, 0, len(assetFiles))

	if err != nil {
		c.UI.Error("Failed to read asset directory files.")
		os.Exit(1)
	}

	for _, file := range assetFiles {
		if file.IsDir() || !strings.HasSuffix(file.Name(), ".js") {
			continue
		}
		jsFiles = append(jsFiles, file)
	}

	if len(jsFiles) == 0 {
		c.UI.Error("No .js files found.")
		os.Exit(1)
	}

	// Inject environment variables into .js files
	for _, file := range jsFiles {
		fmt.Println(file.Name())

		// ENV value
		envName := "reactenv.NODE_ENV_TEST"
		envValue := os.Getenv("ANDROID_HOME")

		if envValue == "" {
			c.UI.Error("Environment variable not found.")
			os.Exit(1)
		}

		// Read .js file
		assetFile, err := os.ReadFile(path.Join(pathToAssets, file.Name()))

		if err != nil {
			c.UI.Error("Failed to read .js file.")
			os.Exit(1)
		}

		// Inject environment variable into .js file
		assetFile = regexp.MustCompile(envName).ReplaceAll(assetFile, []byte(fmt.Sprintf(envValue)))

		// Write .js file
		if err := os.WriteFile(path.Join(pathToAssets, file.Name()), assetFile, 0644); err != nil {
			c.UI.Error("Failed to write .js file.")
			os.Exit(1)
		}
	}

	duration.In(c.UI.SuccessColor, "Injected environment variables")
	return 0
}
