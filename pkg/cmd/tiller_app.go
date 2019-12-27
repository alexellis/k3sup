package cmd

import (
	"fmt"
	"log"
	"os"
	"path"

	execute "github.com/alexellis/go-execute/pkg/v1"

	"github.com/alexellis/k3sup/pkg/config"
	"github.com/alexellis/k3sup/pkg/env"

	"github.com/spf13/cobra"
)

func makeInstallTiller() *cobra.Command {
	var tiller = &cobra.Command{
		Use:          "tiller",
		Short:        "Install tiller",
		Long:         `Install tiller`,
		Example:      `  k3sup app install tiller`,
		SilenceUsage: true,
	}

	tiller.RunE = func(command *cobra.Command, args []string) error {
		kubeConfigPath, _ := command.Flags().GetString("kubeconfig")
		fmt.Printf("Using kubeconfig: %s\n", kubeConfigPath)

		arch, err := getNodeArchitecture(kubeConfigPath, "")

		if err != nil {
			return err
		}

		fmt.Printf("Node architecture: %q\n", arch)

		if arch != "x86_64" && arch != "amd64" {
			return fmt.Errorf("This app is not known to work with the %s architecture", arch)
		}

		userPath, err := config.InitUserDir()
		if err != nil {
			return err
		}

		clientArch, clientOS := env.GetClientArch()

		fmt.Printf("Client: %q, %q\n", clientArch, clientOS)

		log.Printf("User dir established as: %s\n", userPath)

		os.Setenv("HELM_HOME", path.Join(userPath, ".helm"))

		task, err := kubectl(kubeConfigPath, "", "-n", "kube-system", "create", "sa", "tiller").Execute()
		if err != nil {
			return err
		}
		if task.ExitCode != 0 {
			return fmt.Errorf("kubectl exit code %d, stderr: %s",
				task.ExitCode,
				task.Stderr)
		}
		fmt.Println(task.Stdout, task.Stderr)

		task, err = kubectl(kubeConfigPath, "", "create", "clusterrolebinding", "tiller", "--clusterrole", "cluster-admin", "--serviceaccount=kube-system:tiller").Execute()
		if err != nil {
			return err
		}
		if task.ExitCode != 0 {
			return fmt.Errorf("kubectl exit code %d, stderr: %s",
				task.ExitCode,
				task.Stderr)
		}

		fmt.Println(task.Stdout, task.Stderr)

		k3supBin := path.Join(userPath, "bin")
		helmInit := execute.ExecTask{
			Command: path.Join(k3supBin, "helm"),
			Args: []string{
				"init",
				"--skip-refresh", "--upgrade", "--service-account", "tiller",
			},
		}
		res, err := helmInit.Execute()
		if err != nil {
			return err
		}

		fmt.Println(res.Stdout, res.Stderr)

		helmBinary, err := tryDownloadHelm(userPath, clientArch, clientOS, false)
		if err != nil {
			return err
		}

		fmt.Println(`=======================================================================
tiller has been installed
=======================================================================

# You can now use helm with tiller from the installation directory

` + helmBinary + `

` + thanksForUsing)

		return nil
	}

	return tiller
}
