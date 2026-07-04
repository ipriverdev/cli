package cmd

import (
	"context"
	"fmt"

	"github.com/ipriverdev/cli/internal/auth"
	"github.com/ipriverdev/cli/internal/config"
	"github.com/ipriverdev/cli/internal/ui"
	"github.com/spf13/cobra"
)

func whoamiCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "whoami",
		Short: "Display the current authenticated account",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runWhoami(cmd.Context())
		},
	}
}

func runWhoami(ctx context.Context) error {
	if _, err := auth.GetToken(); err != nil {
		ui.Info("You are not logged in.")
		ui.Info("Run `ipriver login` to authenticate.")
		return nil
	}

	cfg, err := config.Load()
	if err != nil {
		return err
	}

	username := cfg.Username

	user, err := auth.CurrentUser(ctx, config.APIHost())
	if err != nil {
		ui.Warn(fmt.Sprintf("Could not fetch account: %s", err))
	} else if user != nil {
		username = auth.DisplayName(user)
		if username != cfg.Username {
			cfg.Username = username
			_ = config.Save(cfg)
		}
	}

	if username == "" {
		username = "unknown"
	}

	ui.Success(fmt.Sprintf("Logged in to %s", config.APIHost()))
	ui.Info("Account: " + ui.Bold(username))
	return nil
}
