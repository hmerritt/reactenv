package command

import (
	"github.com/hmerritt/reactenv/ui"
	"github.com/hmerritt/reactenv/version"
	"github.com/spf13/cobra"
)

func NewRootCommand() *cobra.Command {
	uiInstance := ui.GetUi()
	showVersion := false

	// Setup root CLI
	rootCmd := &cobra.Command{
		Use:   "reactenv",
		Short: "Inject environment variables into a built react app",
		Run: func(cmd *cobra.Command, args []string) {
			if showVersion {
				uiInstance.Output(version.GetVersion().VersionNumber())
				return
			}
			_ = cmd.Help()
		},
		SilenceUsage: true,
	}

	// Completion
	rootCmd.CompletionOptions.DisableDefaultCmd = true
	rootCmd.SetHelpCommand(&cobra.Command{Use: "help", Hidden: true})

	// Flags
	rootCmd.Flags().BoolVar(&showVersion, "version", false, "Show version")

	// Commands
	rootCmd.AddCommand(NewRunCommand(uiInstance))

	return rootCmd
}
