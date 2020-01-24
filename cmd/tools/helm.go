package tools

import (
	"fmt"
	"github.com/alexellis/k3sup/pkg/download"
	"github.com/alexellis/k3sup/pkg/env"
	"github.com/spf13/cobra"
	"os"
)


func MakeDownloadHelm2() *cobra.Command {
	var helm = &cobra.Command{
		Use:          "helm2",
		Short:        "Download helm2",
		Long:      	  "Download helm2, the client side cli tool for managing charts",
		Example:      `k3sup tool download helm 
k3sup tool download helm2 --version v2.16.1`,
		SilenceUsage: true,
	}

	helm.RunE = func(command *cobra.Command, args []string) error {

		version, err := command.Flags().GetString("version")
		if err != nil {
			return err
		}
		outputLocation, locErr := command.Flags().GetString("output-location")
		if locErr != nil {
			return locErr
		}

		return downloadHelm(version, outputLocation, true)

	}

	return helm
}

func MakeDownloadHelm3() *cobra.Command {
	var helm = &cobra.Command{
		Use:          "helm3",
		Short:        "Download helm3",
		Long:      	  "Download helm3, the client side cli tool for managing charts",
		Example:      `k3sup tool download helm3
k3sup tool download helm3 --version v3.0.2`,
		SilenceUsage: true,
	}

	helm.RunE = func(command *cobra.Command, args []string) error {

		version, err := command.Flags().GetString("version")
		if err != nil {
			return err
		}

		outputLocation, locErr := command.Flags().GetString("output-location")
		if locErr != nil {
			return locErr
		}

		return downloadHelm(version, outputLocation, false)

	}

	return helm
}

func downloadHelm(version, outputLocation string, helm2 bool) error {

	if len(outputLocation) == 0 {
		var err error
		outputLocation, err = os.Getwd()
		if err != nil {
			return err
		}
	}

	var location string
	var err error
	outputLocation = download.ExpandPath(outputLocation)

	clientArch, clientOS := env.GetClientArch()
	if len(version) == 0 {
		if helm2 {
			location, err = download.DownloadHelm(outputLocation, clientArch, clientOS, false)
		} else {
			location, err = download.DownloadHelm(outputLocation, clientArch, clientOS, true)
		}
	} else {
		location, err = download.DownloadHelmVersion(outputLocation, clientArch, clientOS, version, true)
		if err != nil {
			return err
		}
	}


	fmt.Printf(`Downloaded helm to %s
`, location)

	return nil
}