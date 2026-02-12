package command

import (
	"os"

	"github.com/hmerritt/reactenv/ui"
	"github.com/hmerritt/reactenv/version"
	"github.com/spf13/cobra"
)

var Ui = ui.GetUi()

func Run() {
	rootCmd := NewRootCommand()
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func NewRootCommand() *cobra.Command {
	showVersion := false

	// Setup root CLI
	rootCmd := &cobra.Command{
		Use:   "reactenv",
		Short: "Inject environment variables into a built react app",
		Run: func(cmd *cobra.Command, args []string) {
			if showVersion {
				Ui.Output(version.GetVersion().VersionNumber())
				return
			}
			_ = cmd.Help()
		},
		SilenceUsage: true,
	}

	// Flags
	rootCmd.Flags().BoolVar(&showVersion, "version", false, "Show version")

	// Commands
	rootCmd.AddCommand(NewCommandRun())

	return rootCmd
}
