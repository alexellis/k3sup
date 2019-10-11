package cmd

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"path"
	"strings"

	"github.com/alexellis/go-execute"

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

	openfaas.Flags().Bool("loadbalancer", false, "add a loadbalancer")
	openfaas.Flags().StringP("namespace", "n", "openfaas", "Namespace for core services")

	openfaas.RunE = func(command *cobra.Command, args []string) error {
		kubeConfigPath := path.Join(os.Getenv("HOME"), ".kube/config")

		if val, ok := os.LookupEnv("KUBECONFIG"); ok {
			kubeConfigPath = val
		}

		if command.Flags().Changed("kubeconfig") {
			kubeConfigPath, _ = command.Flags().GetString("kubeconfig")
		}

		fmt.Printf("Using context: %s\n", kubeConfigPath)

		arch := getArchitecture()
		fmt.Printf("Node architecture: %s\n", arch)

		valuesSuffix := ""
		switch arch {
		case "arm":
			valuesSuffix = "-armhf"
		}

		userPath, err := config.InitUserDir()
		if err != nil {
			return err
		}

		clientArch, clientOS := getClientArch()

		fmt.Printf("Client: %s, %s\n", clientArch, clientOS)

		log.Printf("User dir established as: %s\n", userPath)

		os.Setenv("HELM_HOME", path.Join(userPath, ".helm"))

		if _, statErr := os.Stat(path.Join(path.Join(userPath, ".bin"), "helm")); statErr != nil {
			downloadHelm(userPath, clientArch, clientOS)

			err = helmInit()
			if err != nil {
				return err
			}
		}

		// lb, _ := command.Flags().GetBool("loadbalancer")

		namespace, _ := command.Flags().GetString("namespace")

		err = addHelmRepo("openfaas", "https://openfaas.github.io/faas-netes/")
		if err != nil {
			return err
		}

		err = updateHelmRepos()

		if err != nil {
			return err
		}

		err = kubectl("apply", "-f", "https://raw.githubusercontent.com/openfaas/faas-netes/master/namespaces.yml")

		if err != nil {
			return err
		}

		pass, err := password.Generate(25, 10, 0, false, true)
		if err != nil {
			return err
		}

		_, err = kubectlTask("-n", namespace, "create", "secret", "generic", "basic-auth", "--from-literal=basic-auth-user=admin", `--from-literal=basic-auth-password=`+pass)

		if err != nil {
			return err
		}

		chartPath := path.Join(os.TempDir(), "charts")

		err = fetchChart(chartPath, "openfaas/openfaas")

		if err != nil {
			return err
		}

		outputPath := path.Join(chartPath, "openfaas/rendered")
		err = templateChart(chartPath, "openfaas", namespace, outputPath, "values"+valuesSuffix+".yaml")
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
If basic auth is enabled, you can now log into your gateway.

# Get the faas-cli
curl -SLsf https://cli.openfaas.com | sudo sh

# Forward the gateway to your machine
kubectl rollout status -n openfaas deploy/gateway
kubectl port-forward -n openfaas svc/gateway 8080:8080 &

# Get your password and log in
PASSWORD=$(kubectl get secret -n openfaas basic-auth -o jsonpath="{.data.basic-auth-password}" | base64 --decode; echo)
echo -n $PASSWORD | faas-cli login --username admin --password-stdin

Thank you for using k3sup!`)

		return nil
	}

	return openfaas
}

func getClientArch() (string, string) {
	task := execute.ExecTask{Command: "uname", Args: []string{"-m"}}
	res, err := task.Execute()
	if err != nil {
		log.Println(err)
	}

	arch := strings.TrimSpace(res.Stdout)

	taskOS := execute.ExecTask{Command: "uname", Args: []string{"-s"}}
	resOS, errOS := taskOS.Execute()
	if errOS != nil {
		log.Println(errOS)
	}

	os := strings.TrimSpace(resOS.Stdout)

	return arch, os
}

func getHelmURL(arch, os, version string) string {
	archSuffix := "amd64"
	osSuffix := strings.ToLower(os)

	if strings.HasPrefix(arch, "armv7") {
		archSuffix = "arm"
	} else if strings.HasPrefix(arch, "aarch64") {
		archSuffix = "arm64"
	}

	return fmt.Sprintf("https://get.helm.sh/helm-%s-%s-%s.tar.gz", version, osSuffix, archSuffix)
}

func downloadHelm(userPath, clientArch, clientOS string) error {
	helmURL := getHelmURL(clientArch, clientOS, "v2.14.3")
	fmt.Println(helmURL)
	parsedURL, _ := url.Parse(helmURL)

	res, err := http.DefaultClient.Get(parsedURL.String())
	if err != nil {
		return err
	}

	defer res.Body.Close()
	r := ioutil.NopCloser(res.Body)
	untarErr := Untar(r, path.Join(userPath, ".bin"))
	if untarErr != nil {
		return untarErr
	}

	return nil
}
