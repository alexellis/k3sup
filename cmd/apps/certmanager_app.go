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

func MakeInstallCertManager() *cobra.Command {
	var certManager = &cobra.Command{
		Use:          "cert-manager",
		Short:        "Install cert-manager",
		Long:         "Install cert-manager for obtaining TLS certificates from LetsEncrypt",
		Example:      "k3sup install cert-manager",
		SilenceUsage: true,
	}

	certManager.Flags().StringP("namespace", "n", "cert-manager", "The namespace to install cert-manager")
	certManager.Flags().Bool("update-repo", true, "Update the helm repo")
	certManager.Flags().Bool("helm3", true, "Use helm3, if set to false uses helm2")

	certManager.RunE = func(command *cobra.Command, args []string) error {
		wait, _ := command.Flags().GetBool("wait")
		const certManagerVersion = "v0.12.0"
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

		if namespace != "cert-manager" {
			return fmt.Errorf(`To override the "cert-manager" namespace, install cert-manager via helm manually`)
		}

		userPath, err := config.InitUserDir()
		if err != nil {
			return err
		}

		clientArch, clientOS := env.GetClientArch()

		fmt.Printf("Client: %s, %s\n", clientArch, clientOS)

		log.Printf("User dir established as: %s\n", userPath)

		os.Setenv("HELM_HOME", path.Join(userPath, ".helm"))

		_, err = helm.TryDownloadHelm(userPath, clientArch, clientOS, helm3)
		if err != nil {
			return err
		}

		err = addHelmRepo("jetstack", "https://charts.jetstack.io", helm3)
		if err != nil {
			return err
		}

		updateRepo, _ := certManager.Flags().GetBool("update-repo")

		if updateRepo {
			err = updateHelmRepos(helm3)
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

		err = fetchChart(chartPath, "jetstack/cert-manager", certManagerVersion, helm3)
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
			return fmt.Errorf("error applying CRD from: %s, error: %s", crdsURL, res.Stderr)
		}

		outputPath := path.Join(chartPath, "cert-manager/rendered")
		overrides := map[string]string{}

		if helm3 {
			outputPath := path.Join(chartPath, "cert-manager")

			err := helm3Upgrade(outputPath, "jetstack/cert-manager", namespace,
				"values.yaml",
				"v0.12.0",
				overrides,
				wait)

			if err != nil {
				return err
			}
		} else {
			err = templateChart(chartPath, "cert-manager", namespace, outputPath, "values.yaml", nil)
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

		fmt.Println(certManagerInstallMsg)

		return nil
	}

	return certManager
}

const CertManagerInfoMsg = `# Get started with cert-manager here:
# https://docs.cert-manager.io/en/latest/tutorials/acme/http-validation.html

# Check cert-manager's logs with:

kubectl logs -n cert-manager deploy/cert-manager -f`

const certManagerInstallMsg = `=======================================================================
= cert-manager  has been installed.                                   =
=======================================================================` +
	"\n\n" + CertManagerInfoMsg + "\n\n" + pkg.ThanksForUsing
