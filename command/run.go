package command

import (
	"fmt"
	"os"
	"path"
	"regexp"
	"strings"

	"github.com/hmerritt/reactenv/reactenv"
	"github.com/hmerritt/reactenv/ui"
)

type RunCommand struct {
	*BaseCommand
}

func (c *RunCommand) Synopsis() string {
	return "Inject environment variables into a built react app"
}

func (c *RunCommand) Help() string {
	jsInfo := c.UI.Colorize(".js", c.UI.InfoColor)
	helpText := fmt.Sprintf(`
Usage: reactenv run [options] PATH
  
Inject environment variables into a built react app.

Example:
  $ reactenv run ./dist/assets

    dist/assets
    ├── index.css
    ├── index-csxw0qbp%s
    ├── login.lazy-b839zm%s
    └── user.lazy-c7942lh%s  <- Runs on all %s files in PATH
`, jsInfo, jsInfo, jsInfo, jsInfo)

	return strings.TrimSpace(helpText)
}

func (c *RunCommand) Flags() *FlagMap {
	return GetFlagMap(FlagNamesGlobal)
}

func (c *RunCommand) Run(args []string) int {
	duration := ui.InitDuration(c.UI)

	args = c.Flags().Parse(c.UI, args)

	if len(args) == 0 {
		c.UI.Error("No asset PATH entered.")
		c.exitWithHelp()
	}

	pathToAssets := args[0]

	if _, err := os.Stat(pathToAssets); os.IsNotExist(err) {
		c.UI.Error(fmt.Sprintf("File PATH '%s' does not exist.", pathToAssets))
		c.exitWithHelp()
	}

	// @TODO: Add flag to specify matcher
	fileMatchExpression := `.*\.js`
	_, err := regexp.Compile(fileMatchExpression)

	if err != nil {
		c.UI.Error(fmt.Sprintf("File match expression '%s' is not valid.\n", fileMatchExpression))
		c.UI.Error(fmt.Sprintf("%v", err))
		c.exitWithHelp()
	}

	renv := reactenv.NewReactenv(c.UI)

	err = renv.FindFiles(pathToAssets, fileMatchExpression)

	if err != nil {
		c.UI.Error(fmt.Sprintf("Error reading files in PATH '%s'.\n", pathToAssets))
		c.UI.Error(fmt.Sprintf("%v", err))
		os.Exit(1)
	}

	if len(renv.Files) == 0 {
		c.UI.Error(fmt.Sprintf("No files found in path '%s' using matcher '%s'", pathToAssets, fileMatchExpression))
		os.Exit(1)
	}

	renv.FindOccurrences()

	if renv.OccurrencesTotal == 0 {
		c.UI.Warn(ui.WrapAtLength(fmt.Sprintf("No reactenv environment variables were found in any of the %d '%s' files within '%s', therefore nothing was injected.\n", len(renv.Files), fileMatchExpression, pathToAssets), 0))
		c.UI.Warn(ui.WrapAtLength("Possible causes:", 4))
		c.UI.Warn(ui.WrapAtLength("  - reactenv has already ran on these files", 4))
		c.UI.Warn(ui.WrapAtLength("  - Environment variables were not replaced with `__reactenv.<name>` during build", 4))
		c.UI.Warn("")
		duration.In(c.UI.WarnColor, "")
		return 1
	}

	c.UI.Output(
		fmt.Sprintf(
			"Found %d environment %s in %d/%d matching files:",
			renv.OccurrencesTotal,
			ui.Pluralize("variable", renv.OccurrencesTotal),
			len(renv.Files),
			renv.FilesMatchTotal,
		),
	)

	//
	//
	//

	allOccurrences := 0

	// @TODO:
	// - Loop all JS files, find all occurrences
	// - Log: Total occurrences, Each file occurrences, All found occurrences + the ENV value to be injected
	// - Re-loop all JS files, replace all occurrences
	// - Log errors, warnings and successes

	// Inject environment variables into .js files
	for _, file := range renv.Files {
		// Read .js file
		filePath := path.Join(pathToAssets, (*file).Name())
		fileContents, err := os.ReadFile(filePath)
		fileContentsNew := make([]byte, 0, len(fileContents))

		if err != nil {
			c.UI.Error(fmt.Sprintf("Error when reading file '%s'.\n", (*file).Name()))
			c.UI.Error(fmt.Sprintf("%v", err))
			os.Exit(1)
		}

		// Find every occurrence of `reactenv.`
		occurrences := regexp.MustCompile(`(__reactenv\.[a-zA-Z_$][0-9a-zA-Z_$]*)`).FindAllStringIndex(string(fileContents), -1)
		occurrenceReplacementValues := make([]string, len(occurrences))

		if len(occurrences) == 0 {
			continue
		}

		allOccurrences += len(occurrences)

		// For each occurrence, find the corresponding environment variable,
		// exits if any environment variable is not set.
		for index, occurrence := range occurrences {
			occurrenceText := string(fileContents[occurrence[0]:occurrence[1]])
			envName := strings.Replace(occurrenceText, "__reactenv.", "", 1)
			envValue, envExists := os.LookupEnv(envName)

			if !envExists {
				c.UI.Error(fmt.Sprintf("Environment variable not set: %s", envName))
				os.Exit(1)
			}

			occurrenceReplacementValues[index] = envValue
		}

		// Run replacement of all occurrences
		lastIndex := 0
		for index, occurrence := range occurrences {
			envValue := occurrenceReplacementValues[index]
			start, end := occurrence[0], occurrence[1]

			fileContentsNew = append(fileContentsNew, fileContents[lastIndex:start]...)
			fileContentsNew = append(fileContentsNew, envValue...)
			lastIndex = end
		}
		fileContentsNew = append(fileContentsNew, fileContents[lastIndex:]...)

		// Write .js file
		// if err := os.WriteFile(filePath, fileContentsNew, 0644); err != nil {
		// 	c.UI.Error(fmt.Sprintf("Error when writing to file '%s'.\n", filePath))
		// 	c.UI.Error(fmt.Sprintf("%v", err))
		// 	os.Exit(1)
		// }
	}

	duration.In(c.UI.SuccessColor, fmt.Sprintf("Injected '%d' environment variables", allOccurrences))
	return 0
}

func (c *RunCommand) exitWithHelp() {
	c.UI.Output("\nSee 'reactenv run --help'.")
	os.Exit(1)
}
