package cmd

import (
	"fmt"
	"log"
	"os"
	"path"

	"github.com/alexellis/k3sup/pkg/config"
	"github.com/spf13/cobra"
)

func makeInstallCertManager() *cobra.Command {
	var certManager = &cobra.Command{
		Use:          "cert-manager",
		Short:        "Install cert-manager",
		Long:         "Install cert-manager for obtaining TLS certificates from LetsEncrypt",
		Example:      "k3sup install cert-manager",
		SilenceUsage: true,
	}

	certManager.Flags().StringP("namespace", "n", "cert-manager", "The namespace to install cert-manager")
	certManager.Flags().Bool("update-repo", true, "Update the helm repo")

	certManager.RunE = func(command *cobra.Command, args []string) error {
		kubeConfigPath := getDefaultKubeconfig()

		if command.Flags().Changed("kubeconfig") {
			kubeConfigPath, _ = command.Flags().GetString("kubeconfig")
		}

		fmt.Printf("Using kubeconfig: %s\n", kubeConfigPath)

		namespace, _ := command.Flags().GetString("namespace")

		if namespace != "cert-manager" {
			return fmt.Errorf(`To override the "cert-manager" namespace, install cert-manager via helm manually`)
		}

		userPath, err := config.InitUserDir()
		if err != nil {
			return err
		}

		clientArch, clientOS := getClientArch()

		fmt.Printf("Client: %s, %s\n", clientArch, clientOS)

		log.Printf("User dir established as: %s\n", userPath)

		os.Setenv("HELM_HOME", path.Join(userPath, ".helm"))

		_, err = tryDownloadHelm(userPath, clientArch, clientOS)
		if err != nil {
			return err
		}

		err = addHelmRepo("jetstack", "https://charts.jetstack.io")
		if err != nil {
			return err
		}

		updateRepo, _ := certManager.Flags().GetBool("update-repo")

		if updateRepo {
			err = updateHelmRepos()
			if err != nil {
				return err
			}
		}

		nsRes, nsErr := kubectlTask("create", "namespace", namespace)
		if nsErr != nil {
			return nsErr
		}

		if nsRes.ExitCode != 0 {
			fmt.Printf("[Warning] unable to create namespace %s, may already exist: %s", namespace, nsRes.Stderr)
		}

		chartPath := path.Join(os.TempDir(), "charts")

		err = fetchChart(chartPath, "jetstack/cert-manager")
		if err != nil {
			return err
		}

		outputPath := path.Join(chartPath, "cert-manager/rendered")

		err = templateChart(chartPath, "cert-manager", namespace, outputPath, "values.yaml", nil)
		if err != nil {
			return err
		}

		log.Printf("Applying CRD\n")
		crdsURL := "https://raw.githubusercontent.com/jetstack/cert-manager/release-0.12/deploy/manifests/00-crds.yaml"
		res, err := kubectlTask("apply", "--validate=false", "-f",
			crdsURL)
		if err != nil {
			return err
		}

		if res.ExitCode > 0 {
			return fmt.Errorf("Error applying CRD from: %s, error: %s", crdsURL, res.Stderr)
		}

		applyRes, applyErr := kubectlTask("apply", "-R", "-f", outputPath)
		if applyErr != nil {
			return applyErr
		}

		if applyRes.ExitCode > 0 {
			return fmt.Errorf("Error applying templated YAML files, error: %s", applyRes.Stderr)
		}

		fmt.Println(`=======================================================================
= cert-manager has been installed.                                    =
=======================================================================

# Get started with cert-manager here:
# https://docs.cert-manager.io/en/latest/tutorials/acme/http-validation.html

# Check cert-manager's logs with:

kubectl logs -n cert-manager deploy/cert-manager -f

` + thanksForUsing)

		return nil
	}

	return certManager
}
