package apps

import (
	"fmt"
	"log"
	"os"
	"path"
	"strconv"
	"strings"

	"github.com/alexellis/k3sup/pkg"
	"github.com/alexellis/k3sup/pkg/config"
	"github.com/alexellis/k3sup/pkg/env"
	"github.com/alexellis/k3sup/pkg/helm"
	"github.com/spf13/cobra"
)

func MakeInstallPostgresql() *cobra.Command {
	var postgresql = &cobra.Command{
		Use:          "postgresql",
		Short:        "Install postgresql",
		Long:         `Install postgresql`,
		Example:      `  k3sup app install postgresql`,
		SilenceUsage: true,
	}

	postgresql.Flags().Bool("update-repo", true, "Update the helm repo")
	postgresql.Flags().String("namespace", "default", "Kubernetes namespace for the application")

	postgresql.Flags().Bool("persistence", false, "Enable persistence")

	postgresql.Flags().StringArray("set", []string{},
		"Use custom flags or override existing flags \n(example --set persistence.enabled=true)")

	postgresql.RunE = func(command *cobra.Command, args []string) error {
		kubeConfigPath := getDefaultKubeconfig()

		if command.Flags().Changed("kubeconfig") {
			kubeConfigPath, _ = command.Flags().GetString("kubeconfig")
		}
		updateRepo, _ := postgresql.Flags().GetBool("update-repo")

		fmt.Printf("Using kubeconfig: %s\n", kubeConfigPath)

		userPath, err := config.InitUserDir()
		if err != nil {
			return err
		}

		clientArch, clientOS := env.GetClientArch(true)

		fmt.Printf("Client: %s, %s\n", clientArch, clientOS)
		log.Printf("User dir established as: %s\n", userPath)

		os.Setenv("HELM_HOME", path.Join(userPath, ".helm"))

		ns, _ := postgresql.Flags().GetString("namespace")

		if ns != "default" {
			return fmt.Errorf("please use the helm chart if you'd like to change the namespace to %s", ns)
		}

		_, err = helm.TryDownloadHelm(userPath, clientArch, clientOS, false)
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
		err = fetchChart(chartPath, "stable/postgresql", false)

		if err != nil {
			return err
		}

		persistence, _ := postgresql.Flags().GetBool("persistence")

		overrides := map[string]string{}

		if err != nil {
			return err
		}

		overrides["persistence.enabled"] = strings.ToLower(strconv.FormatBool(persistence))

		customFlags, err := command.Flags().GetStringArray("set")
		if err != nil {
			return fmt.Errorf("error with --set usage: %s", err)
		}

		if err := mergeFlags(overrides, customFlags); err != nil {
			return err
		}

		outputPath := path.Join(chartPath, "postgresql/rendered")

		err = templateChart(chartPath,
			"postgresql",
			ns,
			outputPath,
			"values.yaml",
			overrides)

		if err != nil {
			return err
		}

		err = kubectl("apply", "-R", "-f", outputPath)

		if err != nil {
			return err
		}

		fmt.Println(postgresqlInstallMsg)
		return nil
	}

	return postgresql
}

const PostgresqlInfoMsg = `PostgreSQL can be accessed via port 5432 on the following DNS name from within your cluster:

	postgresql.default.svc.cluster.local - Read/Write connection

To get the password for "postgres" run:

    export POSTGRES_PASSWORD=$(kubectl get secret --namespace default postgresql -o jsonpath="{.data.postgresql-password}" | base64 --decode)

To connect to your database run the following command:

    kubectl run postgresql-client --rm --tty -i --restart='Never' --namespace default --image docker.io/bitnami/postgresql:11.6.0-debian-9-r0 --env="PGPASSWORD=$POSTGRES_PASSWORD" --command -- psql --host postgresql -U postgres -d postgres -p 5432

To connect to your database from outside the cluster execute the following commands:

    kubectl port-forward --namespace default svc/postgresql 5432:5432 &
	PGPASSWORD="$POSTGRES_PASSWORD" psql --host 127.0.0.1 -U postgres -d postgres -p 5432

# Find out more at: https://github.com/helm/charts/tree/master/stable/postgresql`

const postgresqlInstallMsg = `=======================================================================
= PostgreSQL has been installed.                                      =
=======================================================================` +
	"\n\n" + PostgresqlInfoMsg + "\n\n" + pkg.ThanksForUsing
