package cmd

import (
	"github.com/ipriverdev/cli/internal/auth"
	"github.com/ipriverdev/cli/internal/config"
	"github.com/ipriverdev/cli/internal/ui"
	"github.com/spf13/cobra"
)

func logoutCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "logout",
		Short: "Log out of IP River",
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := auth.DeleteCredentials(); err != nil {
				return err
			}

			cfg, err := config.Load()
			if err == nil {
				cfg.Username = ""
				_ = config.Save(cfg)
			}

			ui.Success("Logged out.")
			return nil
		},
	}
}
