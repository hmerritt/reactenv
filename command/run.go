package command

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/samber/lo"
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
	return GetFlagMap(lo.Union(FlagNamesGlobal, []string{"start", "end"}))
}

func (c *RunCommand) strictExit() {
	if c.Flags().Get("strict").Value == true {
		c.UI.Error("\nAn error occured while using the '--strict' flag.")
		os.Exit(1)
	}
}

func (c *RunCommand) Run(args []string) int {
	// Record the total duration of this command
	timeStart := time.Now()

	args = c.Flags().Parse(c.UI, args)

	var pathToAssets string

	if len(args) == 0 {
		c.UI.Error("No file entered.")
		os.Exit(1)
	} else {
		pathToAssets = args[0]
	}

	fmt.Println(pathToAssets)

	c.UI.Output(fmt.Sprintf("%s in %s", c.UI.Colorize("ICS file converted", c.UI.SuccessColor), time.Since(timeStart)))

	return 0
}
