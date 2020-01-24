package cmd

import (
	"fmt"
	"github.com/alexellis/k3sup/cmd/tools"
	"github.com/spf13/cobra"
	"strings"
)

func MakeTools() *cobra.Command {
	var toolCommand = &cobra.Command{
		Use:   "tool",
		Short: "Download Tools",
		Long:  `Download a useful selection of download`,
		Example: `  k3sup tool download`,
		SilenceUsage: false,
	}

	var download = &cobra.Command{
		Use:   "download",
		Short: "Download a suported cli or tool",
		Example: `  k3sup tool download [tool]
  k3sup tool download helm3 --output-location ~/
  sudo k3sup tool download faas-cli --output-location /usr/local/bin
  k3sup tool download --help`,
		SilenceUsage: true,
	}


	download.RunE = func(command *cobra.Command, args []string) error {
		if len(args) == 0 {
			fmt.Printf("You can install: %s\n%s\n\n", strings.TrimRight("\n - "+strings.Join(getTools(), "\n - "), "\n - "),
				`Run k3sup tool download NAME --help to see configuration options.`)
			return nil
		}

		return nil
	}

	toolCommand.PersistentFlags().String("version",  "", "The version of the tool to install, defaults to latest")
	toolCommand.PersistentFlags().String("output-location", "", "The location to download the tool to, defaults to current location")

	toolCommand.AddCommand(download)
	download.AddCommand(tools.MakeDownloadFaaSCLI())
	download.AddCommand(tools.MakeDownloadHelm2())
	download.AddCommand(tools.MakeDownloadHelm3())
	download.AddCommand(tools.MakeDownloadKubectl())

	return toolCommand
}

func getTools() []string {
	return []string{"faas-cli",
		"helm2",
		"helm3",
		"kubectl",
	}
}