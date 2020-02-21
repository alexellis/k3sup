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

func MakeInstallCrossplane() *cobra.Command {
	var crossplane = &cobra.Command{
		Use:   "crossplane",
		Short: "Install Crossplane",
		Long: `Install Crossplane to deploy managed services across cloud providers and
schedule workloads to any Kubernetes cluster`,
		Example:      `  k3sup app install crossplane`,
		SilenceUsage: true,
	}

	crossplane.Flags().StringP("namespace", "n", "crossplane-system", "The namespace used for installation")
	crossplane.Flags().Bool("update-repo", true, "Update the helm repo")
	crossplane.Flags().Bool("helm3", true, "Use helm3, if set to false uses helm2")

	crossplane.RunE = func(command *cobra.Command, args []string) error {
		wait, _ := command.Flags().GetBool("wait")
		kubeConfigPath := getDefaultKubeconfig()

		if command.Flags().Changed("kubeconfig") {
			kubeConfigPath, _ = command.Flags().GetString("kubeconfig")
		}

		fmt.Printf("Using kubeconfig: %s\n", kubeConfigPath)

		helm3, _ := command.Flags().GetBool("helm3")

		if helm3 {
			fmt.Println("Using helm3")
		}
		namespace, _ := command.Flags().GetString("namespace")

		if namespace != "crossplane-system" {
			return fmt.Errorf(`to override the namespace, install crossplane via helm manually`)
		}

		arch := getNodeArchitecture()
		if !strings.Contains(arch, "64") {
			return fmt.Errorf(`crossplane is currently only supported on 64-bit architectures`)
		}
		fmt.Printf("Node architecture: %q\n", arch)

		userPath, err := config.InitUserDir()
		if err != nil {
			return err
		}

		clientArch, clientOS := env.GetClientArch()

		fmt.Printf("Client: %q, %q\n", clientArch, clientOS)

		log.Printf("User dir established as: %s\n", userPath)

		os.Setenv("HELM_HOME", path.Join(userPath, ".helm"))

		_, err = helm.TryDownloadHelm(userPath, clientArch, clientOS, helm3)
		if err != nil {
			return err
		}

		err = addHelmRepo("crossplane-alpha", "https://charts.crossplane.io/alpha", helm3)
		if err != nil {
			return err
		}

		updateRepo, _ := crossplane.Flags().GetBool("update-repo")

		if updateRepo {
			err = updateHelmRepos(helm3)
			if err != nil {
				return err
			}
		}

		chartPath := path.Join(os.TempDir(), "charts")

		err = fetchChart(chartPath, "crossplane-alpha/crossplane", defaultVersion, helm3)
		if err != nil {
			return err
		}

		if helm3 {

			outputPath := path.Join(chartPath, "crossplane")

			_, nsErr := kubectlTask("create", "namespace", "crossplane-system")
			if nsErr != nil && !strings.Contains(nsErr.Error(), "AlreadyExists") {
				return nsErr
			}

			err := helm3Upgrade(outputPath, "crossplane-alpha/crossplane",
				namespace, "values.yaml", "", map[string]string{}, wait)
			if err != nil {
				return err
			}

		} else {
			outputPath := path.Join(chartPath, "crossplane-alpha/crossplane")
			err = templateChart(chartPath, "crossplane", namespace, outputPath, "values.yaml", map[string]string{})
			if err != nil {
				return err
			}

			applyRes, applyErr := kubectlTask("apply", "-R", "-f", outputPath)
			if applyErr != nil {
				return applyErr
			}

			if applyRes.ExitCode > 0 {
				return fmt.Errorf("error applying templated YAML files, error: %s", applyRes.Stderr)
			}
		}

		fmt.Println(crossplaneInstallMsg)
		return nil
	}

	return crossplane
}

const CrossplanInfoMsg = `# Get started by installing a stack for your favorite provider:
* stack-gcp: https://crossplane.io/docs/master/install-crossplane.html#gcp-stack
* stack-aws: https://crossplane.io/docs/master/install-crossplane.html#aws-stack
* stack-azure: https://crossplane.io/docs/master/install-crossplane.html#azure-stack

Learn more about Crossplane: https://crossplaneio.github.io/docs/`

const crossplaneInstallMsg = `=======================================================================
= Crossplane has been installed.                                      =
=======================================================================` +
	"\n\n" + CrossplanInfoMsg + "\n\n" + pkg.ThanksForUsing
