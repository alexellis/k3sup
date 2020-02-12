package apps

import (
	"fmt"
	"log"
	"os"
	"path"

	"github.com/alexellis/k3sup/pkg"
	"github.com/alexellis/k3sup/pkg/config"
	"github.com/alexellis/k3sup/pkg/env"
	"github.com/alexellis/k3sup/pkg/helm"
	"github.com/spf13/cobra"
)

func MakeInstallMetricsServer() *cobra.Command {
	var metricsServer = &cobra.Command{
		Use:          "metrics-server",
		Short:        "Install metrics-server",
		Long:         `Install metrics-server to provide metrics on nodes and Pods in your cluster.`,
		Example:      `  k3sup app install metrics-server --namespace kube-system --helm3`,
		SilenceUsage: true,
	}

	metricsServer.Flags().StringP("namespace", "n", "kube-system", "The namespace used for installation")
	metricsServer.Flags().Bool("helm3", true, "Use helm3, if set to false uses helm2")

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

		helm3, _ := command.Flags().GetBool("helm3")

		if helm3 {
			fmt.Println("Using helm3")
		}

		clientArch, clientOS := env.GetClientArch()
		fmt.Printf("Client: %q, %q\n", clientArch, clientOS)
		log.Printf("User dir established as: %s\n", userPath)
		os.Setenv("HELM_HOME", path.Join(userPath, ".helm"))

		if helm3 {
			os.Setenv("HELM_VERSION", helm3Version)
		}

		_, err = helm.TryDownloadHelm(userPath, clientArch, clientOS, helm3)
		if err != nil {
			return err
		}

		err = updateHelmRepos(helm3)
		if err != nil {
			return err
		}

		chartPath := path.Join(os.TempDir(), "charts")
		err = fetchChart(chartPath, "stable/metrics-server", helm3)

		if err != nil {
			return err
		}

		overrides := map[string]string{}
		overrides["args"] = `{--kubelet-insecure-tls,--kubelet-preferred-address-types=InternalIP\,ExternalIP\,Hostname}`
		fmt.Println("Chart path: ", chartPath)

		wait := false

		if helm3 {
			outputPath := path.Join(chartPath, "metrics-server")

			err := helm3Upgrade(outputPath, "stable/metrics-server", namespace,
				"values.yaml",
				overrides, wait)

			if err != nil {
				return err
			}

		} else {
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

			applyRes, applyErr := kubectlTask("apply", "-n", namespace, "-R", "-f", outputPath)
			if applyErr != nil {
				return applyErr
			}
			if applyRes.ExitCode > 0 {
				return fmt.Errorf("error applying templated YAML files, error: %s", applyRes.Stderr)
			}

		}

		fmt.Println(`=======================================================================
= metrics-server has been installed.                                  =
=======================================================================

# It can take a few minutes for the metrics-server to collect data
# from the cluster. Try these commands and wait a few moments if
# no data is showing.

` + MetricsInfoMsg + `

` + pkg.ThanksForUsing)

		return nil
	}

	return metricsServer
}

const MetricsInfoMsg = `# Check pod usage

kubectl top pod

# Check node usage

kubectl top node


# Find out more at:
# https://github.com/helm/charts/tree/master/stable/metrics-server`
