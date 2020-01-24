package tools

import (
	"fmt"
	"github.com/alexellis/k3sup/pkg/download"
	"github.com/alexellis/k3sup/pkg/env"
	"github.com/spf13/cobra"
	"os"
	"path"
)

func MakeDownloadFaaSCLI() *cobra.Command {
	var faasCli = &cobra.Command{
		Use:          "faas-cli",
		Short:        "Download faas-cli",
		Long:      	  "Download faas-cli which is used for managing openfaas functions",
		Example:      "k3sup tool download faas-cli",
		SilenceUsage: true,
	}

	faasCli.RunE = func(command *cobra.Command, args []string) error {

		version, err := command.Flags().GetString("version")
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

		outputLocation = path.Join(outputLocation, "faas-cli")
		outputLocation = download.ExpandPath(outputLocation)

		clientArch, clientOS := env.GetClientArch()
		if err := download.FaaSCLI(version, outputLocation, clientArch, clientOS); err != nil {
			return err
		}

		fmt.Printf(`Downloaded faas-cli to %s
`, outputLocation)

		return nil
	}

	return faasCli
}


