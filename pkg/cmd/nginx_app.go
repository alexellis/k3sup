package cmd

import (
	"fmt"
	"log"
	"os"
	"path"

	"github.com/alexellis/k3sup/pkg/config"

	"github.com/spf13/cobra"
)

func makeInstallNginx() *cobra.Command {
	var nginx = &cobra.Command{
		Use:          "nginx-ingress",
		Short:        "Install nginx-ingress",
		Long:         `Install nginx-ingress`,
		Example:      `  k3sup app install nginx-ingress --namespace default`,
		SilenceUsage: true,
	}

	nginx.Flags().StringP("namespace", "n", "default", "The namespace used for installation")
	nginx.Flags().Bool("update-repo", true, "Update the helm repo")

	nginx.RunE = func(command *cobra.Command, args []string) error {
		kubeConfigPath := getDefaultKubeconfig()

		if command.Flags().Changed("kubeconfig") {
			kubeConfigPath, _ = command.Flags().GetString("kubeconfig")
		}

		updateRepo, _ := nginx.Flags().GetBool("update-repo")

		fmt.Printf("Using kubeconfig: %s\n", kubeConfigPath)

		userPath, err := config.InitUserDir()
		if err != nil {
			return err
		}
		namespace, _ := command.Flags().GetString("namespace")

		if namespace != "default" {
			return fmt.Errorf(`to override the "default", install via tiller`)
		}

		clientArch, clientOS := getClientArch()

		fmt.Printf("Client: %s, %s\n", clientArch, clientOS)
		log.Printf("User dir established as: %s\n", userPath)

		os.Setenv("HELM_HOME", path.Join(userPath, ".helm"))

		err = tryDownloadHelm(userPath, clientArch, clientOS)
		if err != nil {
			return err
		}

		if updateRepo {
			err = updateHelmRepos()
			if err != nil {
				return err
			}
		}

		chartPath := path.Join(os.TempDir(), "charts")
		err = fetchChart(chartPath, "stable/nginx-ingress")

		if err != nil {
			return err
		}

		overrides := map[string]string{}
		fmt.Println("Chart path: ", chartPath)

		outputPath := path.Join(chartPath, "nginx-ingress/rendered")

		// ns:="kube-system"
		ns := "default"
		err = templateChart(chartPath,
			"nginx-ingress",
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

		fmt.Println(`=======================================================================
= nginx-ingress has been installed.                                   =
=======================================================================


# If you're using a local environment such as "minikube" or "KinD",
# then try the inlets operator with "k3sup app install inlets-operator"

# If you're using a managed Kubernetes service, then you'll find 
# your LoadBalancer's IP under "EXTERNAL-IP" via:

kubectl get svc nginx-ingress-controller

# Find out more at:
# https://github.com/helm/charts/tree/master/stable/nginx-ingress

Thank you for using k3sup!`)

		return nil
	}

	return nginx
}
