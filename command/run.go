package command

import (
	"fmt"
	"os"
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
		c.UI.Warn(ui.WrapAtLength(fmt.Sprintf("No reactenv environment variables were found in any of the %d '%s' files within '%s', therefore nothing was injected.\n", renv.FilesMatchTotal, fileMatchExpression, pathToAssets), 0))
		c.UI.Warn(ui.WrapAtLength("Possible causes:", 4))
		c.UI.Warn(ui.WrapAtLength("  - reactenv has already ran on these files", 4))
		c.UI.Warn(ui.WrapAtLength("  - Environment variables were not replaced with `__reactenv.<name>` during build", 4))
		c.UI.Warn("")
		duration.In(c.UI.WarnColor, "")
		return 1
	}

	c.UI.Output(
		fmt.Sprintf(
			"Found %d reactenv environment %s in %d/%d matching files:",
			renv.OccurrencesTotal,
			ui.Pluralize("variable", renv.OccurrencesTotal),
			len(renv.Files),
			renv.FilesMatchTotal,
		),
	)
	for fileIndex, fileOccurrencesTotal := range renv.OccurrencesByFile {
		c.UI.Output(
			fmt.Sprintf(
				"  - %4dx in %s",
				len(fileOccurrencesTotal.Occurrences),
				(*renv.Files[fileIndex]).Name(),
			),
		)
	}
	c.UI.Output("")

	c.UI.Output(fmt.Sprintf("Environment %s checklist (ticked if value has been set):", ui.Pluralize("variable", renv.OccurrencesTotal)))
	envValuesMissing := 0
	for occurrenceKey := range renv.OccurrenceKeys {
		check := "✅"
		if _, ok := renv.OccurrenceKeysReplacement[occurrenceKey]; !ok {
			check = "❌"
			envValuesMissing++
		}
		c.UI.Output(fmt.Sprintf("  - %4s %s", check, occurrenceKey))
	}
	c.UI.Output("")

	if envValuesMissing > 0 {
		c.UI.Error(fmt.Sprintf("Environment %s not set. See above checklist for missing values.", ui.Pluralize("variable", envValuesMissing)))
		os.Exit(1)
	}

	renv.ReplaceOccurrences()

	duration.In(c.UI.SuccessColor, fmt.Sprintf("Injected all environment variables"))
	return 0
}

func (c *RunCommand) exitWithHelp() {
	c.UI.Output("\nSee 'reactenv run --help'.")
	os.Exit(1)
}
