package cmd

import (
	"fmt"
	"log"
	"os"
	"path"

	"github.com/alexellis/k3sup/pkg/config"
	"github.com/alexellis/k3sup/pkg/env"

	"github.com/spf13/cobra"
)

func makeInstallNginx() *cobra.Command {
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

	nginx.RunE = func(command *cobra.Command, args []string) error {
		kubeConfigPath, _ := command.Flags().GetString("kubeconfig")

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

		clientArch, clientOS := env.GetClientArch()

		fmt.Printf("Client: %s, %s\n", clientArch, clientOS)
		log.Printf("User dir established as: %s\n", userPath)

		os.Setenv("HELM_HOME", path.Join(userPath, ".helm"))

		_, err = tryDownloadHelm(userPath, clientArch, clientOS, false)
		if err != nil {
			return err
		}

		if updateRepo {
			err = updateHelmRepos(false)
			if err != nil {
				return err
			}
		}

		chartPath := path.Join(os.TempDir(), "charts")
		err = fetchChart(chartPath, "stable/nginx-ingress", false)

		if err != nil {
			return err
		}

		overrides := map[string]string{}

		overrides["defaultBackend.enabled"] = "false"

		arch, err := getNodeArchitecture(command)

		if err != nil {
			return err
		}

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

		outputPath := path.Join(chartPath, "nginx-ingress/rendered")

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

		res, err := kubectl(command, "apply", "-R", "-f", outputPath).Execute()

		if err != nil {
			return err
		}

		if res.ExitCode != 0 {
			return fmt.Errorf("kubectl exit code %d, stderr: %s",
				res.ExitCode,
				res.Stderr)
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

` + thanksForUsing)

		return nil
	}

	return nginx
}
