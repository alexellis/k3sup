package apps

import (
	"fmt"
	"log"
	"os"
	"path"

	execute "github.com/alexellis/go-execute/pkg/v1"
	"github.com/alexellis/k3sup/pkg"
	"github.com/alexellis/k3sup/pkg/env"
	"github.com/alexellis/k3sup/pkg/helm"
	"github.com/spf13/cobra"
)

func MakeInstallTiller() *cobra.Command {
	var tiller = &cobra.Command{
		Use:          "tiller",
		Short:        "Install tiller",
		Long:         `Install tiller`,
		Example:      `  k3sup app install tiller`,
		SilenceUsage: true,
	}

	tiller.RunE = func(command *cobra.Command, args []string) error {
		kubeConfigPath := getDefaultKubeconfig()

		if command.Flags().Changed("kubeconfig") {
			kubeConfigPath, _ = command.Flags().GetString("kubeconfig")
		}

		fmt.Printf("Using kubeconfig: %s\n", kubeConfigPath)

		arch := getNodeArchitecture()
		fmt.Printf("Node architecture: %q\n", arch)

		if arch != "x86_64" && arch != "amd64" {
			return fmt.Errorf("This app is not known to work with the %s architecture", arch)
		}

		userPath, err := getUserPath()
		if err != nil {
			return err
		}

		clientArch, clientOS := env.GetClientArch()

		fmt.Printf("Client: %q, %q\n", clientArch, clientOS)

		log.Printf("User dir established as: %s\n", userPath)

		os.Setenv("HELM_HOME", path.Join(userPath, ".helm"))

		task, err := kubectlTask("-n", "kube-system", "create", "sa", "tiller")
		if err != nil {
			return err
		}

		fmt.Println(task.Stdout, task.Stderr)

		task, err = kubectlTask("create", "clusterrolebinding", "tiller", "--clusterrole", "cluster-admin", "--serviceaccount=kube-system:tiller")
		if err != nil {
			return err
		}
		fmt.Println(task.Stdout, task.Stderr)

		k3supBin := path.Join(userPath, "bin")
		helmInit := execute.ExecTask{
			Command: path.Join(k3supBin, "helm"),
			Args: []string{
				"init",
				"--skip-refresh", "--upgrade", "--service-account", "tiller",
			},
			StreamStdio: true,
		}

		res, err := helmInit.Execute()
		if err != nil {
			return err
		}

		fmt.Println(res.Stdout, res.Stderr)

		_, err = helm.TryDownloadHelm(userPath, clientArch, clientOS, false)
		if err != nil {
			return err
		}

		fmt.Println(tillerInstallMsg)

		return nil
	}

	return tiller
}

func getClientArch() (string, string) {
	clientArch, clientOS := env.GetClientArch()
	return clientArch, clientOS
}

func getHelmBinaryPath() string {
	userPath, _ := getUserPath()
	helmBinaryPath := path.Join(path.Join(userPath, "bin"), "helm")
	return helmBinaryPath
}

var helmBinaryPath = getHelmBinaryPath()

var TillerInfoMsg = `# You can now use helm with tiller from the installation directory
` + helmBinaryPath

var tillerInstallMsg = `=======================================================================
= tiller has been installed.                   	                      =
=======================================================================` +
	"\n\n" + TillerInfoMsg + "\n\n" + pkg.ThanksForUsing
