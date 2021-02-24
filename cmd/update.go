package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

func MakeUpdate() *cobra.Command {
	var command = &cobra.Command{
		Use:   "update",
		Short: "Print update instructions",
		Long: `Print instructions for updating your version of k3sup.

` + SupportMsg,
		Example:      `  k3sup update`,
		SilenceUsage: false,
	}
	command.Run = func(cmd *cobra.Command, args []string) {
		fmt.Println(k3supUpdate)
	}
	return command
}

const k3supUpdate = `You can update k3sup with the following:

# Remove cached versions of tools
rm -rf $HOME/.k3sup

# For Linux/MacOS:
curl -SLfs https://get.k3sup.dev | sudo sh

# For Windows (using Git Bash)
curl -SLfs https://get.k3sup.dev | sh

# Or download from GitHub: https://github.com/alexellis/k3sup/releases

Thanks for using k3sup!`
