package cmd

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
)

func MakeApps() *cobra.Command {
	var command = &cobra.Command{
		Use:          "app",
		Short:        "Manage Kubernetes apps",
		Long:         `Manage Kubernetes apps`,
		Example:      `  k3sup app install`,
		SilenceUsage: false,
	}

	var install = &cobra.Command{
		Use:   "install",
		Short: "Install a Kubernetes app",
		Example: `  k3sup app install [APP]
  k3sup app install openfaas --help
  k3sup app install inlets-operator --token-file $HOME/do
  k3sup app install --help`,
		SilenceUsage: true,
	}

	install.PersistentFlags().String("kubeconfig", "kubeconfig", "Local path for your kubeconfig file")

	install.RunE = func(command *cobra.Command, args []string) error {

		if len(args) == 0 {
			fmt.Printf("You can install: %s\n", strings.TrimRight(strings.Join(getApps(), ", "), ", "))
			return nil
		}

		return nil
	}

	command.AddCommand(install)
	install.AddCommand(makeInstallOpenFaaS())
	install.AddCommand(makeInstallMetricsServer())
	install.AddCommand(makeInstallInletsOperator())
	install.AddCommand(makeInstallCertManager())
	install.AddCommand(makeInstallOpenFaaSIngress())
	install.AddCommand(makeInstallNginx())
	install.AddCommand(makeInstallChart())
	install.AddCommand(makeInstallTiller())
	install.AddCommand(makeInstallLinkerd())
	install.AddCommand(makeInstallCronConnector())
	install.AddCommand(makeInstallKafkaConnector())

	return command
}

func getApps() []string {
	return []string{"openfaas", "nginx-ingress", "cert-manager",
		"openfaas-ingress", "inlets-operator", "metrics-server",
		"chart", "tiller", "linkerd", "cron-connector", "kafka-connector"}
}
