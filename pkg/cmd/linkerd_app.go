package cmd

import (
	"bufio"
	"fmt"
	"github.com/alexellis/k3sup/pkg/config"
	"github.com/spf13/cobra"
	"io/ioutil"
	"log"
	"os"
)

func makeInstallLinkerd() *cobra.Command {
	var linkerd = &cobra.Command{
		Use:          "linkerd",
		Short:        "Install linkerd",
		Long:         `Install linkerd`,
		Example:      `  k3sup app install linkerd`,
		SilenceUsage: true,
	}

	linkerd.RunE = func(command *cobra.Command, args []string) error {
		kubeConfigPath := getDefaultKubeconfig()

		if command.Flags().Changed("kubeconfig") {
			kubeConfigPath, _ = command.Flags().GetString("kubeconfig")
		}
		fmt.Printf("Using kubeconfig: %s\n", kubeConfigPath)
		arch := getArchitecture()
		fmt.Printf("Node architecture: %q\n", arch)

		userPath, err := config.InitUserDir()
		if err != nil {
			return err
		}

		_, clientOS := getClientArch()

		fmt.Printf("Client: %q\n", clientOS)

		log.Printf("User dir established as: %s\n", userPath)

		err = downloadLinkerd(userPath, clientOS)
		if err != nil {
			return err
		}
		_, err = linkerdCli("check", "--pre")
		if err != nil {
			return err
		}

		res, err := linkerdCli("install")
		if err != nil {
			return err
		}
		file, err := ioutil.TempFile("", "linkerd")
		w := bufio.NewWriter(file)
		_, err = w.WriteString(res.Stdout)
		if err != nil {
			return err
		}
		w.Flush()

		err = kubectl("apply", "-R", "-f", file.Name())
		if err != nil {
			return err
		}

		defer os.Remove(file.Name())

		_, err = linkerdCli("check")
		if err != nil {
			return err
		}
		fmt.Println(`=======================================================================
= Linkerd has been installed.                                        =
=======================================================================

# Get the linkerd-cli
curl -sL https://run.linkerd.io/install | sh

# Find out more at:
# https://github.com/openfaas/faas

Thank you for using k3sup!`)
		return nil
	}

	return linkerd
}
