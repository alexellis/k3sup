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

func makeInstallMongoDB() *cobra.Command {
	var command = &cobra.Command{
		Use:          "mongodb",
		Short:        "Install mongodb",
		Long:         `Install mongodb`,
		Example:      `  k3sup app install mongodb`,
		SilenceUsage: true,
	}
	command.Flags().String("namespace", "default", "Namespace for the app")

	command.Flags().StringArray("set", []string{},
		"Use custom flags or override existing flags \n(example --set=prometheus.enabled=false)")

	command.RunE = func(command *cobra.Command, args []string) error {

		kubeConfigPath := getDefaultKubeconfig()

		if command.Flags().Changed("kubeconfig") {
			kubeConfigPath, _ = command.Flags().GetString("kubeconfig")
		}

		fmt.Printf("Using kubeconfig: %s\n", kubeConfigPath)

		namespace, _ := command.Flags().GetString("namespace")

		userPath, err := config.InitUserDir()
		if err != nil {
			return err
		}

		clientArch, clientOS := env.GetClientArch()

		fmt.Printf("Client: %q, %q\n", clientArch, clientOS)

		log.Printf("User dir established as: %s\n", userPath)

		os.Setenv("HELM_HOME", path.Join(userPath, ".helm"))

		os.Setenv("HELM_VERSION", helm3Version)
		helm3 := true
		_, err = tryDownloadHelm(userPath, clientArch, clientOS, helm3)
		if err != nil {
			return err
		}

		err = addHelmRepo("stable", "https://kubernetes-charts.storage.googleapis.com/", helm3)
		if err != nil {
			return fmt.Errorf("unable to add repo %s", err)
		}

		updateRepo, _ := command.Flags().GetBool("update-repo")

		if updateRepo {
			err = updateHelmRepos(helm3)
			if err != nil {
				return fmt.Errorf("unable to update repos %s", err)
			}
		}

		chartPath := path.Join(os.TempDir(), "charts")

		err = fetchChart(chartPath, "stable/mongodb", helm3)

		if err != nil {
			return fmt.Errorf("unable fetch chart %s", err)
		}

		overrides := map[string]string{}

		outputPath := path.Join(chartPath, "mongodb")

		customFlags, err := command.Flags().GetStringArray("set")
		if err != nil {
			return fmt.Errorf("error with --set usage: %s", err)
		}

		if err := mergeFlags(overrides, customFlags); err != nil {
			return err
		}

		err = helm3Upgrade(outputPath, "stable/mongodb",
			namespace, "values.yaml", overrides, false)
		if err != nil {
			return fmt.Errorf("unable to mongodb chart with helm %s", err)
		}
		fmt.Println(mongoDBPostInstallMsg)
		return nil
	}
	return command
}

const mongoDBPostInstallMsg = `=======================================================================
=                  MongoDB has been installed.                        =
=======================================================================` +
	"\n\n" + thanksForUsing
