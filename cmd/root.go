package cmd

import (
	"context"
	"fmt"
	"os"

	"github.com/ipriverdev/cli/internal/app"
	"github.com/spf13/cobra"
)

var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

func SetVersion(v, c, d string) {
	version = v
	commit = c
	date = d
}

const banner = `
  ___ ____    ____  _                
 |_ _|  _ \  |  _ \(_)_   _____ _ __ 
  | || |_) | | |_) | \ \ / / _ \ '__|
  | ||  __/  |  _ <| |\ V /  __/ |   
 |___|_|     |_| \_\_| \_/ \___|_|   
`

var rootCmd = &cobra.Command{
	Use:           "ipriver",
	Short:         "The IP River command-line tool",
	SilenceErrors: true,
	SilenceUsage:  true,
	Long: banner + `
Welcome to the IP River Portal CLI!

Use ` + "`ipriver -v`" + ` to display the current version.
Here are the base commands:`,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		if noColor, _ := cmd.Flags().GetBool("no-color"); noColor {
			_ = os.Setenv("NO_COLOR", "1")
		}
		if app.OutputFormat != "" && app.OutputFormat != "json" {
			return fmt.Errorf("unknown output format %q (supported: json)", app.OutputFormat)
		}
		return nil
	},
}

func Execute(ctx context.Context) error {
	return rootCmd.ExecuteContext(ctx)
}

func init() {
	rootCmd.PersistentFlags().StringVar(&app.OutputFormat, "format", "", "Output format (json)")
	rootCmd.PersistentFlags().Bool("no-color", false, "Disable color output")
	rootCmd.PersistentFlags().BoolP("version", "v", false, "Print version information")

	rootCmd.InitDefaultHelpFlag()
	if f := rootCmd.Flags().Lookup("help"); f != nil {
		f.Usage = "Help for IP River"
	}

	rootCmd.RunE = func(cmd *cobra.Command, args []string) error {
		if v, _ := cmd.Flags().GetBool("version"); v {
			fmt.Println(VersionInfo())
			return nil
		}
		return cmd.Help()
	}

	rootCmd.AddCommand(loginCmd())
	rootCmd.AddCommand(logoutCmd())
	rootCmd.AddCommand(whoamiCmd())
	rootCmd.AddCommand(addressCmd())
	rootCmd.AddCommand(checkCmd())
	rootCmd.AddCommand(servicesCmd())
	rootCmd.AddCommand(ordersCmd())
	rootCmd.AddCommand(ticketsCmd())
	rootCmd.AddCommand(catalogueCmd())
	rootCmd.AddCommand(completionCmd())

	rootCmd.CompletionOptions.DisableDefaultCmd = true
	rootCmd.SetHelpCommand(&cobra.Command{Hidden: true})
}

func VersionInfo() string {
	return fmt.Sprintf("%s (commit: %s, built: %s)", version, commit, date)
}
