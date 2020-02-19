package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// MakeUpdate returns the update sub command of k3sup
func MakeUpdate() *cobra.Command {
	var command = &cobra.Command{
		Use:          "update",
		Short:        "Print update instructions",
		Example:      `  k3sup update`,
		SilenceUsage: false,
	}
	command.Run = func(cmd *cobra.Command, args []string) {
		fmt.Println(k3supUpdate)
	}
	return command
}

const k3supUpdate = `You can update k3sup with the following:

# For Linux/MacOS:
curl -SLfs https://get.k3sup.dev | sudo sh

# For Windows (using Git Bash)
curl -SLfs https://get.k3sup.dev | sh

# Or download from GitHub: https://github.com/alexellis/k3sup/releases

Thanks for using k3sup!`
