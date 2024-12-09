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

		// 1. Find all `reactenv.` occurrences
		// 2. For each occurence, find the corresponding ENV variable
		// 3. If any are missing, exit with error
		// 4. Run replacement of all occurrences.

		// Read .js file
		fileContents, err := os.ReadFile(path.Join(pathToAssets, file.Name()))

		if err != nil {
			c.UI.Error("Failed to read .js file.")
			os.Exit(1)
		}

		// Find every occurrence of `reactenv.`
		occurrences := regexp.MustCompile(`(reactenv\.[a-zA-Z_$][0-9a-zA-Z_$]*)`).FindAllString(string(fileContents), -1)
		occurrenceValues := make([]string, len(occurrences))

		if len(occurrences) == 0 {
			continue
		}

		// For each occurrence, find the corresponding environment variable,
		// exits if any environment variable is not set.
		for index, occurrence := range occurrences {
			envName := strings.Replace(occurrence, "reactenv.", "", 1)
			envValue, envExists := os.LookupEnv(envName)

			if !envExists {
				c.UI.Error(fmt.Sprintf("Environment variable not set: %s", envName))
				os.Exit(1)
			}

			occurrenceValues[index] = envValue
		}

		// Run replacement of all occurrences
		for index, occurrence := range occurrences {
			envValue := occurrenceValues[index]
			fileContents = regexp.MustCompile(fmt.Sprintf("\"%s\"", occurrence)).ReplaceAll(fileContents, []byte(fmt.Sprintf("\"%s\"", envValue)))
		}

		// Write .js file
		if err := os.WriteFile(path.Join(pathToAssets, file.Name()), fileContents, 0644); err != nil {
			c.UI.Error("Failed to write .js file.")
			os.Exit(1)
		}
	}

	duration.In(c.UI.SuccessColor, "Injected environment variables")
	return 0
}
