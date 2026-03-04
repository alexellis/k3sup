package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

func MakePro() *cobra.Command {
	var command = &cobra.Command{
		Use:   "pro",
		Short: "Learn about K3sup Pro",
		Long: `K3sup Pro is built for professionals, teams, and homelabs:

  - IaaC/GitOps workflow with plan and apply commands
  - Parallel installation across many nodes
  - Rolling upgrades and day-2 operations
  - Uninstall, exec, and get-config across your fleet
  - Pre-download K3s binaries for efficient installations across many nodes
  - Integrates directly with https://slicervm.com

Learn more at https://github.com/alexellis/k3sup#k3sup-pro`,
		SilenceUsage: true,
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println(cmd.Long)
		},
	}

	return command
}
