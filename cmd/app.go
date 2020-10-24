package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

const getScript = "curl -sfL https://get.k3s.io"

func MakeApps() *cobra.Command {
	var command = &cobra.Command{
		Use:   "app",
		Short: "Install Kubernetes apps from helm charts or YAML files",
		Long: `Install Kubernetes apps from helm charts or YAML files using the "install" 
command. Helm 2 is used by default unless a --helm3 flag exists and is passed. 
You can also find the post-install message for each app with the "info" 
command.`,
		Example: `  k3sup app install
  k3sup app install openfaas --helm3 --gateways=2
  k3sup app info inlets-operator`,
		SilenceUsage: true,
	}

	command.RunE = func(cmd *cobra.Command, args []string) error {
		return fmt.Errorf(`
The "k3sup app install/info" command has moved to a new home.
You can now install Kubernetes apps via the arkade project.

To find out more about this decision, see the following issue:
https://github.com/alexellis/k3sup/issues/217

Example:

  k3sup app install  -> arkade install
  k3sup app info     -> arkade info

curl -sSL https://dl.get-arkade.dev/ | sudo sh

Read more about https://get-arkade.dev/`)
	}

	return command
}
