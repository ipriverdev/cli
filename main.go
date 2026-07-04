package main

import (
	"context"
	"os"
	"os/signal"

	"github.com/ipriverdev/cli/cmd"
	"github.com/ipriverdev/cli/internal/ui"
)

var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

func main() {
	cmd.SetVersion(version, commit, date)

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	if err := cmd.Execute(ctx); err != nil {
		ui.Error(err.Error())
		os.Exit(1) //nolint:gocritic // intentional exit after error
	}
}
