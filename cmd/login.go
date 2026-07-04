package cmd

import (
	"context"
	"fmt"

	"github.com/ipriverdev/cli/internal/auth"
	"github.com/ipriverdev/cli/internal/config"
	"github.com/ipriverdev/cli/internal/ui"
	"github.com/spf13/cobra"
)

func loginCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "login",
		Short: "Log in to IP River",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runLogin(cmd.Context())
		},
	}
}

func runLogin(ctx context.Context) error {
	cfg, err := config.Load()
	if err != nil {
		return err
	}

	flow := auth.NewDeviceFlow(config.Host())

	result, err := flow.Login(ctx)
	if err != nil {
		return err
	}

	if result.User != nil {
		cfg.Username = auth.DisplayName(result.User)
	}
	if err := config.Save(cfg); err != nil {
		return fmt.Errorf("save config: %w", err)
	}

	ui.Success("Authentication complete.")
	if cfg.Username != "" {
		ui.Info("Logged in as " + ui.Bold(cfg.Username))
	}

	return nil
}
