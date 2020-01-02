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

func makeInstallCronConnector() *cobra.Command {
	var command = &cobra.Command{
		Use:          "cron-connector",
		Short:        "Install cron-connector for OpenFaaS",
		Long:         `Install cron-connector for OpenFaaS`,
		Example:      `  k3sup app install cron-connector`,
		SilenceUsage: true,
	}

	command.Flags().StringP("namespace", "n", "openfaas", "The namespace used for installation")
	command.Flags().Bool("update-repo", true, "Update the helm repo")

	command.RunE = func(command *cobra.Command, args []string) error {

		updateRepo, _ := command.Flags().GetBool("update-repo")

		userPath, err := config.InitUserDir()
		if err != nil {
			return err
		}

		namespace, _ := command.Flags().GetString("namespace")

		if namespace != "openfaas" {
			return fmt.Errorf(`to override the "openfaas", install via tiller`)
		}

		clientArch, clientOS := env.GetClientArch()

		fmt.Printf("Client: %s, %s\n", clientArch, clientOS)
		log.Printf("User dir established as: %s\n", userPath)

		os.Setenv("HELM_HOME", path.Join(userPath, ".helm"))

		_, err = tryDownloadHelm(userPath, clientArch, clientOS, false)
		if err != nil {
			return err
		}

		err = addHelmRepo("openfaas", "https://openfaas.github.io/faas-netes/", false)
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
		err = fetchChart(chartPath, "openfaas/cron-connector", false)

		if err != nil {
			return err
		}

		overrides := map[string]string{}

		arch, err := getNodeArchitecture(command)

		if err != nil {
			return err
		}

		fmt.Printf("Node architecture: %q\n", arch)

		fmt.Println("Chart path: ", chartPath)

		outputPath := path.Join(chartPath, "cron-connector/rendered")

		ns := namespace
		err = templateChart(chartPath,
			"cron-connector",
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
= cron-connector has been installed.                                   =
=======================================================================

# Example usage to trigger nodeinfo every 5 minutes:

faas-cli store deploy nodeinfo \
  --annotation schedule="*/5 * * * *" \
  --annotation topic=cron-function

# View the connector's logs:

kubectl logs deploy/cron-connector -n openfaas -f

# Find out more on the project homepage:

# https://github.com/openfaas-incubator/cron-connector/

` + thanksForUsing)

		return nil
	}

	return command
}
