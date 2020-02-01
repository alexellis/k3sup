package apps

import (
	"fmt"
	"log"
	"os"
	"path"
	"strings"

	"github.com/alexellis/k3sup/pkg"
	"github.com/alexellis/k3sup/pkg/config"
	"github.com/alexellis/k3sup/pkg/env"
	"github.com/alexellis/k3sup/pkg/helm"
	"github.com/spf13/cobra"
)

func MakeInstallChart() *cobra.Command {
	var chartCmd = &cobra.Command{
		Use:   "chart",
		Short: "Install the specified helm chart",
		Long: `Install the specified helm chart without using tiller.
Note: You may need to install a CRD or run other additional steps
before using the generic helm chart installer command.`,
		Example: `  k3sup install chart --repo-name stable/nginx-ingress \
     --set controller.service.type=NodePort
  k3sup app install chart --repo-name inlets/inlets-operator \
     --repo-url https://inlets.github.io/inlets-operator/`,
		SilenceUsage: true,
	}

	chartCmd.Flags().StringP("namespace", "n", "default", "The namespace to install the chart")
	chartCmd.Flags().String("repo", "", "The chart repo to install from")
	chartCmd.Flags().String("values-file", "", "Give the values.yaml file to use from the upstream chart repo")
	chartCmd.Flags().String("repo-name", "", "Chart name")
	chartCmd.Flags().String("repo-url", "", "Chart repo")

	chartCmd.Flags().StringArray("set", []string{}, "Set individual values in the helm chart")

	chartCmd.RunE = func(command *cobra.Command, args []string) error {
		chartRepoName, _ := command.Flags().GetString("repo-name")
		chartRepoURL, _ := command.Flags().GetString("repo-url")

		chartName := chartRepoName
		if index := strings.Index(chartRepoName, "/"); index > -1 {
			chartName = chartRepoName[index+1:]
		}

		chartPrefix := chartRepoName
		if index := strings.Index(chartRepoName, "/"); index > -1 {
			chartPrefix = chartRepoName[:index]
		}

		if len(chartRepoName) == 0 {
			return fmt.Errorf("--repo-name required")
		}

		kubeConfigPath := getDefaultKubeconfig()

		if command.Flags().Changed("kubeconfig") {
			kubeConfigPath, _ = command.Flags().GetString("kubeconfig")
		}

		fmt.Printf("Using kubeconfig: %s\n", kubeConfigPath)

		namespace, _ := command.Flags().GetString("namespace")

		userPath, err := config.InitUserDir()
		if err != nil {
			return err
		}

		clientArch, clientOS := env.GetClientArch(true)

		fmt.Printf("Client: %s, %s\n", clientArch, clientOS)

		log.Printf("User dir established as: %s\n", userPath)

		os.Setenv("HELM_HOME", path.Join(userPath, ".helm"))

		_, err = helm.TryDownloadHelm(userPath, clientArch, clientOS, false)
		if err != nil {
			return err
		}

		if len(chartRepoURL) > 0 {
			err = addHelmRepo(chartPrefix, chartRepoURL, false)
			if err != nil {
				return err
			}
		}

		err = updateHelmRepos(false)
		if err != nil {
			return err
		}

		res, kcErr := kubectlTask("get", "namespace", namespace)

		if kcErr != nil {
			return err
		}

		if res.ExitCode != 0 {
			err = kubectl("create", "namespace", namespace)
			if err != nil {
				return err
			}
		}

		chartPath := path.Join(os.TempDir(), "charts")

		err = fetchChart(chartPath, chartRepoName, false)
		if err != nil {
			return err
		}

		outputPath := path.Join(chartPath, "chart/rendered")

		setMap := map[string]string{}
		setVals, _ := chartCmd.Flags().GetStringArray("set")

		for _, setV := range setVals {
			var k string
			var v string

			if index := strings.Index(setV, "="); index > -1 {
				k = setV[:index]
				v = setV[index+1:]
				setMap[k] = v
			}
		}

		err = templateChart(chartPath, chartName, namespace, outputPath, "values.yaml", setMap)
		if err != nil {
			return err
		}

		err = kubectl("apply", "--namespace", namespace, "-R", "-f", outputPath)
		if err != nil {
			return err
		}

		fmt.Println(
			`=======================================================================
chart ` + chartRepoName + ` installed.
=======================================================================
		
` + pkg.ThanksForUsing)

		return nil
	}

	return chartCmd
}
