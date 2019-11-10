package cmd

import (
	"fmt"
	"log"
	"os"
	"path"
	"strconv"
	"strings"

	"github.com/sethvargo/go-password/password"

	"github.com/alexellis/k3sup/pkg/config"

	"github.com/spf13/cobra"
)

func makeInstallOpenFaaS() *cobra.Command {
	var openfaas = &cobra.Command{
		Use:          "openfaas",
		Short:        "Install openfaas",
		Long:         `Install openfaas`,
		Example:      `  k3sup app install openfaas --loadbalancer`,
		SilenceUsage: true,
	}

	openfaas.Flags().BoolP("basic-auth", "a", true, "Enable authentication")
	openfaas.Flags().BoolP("load-balancer", "l", false, "Add a loadbalancer")
	openfaas.Flags().StringP("namespace", "n", "openfaas", "The namespace for the core services")
	openfaas.Flags().Bool("update-repo", true, "Update the helm repo")

	openfaas.RunE = func(command *cobra.Command, args []string) error {
		kubeConfigPath := getDefaultKubeconfig()

		if command.Flags().Changed("kubeconfig") {
			kubeConfigPath, _ = command.Flags().GetString("kubeconfig")
		}

		fmt.Printf("Using kubeconfig: %s\n", kubeConfigPath)

		namespace, _ := command.Flags().GetString("namespace")

		if namespace != "openfaas" {
			return fmt.Errorf(`to override the "openfaas", install OpenFaaS via helm manually`)
		}

		arch := getArchitecture()
		fmt.Printf("Node architecture: %q\n", arch)

		valuesSuffix := getValuesSuffix(arch)

		userPath, err := config.InitUserDir()
		if err != nil {
			return err
		}

		clientArch, clientOS := getClientArch()

		fmt.Printf("Client: %q, %q\n", clientArch, clientOS)

		log.Printf("User dir established as: %s\n", userPath)

		os.Setenv("HELM_HOME", path.Join(userPath, ".helm"))

		_, err = tryDownloadHelm(userPath, clientArch, clientOS)
		if err != nil {
			return err
		}

		err = addHelmRepo("openfaas", "https://openfaas.github.io/faas-netes/")
		if err != nil {
			return err
		}

		updateRepo, _ := openfaas.Flags().GetBool("update-repo")

		if updateRepo {
			err = updateHelmRepos()
			if err != nil {
				return err
			}
		}

		if err != nil {
			return err
		}

		_, err = kubectlTask("apply", "-f",
			"https://raw.githubusercontent.com/openfaas/faas-netes/master/namespaces.yml")

		if err != nil {
			return err
		}

		pass, err := password.Generate(25, 10, 0, false, true)
		if err != nil {
			return err
		}

		res, secretErr := kubectlTask("-n", namespace, "create", "secret", "generic",
			"basic-auth",
			"--from-literal=basic-auth-user=admin",
			`--from-literal=basic-auth-password=`+pass)

		if secretErr != nil {
			return secretErr
		}

		if res.ExitCode != 0 {
			fmt.Printf("[Warning] unable to create secret %s, may already exist: %s", "basic-auth", res.Stderr)
		}

		chartPath := path.Join(os.TempDir(), "charts")

		err = fetchChart(chartPath, "openfaas/openfaas")

		if err != nil {
			return err
		}

		overrides := map[string]string{}

		basicAuth, _ := command.Flags().GetBool("basic-auth")
		overrides["basicAuth"] = strings.ToLower(strconv.FormatBool(basicAuth))

		overrides["serviceType"] = "NodePort"
		lb, _ := command.Flags().GetBool("load-balancer")
		if lb {
			overrides["serviceType"] = "LoadBalancer"
		}

		outputPath := path.Join(chartPath, "openfaas/rendered")
		err = templateChart(chartPath, "openfaas",
			namespace,
			outputPath,
			"values"+valuesSuffix+".yaml",
			overrides)

		if err != nil {
			return err
		}

		err = kubectl("apply", "-R", "-f", outputPath)

		if err != nil {
			return err
		}

		fmt.Println(`=======================================================================
= OpenFaaS has been installed.                                        =
=======================================================================

# Get the faas-cli
curl -SLsf https://cli.openfaas.com | sudo sh

# Forward the gateway to your machine
kubectl rollout status -n openfaas deploy/gateway
kubectl port-forward -n openfaas svc/gateway 8080:8080 &

# If basic auth is enabled, you can now log into your gateway:
PASSWORD=$(kubectl get secret -n openfaas basic-auth -o jsonpath="{.data.basic-auth-password}" | base64 --decode; echo)
echo -n $PASSWORD | faas-cli login --username admin --password-stdin

faas-cli store deploy figlet
faas-cli list

# For Raspberry Pi
faas-cli store list \
 --platform armhf

faas-cli store deploy figlet \
 --platform armhf

# Find out more at:
# https://github.com/openfaas/faas

Thank you for using k3sup!`)

		return nil
	}

	return openfaas
}

func getValuesSuffix(arch string) string {
	var valuesSuffix string
	switch arch {
	case "arm":
		valuesSuffix = "-armhf"
		break
	case "arm64", "aarch64":
		valuesSuffix = "-arm64"
		break
	default:
		valuesSuffix = ""
	}
	return valuesSuffix
}
