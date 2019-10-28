package cmd

import (
	"fmt"
	"log"
	"os"
	"path"

	"github.com/alexellis/k3sup/pkg/config"

	"github.com/spf13/cobra"
)

func makeInstallMetricsServer() *cobra.Command {
	var metricsServer = &cobra.Command{
		Use:          "metrics-server",
		Short:        "Install metrics-server",
		Long:         `Install metrics-server`,
		Example:      `  k3sup app install metrics-server --namespace default`,
		SilenceUsage: true,
	}

	metricsServer.Flags().StringP("namespace", "n", "default", "The namespace used for installation")

	metricsServer.RunE = func(command *cobra.Command, args []string) error {
		kubeConfigPath := getDefaultKubeconfig()

		if command.Flags().Changed("kubeconfig") {
			kubeConfigPath, _ = command.Flags().GetString("kubeconfig")
		}

		fmt.Printf("Using kubeconfig: %s\n", kubeConfigPath)

		userPath, err := config.InitUserDir()
		if err != nil {
			return err
		}
		namespace, _ := command.Flags().GetString("namespace")

		if namespace != "default" {
			return fmt.Errorf(`to override the "default", install via tiller`)
		}

		clientArch, clientOS := getClientArch()

		fmt.Printf("Client: %s, %s\n", clientArch, clientOS)
		log.Printf("User dir established as: %s\n", userPath)

		os.Setenv("HELM_HOME", path.Join(userPath, ".helm"))

		err = tryDownloadHelm(userPath, clientArch, clientOS)
		if err != nil {
			return err
		}

		err = updateHelmRepos()
		if err != nil {
			return err
		}

		chartPath := path.Join(os.TempDir(), "charts")
		err = fetchChart(chartPath, "stable/metrics-server")

		if err != nil {
			return err
		}

		overrides := map[string]string{}
		overrides["args"] = `{--kubelet-insecure-tls,--kubelet-preferred-address-types=InternalIP\,ExternalIP\,Hostname}`
		fmt.Println("Chart path: ", chartPath)
		outputPath := path.Join(chartPath, "metrics-server/rendered")

		// ns:="kube-system"
		ns := "default"
		err = templateChart(chartPath,
			"metrics-server",
			ns,
			outputPath,
			"values.yaml",
			overrides)

		if err != nil {
			return err
		}

		err = kubectl("apply", "-R", "-f", outputPath)

		if err != nil {
			return err
		}

		fmt.Println(`=======================================================================
= metrics-server has been installed.                                  =
=======================================================================

# It can take a few minutes for the metrics-server to collect data
# from the cluster. Try these commands and wait a few moments if
# no data is showing.

# Check pod usage

kubectl top pod

# Check node usage

kubectl top node


# Find out more at:
# https://github.com/helm/charts/tree/master/stable/metrics-server

Thank you for using k3sup!`)

		return nil
	}

	return metricsServer
}
