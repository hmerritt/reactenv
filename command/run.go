package command

import (
	"fmt"
	"os"
	"regexp"
	"strings"

	"github.com/hmerritt/reactenv/reactenv"
	"github.com/hmerritt/reactenv/ui"
	"github.com/spf13/cobra"
)

type RunCommand struct{}

func (c *RunCommand) Synopsis() string {
	return "Inject environment variables into a built react app"
}

func (c *RunCommand) Help() string {
	jsInfo := Ui.Colorize(".js", Ui.InfoColor)
	helpText := fmt.Sprintf(`
Usage: reactenv run [options] PATH

Inject environment variables into a built react app.

Example:
  $ reactenv run ./dist

    dist/
    ├── login/
    │   ├── login.css
    │   └── login.lazy-b839zm%s
    ├── user/
    │   ├── user.css
    │   └── user.lazy-c7942lh%s <- Runs on all %s files in PATH (recursively)
    ├── index.html
    ├── index-csxw0qbp%s
    ├── robots.txt
    └── sitemap.xml
`, jsInfo, jsInfo, jsInfo, jsInfo)

	return strings.TrimSpace(helpText)
}

func NewCommandRun() *cobra.Command {
	run := &RunCommand{}

	cmd := &cobra.Command{
		Use:   "run PATH",
		Short: run.Synopsis(),
		Args:  cobra.ArbitraryArgs,
		Run: func(cmd *cobra.Command, args []string) {
			run.Run(args)
		},
	}

	cmd.SetHelpFunc(func(cmd *cobra.Command, args []string) {
		Ui.Output(run.Help())
	})

	return cmd
}

func (c *RunCommand) Run(args []string) int {
	duration := ui.InitDuration(Ui)

	if len(args) == 0 {
		Ui.Error("No asset PATH entered.")
		c.exitWithHelp()
	}

	pathToAssets := args[0]

	if _, err := os.Stat(pathToAssets); os.IsNotExist(err) {
		Ui.Error(fmt.Sprintf("File PATH '%s' does not exist.", pathToAssets))
		c.exitWithHelp()
	}

	// @TODO: Add flag to specify matcher
	fileMatchExpression := `.*\.js$`
	_, err := regexp.Compile(fileMatchExpression)

	if err != nil {
		Ui.Error(fmt.Sprintf("File match expression '%s' is not valid.\n", fileMatchExpression))
		Ui.Error(fmt.Sprintf("%v", err))
		c.exitWithHelp()
	}

	renv := reactenv.NewReactenv(Ui)

	err = renv.FindFiles(pathToAssets, fileMatchExpression)

	if err != nil {
		Ui.Error(fmt.Sprintf("Error reading files in PATH '%s'.\n", pathToAssets))
		Ui.Error(fmt.Sprintf("%v", err))
		os.Exit(1)
	}

	if len(renv.Files) == 0 {
		Ui.Error(fmt.Sprintf("No files found in path '%s' using matcher '%s'", pathToAssets, fileMatchExpression))
		os.Exit(1)
	}

	err = renv.FindOccurrences()

	if err != nil {
		Ui.Error(fmt.Sprintf("There was an error while searching for __reactenv variables in the %d '%s' files within '%s', therefore nothing was injected.\n", renv.FilesMatchTotal, fileMatchExpression, pathToAssets))
		Ui.Error(fmt.Sprintf("%v", err))
		os.Exit(1)
	}

	if renv.OccurrencesTotal == 0 {
		Ui.Warn(ui.WrapAtLength(fmt.Sprintf("No reactenv environment variables were found in any of the %d '%s' files within '%s', therefore nothing was injected.\n", renv.FilesMatchTotal, fileMatchExpression, pathToAssets), 0))
		Ui.Warn(ui.WrapAtLength("Possible causes:", 4))
		Ui.Warn(ui.WrapAtLength("  - reactenv has already ran on these files", 4))
		Ui.Warn(ui.WrapAtLength("  - Environment variables were not replaced with `__reactenv.<name>` during build", 4))
		Ui.Warn("")
		duration.In(Ui.WarnColor, "")
		return 1
	}

	Ui.Output(
		fmt.Sprintf(
			"Found %d reactenv environment %s in %d/%d matching files:",
			renv.OccurrencesTotal,
			ui.Pluralize("variable", renv.OccurrencesTotal),
			len(renv.Files),
			renv.FilesMatchTotal,
		),
	)
	for fileIndex, fileOccurrencesTotal := range renv.OccurrencesByFile {
		Ui.Output(
			fmt.Sprintf(
				"  - %4dx in %s",
				len(fileOccurrencesTotal.Occurrences),
				renv.FileRelPaths[fileIndex],
			),
		)
	}
	Ui.Output("")

	Ui.Output(fmt.Sprintf("Environment %s checklist (ticked if value has been set):", ui.Pluralize("variable", renv.OccurrencesTotal)))
	envValuesMissing := 0
	for occurrenceKey := range renv.OccurrenceKeys {
		check := "✅"
		if _, ok := renv.OccurrenceKeysReplacement[occurrenceKey]; !ok {
			check = "❌"
			envValuesMissing++
		}
		Ui.Output(fmt.Sprintf("  - %4s %s", check, occurrenceKey))
	}
	Ui.Output("")

	if envValuesMissing > 0 {
		Ui.Error(fmt.Sprintf("Environment %s not set. See above checklist for missing values.", ui.Pluralize("variable", envValuesMissing)))
		os.Exit(1)
	}

	renv.ReplaceOccurrences()

	duration.In(Ui.SuccessColor, fmt.Sprintf("Injected all environment variables"))
	return 0
}

func (c *RunCommand) exitWithHelp() {
	Ui.Output("\nSee 'reactenv run --help'.")
	os.Exit(1)
}
