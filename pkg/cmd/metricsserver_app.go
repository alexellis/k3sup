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
		Example:      `  k3sup app install metrics-server --namespace kube-system`,
		SilenceUsage: true,
	}

	metricsServer.Flags().StringP("namespace", "n", "kube-system", "The namespace used for installation")

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

		if namespace != "kube-system" {
			return fmt.Errorf(`to override the "kube-system", install via tiller`)
		}

		clientArch, clientOS := getClientArch()

		fmt.Printf("Client: %s, %s\n", clientArch, clientOS)
		log.Printf("User dir established as: %s\n", userPath)

		os.Setenv("HELM_HOME", path.Join(userPath, ".helm"))

		_, err = tryDownloadHelm(userPath, clientArch, clientOS)
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

		err = templateChart(chartPath,
			"metrics-server",
			namespace,
			outputPath,
			"values.yaml",
			overrides)

		if err != nil {
			return err
		}

		applyRes, applyErr := kubectlTask("apply", "-R", "-f", outputPath)
		if applyErr != nil {
			return applyErr
		}

		if applyRes.ExitCode > 0 {
			return fmt.Errorf("Error applying templated YAML files, error: %s", applyRes.Stderr)
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

` + thanksForUsing)

		return nil
	}

	return metricsServer
}
