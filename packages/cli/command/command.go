package command

import (
	"fmt"
	"os"

	"github.com/hmerritt/reactenv/version"

	"github.com/mitchellh/cli"
)

func Run() {
	// Initiate new CLI app
	app := cli.NewCLI("reactenv", version.GetVersion().VersionNumber())
	app.Args = os.Args[1:]

	// Feed active commands to CLI app
	app.Commands = map[string]cli.CommandFactory{
		"run": func() (cli.Command, error) {
			return &RunCommand{
				BaseCommand: GetBaseCommand(),
			}, nil
		},
	}

	// Run app
	exitStatus, err := app.Run()
	if err != nil {
		os.Stderr.WriteString(fmt.Sprint(err))
	}

	// Exit without an error if no arguments were passed
	if len(app.Args) == 0 {
		os.Exit(0)
	}

	os.Exit(exitStatus)
}
