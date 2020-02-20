package apps

import (
	"fmt"
	"log"
	"os"
	"path"

	"golang.org/x/crypto/bcrypt"

	"github.com/alexellis/k3sup/pkg"
	"github.com/alexellis/k3sup/pkg/config"
	"github.com/alexellis/k3sup/pkg/env"
	"github.com/alexellis/k3sup/pkg/helm"
	"github.com/sethvargo/go-password/password"
	"github.com/spf13/cobra"
)

func MakeInstallRegistry() *cobra.Command {
	var registry = &cobra.Command{
		Use:          "docker-registry",
		Short:        "Install a Docker registry",
		Long:         `Install a Docker registry`,
		Example:      `  k3sup app install registry --namespace default`,
		SilenceUsage: true,
	}

	registry.Flags().StringP("namespace", "n", "default", "The namespace used for installation")
	registry.Flags().Bool("update-repo", true, "Update the helm repo")
	registry.Flags().Bool("helm3", true, "Use helm3, if set to false uses helm2")
	registry.Flags().StringP("username", "u", "admin", "Username for the registry")
	registry.Flags().StringP("password", "p", "", "Password for the registry, leave blank to generate")

	registry.RunE = func(command *cobra.Command, args []string) error {
		kubeConfigPath := getDefaultKubeconfig()

		if command.Flags().Changed("kubeconfig") {
			kubeConfigPath, _ = command.Flags().GetString("kubeconfig")
		}

		updateRepo, _ := registry.Flags().GetBool("update-repo")

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

		username, _ := command.Flags().GetString("username")

		pass, _ := command.Flags().GetString("password")
		if len(pass) == 0 {
			key, err := password.Generate(20, 10, 0, false, true)
			if err != nil {
				return err
			}

			pass = key
		}

		val, err := bcrypt.GenerateFromPassword([]byte(pass), bcrypt.DefaultCost)

		if err != nil {
			return err
		}

		htPasswd := fmt.Sprintf("%s:%s\n", username, string(val))

		err = addHelmRepo("stable", "https://kubernetes-charts.storage.googleapis.com", helm3)
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
		err = fetchChart(chartPath, "stable/docker-registry", helm3)

		if err != nil {
			return err
		}

		overrides := map[string]string{}

		overrides["persistence.enabled"] = "false"
		overrides["secrets.htpasswd"] = string(htPasswd)

		arch := getNodeArchitecture()
		fmt.Printf("Node architecture: %q\n", arch)

		fmt.Println("Chart path: ", chartPath)

		wait := false
		ns := "default"

		if helm3 {
			outputPath := path.Join(chartPath, "docker-registry")

			err := helm3Upgrade(outputPath, "stable/docker-registry", ns,
				"values.yaml",
				overrides, wait)

			if err != nil {
				return err
			}
		} else {
			outputPath := path.Join(chartPath, "docker-registry/rendered")

			err = templateChart(chartPath,
				"docker-registry",
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

		fmt.Println(registryInstallMsg)
		fmt.Printf("Registry credentials: %s %s\nexport PASSWORD=%s\n", username, pass, pass)

		return nil
	}

	return registry
}

const registryInfoMsg = `# Your docker-registry has been configured

kubectl logs deploy/docker-registry

export IP="192.168.0.11" # Set to WiFI/ethernet adapter
export PASSWORD="" # See below
kubectl port-forward svc/docker-registry --address 0.0.0.0 5000 &

docker login $IP:5000 --username admin --password $PASSWORD
docker tag alpine:3.11 $IP:5000/alpine:3.11
docker push $IP:5000/alpine:3.11

# Find out more at:
# https://github.com/helm/charts/tree/master/stable/registry`

const registryInstallMsg = `=======================================================================
= docker-registry has been installed.                                 =
=======================================================================` +
	"\n\n" + registryInfoMsg + "\n\n" + pkg.ThanksForUsing
