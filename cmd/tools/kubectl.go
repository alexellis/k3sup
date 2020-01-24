package tools

import (
	"github.com/alexellis/k3sup/pkg/download"
	"github.com/alexellis/k3sup/pkg/env"
	"github.com/spf13/cobra"
	"os"
	"path"
)

func MakeDownloadKubectl() *cobra.Command {
	var kubectlCmd = &cobra.Command{
		Use:          "kubectl",
		Short:        "Download kubectl",
		Long:         "Download kubectl, the client side cli tool for managing kubernetes",
		Example:      "k3sup tool download kubectl",
		SilenceUsage: true,
	}

	kubectlCmd.RunE = func(command *cobra.Command, args []string) error {

		version, err := command.Flags().GetString("version")
		if len(version) == 0 {
			version = ""
		}

		if err != nil {
			return err
		}
		outputLocation, locErr := command.Flags().GetString("output-location")
		if locErr != nil {
			return locErr
		}

		if len(outputLocation) == 0 {
			outputLocation, err = os.Getwd()
			if err != nil {
				return err
			}
		}

		outputLocation = path.Join(outputLocation, "kubectl")
		outputLocation = download.ExpandPath(outputLocation)

		clientArch, clientOS := env.GetClientArch()

		if err := download.DownloadKubectl(version, clientArch, clientOS, outputLocation); err != nil {
			return err
		}

		return nil
	}
	return kubectlCmd
}