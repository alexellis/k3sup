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

func MakeInstallNginx() *cobra.Command {
	var nginx = &cobra.Command{
		Use:          "nginx-ingress",
		Short:        "Install nginx-ingress",
		Long:         `Install nginx-ingress. This app can be installed with Host networking for cases where an external LB is not available. please see the --host-mode flag and the nginx-ingress docs for more info`,
		Example:      `  k3sup app install nginx-ingress --namespace default`,
		SilenceUsage: true,
	}

	nginx.Flags().StringP("namespace", "n", "default", "The namespace used for installation")
	nginx.Flags().Bool("update-repo", true, "Update the helm repo")
	nginx.Flags().Bool("host-mode", false, "If we should install nginx-ingress in host mode.")
	nginx.Flags().Bool("helm3", true, "Use helm3, if set to false uses helm2")

	nginx.RunE = func(command *cobra.Command, args []string) error {
		kubeConfigPath := getDefaultKubeconfig()

		if command.Flags().Changed("kubeconfig") {
			kubeConfigPath, _ = command.Flags().GetString("kubeconfig")
		}

		updateRepo, _ := nginx.Flags().GetBool("update-repo")

		fmt.Printf("Using kubeconfig: %s\n", kubeConfigPath)
		helm3, _ := command.Flags().GetBool("helm3")

		if helm3 {
			fmt.Println("Using helm3")
		}

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
		err = fetchChart(chartPath, "stable/nginx-ingress", defaultVersion, helm3)

		if err != nil {
			return err
		}

		overrides := map[string]string{}

		overrides["defaultBackend.enabled"] = "false"

		arch := getNodeArchitecture()
		fmt.Printf("Node architecture: %q\n", arch)

		switch arch {
		case "amd64":
			// use default image
		case "arm", "arm64":
			overrides["controller.image.repository"] = fmt.Sprintf("quay.io/kubernetes-ingress-controller/nginx-ingress-controller-%v", arch)
		default:
			return fmt.Errorf("architecture %v is not supported by ingress-nginx", arch)
		}

		hostMode, flagErr := command.Flags().GetBool("host-mode")
		if flagErr != nil {
			return flagErr
		}
		if hostMode {
			fmt.Println("Running in host networking mode")
			overrides["controller.hostNetwork"] = "true"
			overrides["controller.daemonset.useHostPort"] = "true"
			overrides["dnsPolicy"] = "ClusterFirstWithHostNet"
			overrides["controller.kind"] = "DaemonSet"
		}
		fmt.Println("Chart path: ", chartPath)

		ns := "default"

		if helm3 {
			outputPath := path.Join(chartPath, "nginx-ingress")

			err := helm3Upgrade(outputPath, "stable/nginx-ingress", ns,
				"values.yaml",
				"",
				overrides)

			if err != nil {
				return err
			}
		} else {
			outputPath := path.Join(chartPath, "nginx-ingress/rendered")

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
		}

		fmt.Println(nginxIngressInstallMsg)

		return nil
	}

	return nginx
}

const NginxIngressInfoMsg = `# If you're using a local environment such as "minikube" or "KinD",
# then try the inlets operator with "k3sup app install inlets-operator"

# If you're using a managed Kubernetes service, then you'll find
# your LoadBalancer's IP under "EXTERNAL-IP" via:

kubectl get svc nginx-ingress-controller

# Find out more at:
# https://github.com/helm/charts/tree/master/stable/nginx-ingress`

const nginxIngressInstallMsg = `=======================================================================
= nginx-ingress has been installed.                                   =
=======================================================================` +
	"\n\n" + NginxIngressInfoMsg + "\n\n" + pkg.ThanksForUsing
