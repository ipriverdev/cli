package cmd

import (
	"os"

	"github.com/spf13/cobra"
)

func completionCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "completion [bash|zsh|fish|powershell]",
		Short: "Generate shell completion scripts",
		Long: `Generate shell completion scripts for ipriver.

To load completions:

  bash:
    source <(ipriver completion bash)
    # Or persist:
    ipriver completion bash > /etc/bash_completion.d/ipriver

  zsh:
    ipriver completion zsh > "${fpath[1]}/_ipriver"
    # Then restart your shell or run: compinit

  fish:
    ipriver completion fish | source
    # Or persist:
    ipriver completion fish > ~/.config/fish/completions/ipriver.fish

  powershell:
    ipriver completion powershell | Out-String | Invoke-Expression
`,
		DisableFlagsInUseLine: true,
		ValidArgs:             []string{"bash", "zsh", "fish", "powershell"},
		Args:                  cobra.MatchAll(cobra.ExactArgs(1), cobra.OnlyValidArgs),
		RunE: func(cmd *cobra.Command, args []string) error {
			switch args[0] {
			case "bash":
				return rootCmd.GenBashCompletion(os.Stdout)
			case "zsh":
				return rootCmd.GenZshCompletion(os.Stdout)
			case "fish":
				return rootCmd.GenFishCompletion(os.Stdout, true)
			case "powershell":
				return rootCmd.GenPowerShellCompletionWithDesc(os.Stdout)
			}
			return nil
		},
	}
	return cmd
}
