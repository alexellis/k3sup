package apps

import (
	"fmt"
	"log"
	"os"
	"path"

	"github.com/alexellis/k3sup/pkg/config"
	"github.com/alexellis/k3sup/pkg/env"
	"github.com/alexellis/k3sup/pkg/helm"
	"github.com/spf13/cobra"
)

func MakeInstallHelmOperator() *cobra.Command {
	var operator = &cobra.Command{
		Use:          "helm-operator",
		Short:        "Install a the helm operator",
		Long:         `Install a the helm operator`,
		Example:      `  k3sup app install helm-operator --namespace default`,
		SilenceUsage: true,
	}

	operator.Flags().StringP("namespace", "n", "default", "The namespace used for installation")
	operator.Flags().Bool("update-repo", true, "Update the helm repo")
	operator.Flags().Bool("helm3", true, "Use helm3, if set to false uses helm2")

	operator.RunE = func(command *cobra.Command, args []string) error {
		kubeConfigPath := getDefaultKubeconfig()
		wait, _ := command.Flags().GetBool("wait")

		if command.Flags().Changed("kubeconfig") {
			kubeConfigPath, _ = command.Flags().GetString("kubeconfig")
		}

		updateRepo, _ := operator.Flags().GetBool("update-repo")

		fmt.Printf("Using kubeconfig: %s\n", kubeConfigPath)
		helm3, _ := command.Flags().GetBool("helm3")

		userPath, err := config.InitUserDir()
		if err != nil {
			return err
		}
		namespace, _ := command.Flags().GetString("namespace")
		if namespace != "default" {
			return fmt.Errorf(`to override the "default", install via tiller`)
		}

		clientArch, clientOS := env.GetClientArch()

		fmt.Printf("Client: %s, %s\n", clientArch, clientOS)
		log.Printf("User dir established as: %s\n", userPath)

		os.Setenv("HELM_HOME", path.Join(userPath, ".helm"))

		_, err = helm.TryDownloadHelm(userPath, clientArch, clientOS, helm3)
		if err != nil {
			return err
		}

		err = addHelmRepo("fluxcd", "https://charts.fluxcd.io", helm3)
		if err != nil {
			return err
		}

		if updateRepo {
			err = updateHelmRepos(helm3)
			if err != nil {
				return err
			}
		}

		chartPath := path.Join(os.TempDir(), "charts")
		err = fetchChart(chartPath, "fluxcd/helm-operator", defaultVersion, helm3)

		if err != nil {
			return err
		}

		overrides := map[string]string{}

		arch := getNodeArchitecture()

		fmt.Printf("Node architecture: %q\n", arch)
		//TODO: override images if arch is arm
		//overrides["persistence.enabled"] = "false"

		fmt.Println("Chart path: ", chartPath)

		ns := "default"

		log.Printf("Applying CRD\n")
		crdsURL := "https://raw.githubusercontent.com/jetstack/cert-manager/release-0.12/deploy/manifests/00-crds.yaml"
		res, err := kubectlTask("apply", "--validate=false", "-f",
			crdsURL)
		if err != nil {
			return err
		}

		if res.ExitCode > 0 {
			return fmt.Errorf("error applying CRD from: %s, error: %s", crdsURL, res.Stderr)
		}

		if helm3 {
			outputPath := path.Join(chartPath, "helm-operator")

			err := helm3Upgrade(outputPath, "fluxcd/helm-operator", ns,
				"values.yaml",
				defaultVersion,
				overrides,
				wait)

			if err != nil {
				return err
			}
		} else {
			outputPath := path.Join(chartPath, "helm-operator/rendered")

			err = templateChart(chartPath,
				"helm-operator",
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
		}

		fmt.Println(helmOperatorInstallMsg)
		return nil
	}

	return operator
}

const helmOperatorInstallMsg = `# The helm-operator has been configured.
# By default it will use the ssh-key from k3sup app install fluxcd

# Find out more at:
# https://docs.fluxcd.io/projects/helm-operator/en/latest/references/helmrelease-custom-resource.html`
