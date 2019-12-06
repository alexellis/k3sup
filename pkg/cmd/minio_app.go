package cmd

import (
	"fmt"
	"log"
	"os"
	"path"

	"github.com/alexellis/k3sup/pkg/config"
	"github.com/sethvargo/go-password/password"
	"github.com/spf13/cobra"
)

func makeInstallMinio() *cobra.Command {
	var minio = &cobra.Command{
		Use:          "minio",
		Short:        "Install minio",
		Long:         `Install minio`,
		Example:      `  k3sup app install minio`,
		SilenceUsage: true,
	}

	minio.Flags().Bool("update-repo", true, "Update the helm repo")
	minio.Flags().String("access-key", "", "Provide an access key to override the pre-generated value")
	minio.Flags().String("secret-key", "", "Provide a secret key to override the pre-generated value")
	minio.Flags().Bool("distributed", false, "Deploy Minio in Distributed Mode")

	minio.RunE = func(command *cobra.Command, args []string) error {
		kubeConfigPath := getDefaultKubeconfig()

		if command.Flags().Changed("kubeconfig") {
			kubeConfigPath, _ = command.Flags().GetString("kubeconfig")
		}
		updateRepo, _ := minio.Flags().GetBool("update-repo")

		fmt.Printf("Using kubeconfig: %s\n", kubeConfigPath)

		userPath, err := config.InitUserDir()
		if err != nil {
			return err
		}

		clientArch, clientOS := getClientArch()

		fmt.Printf("Client: %s, %s\n", clientArch, clientOS)
		log.Printf("User dir established as: %s\n", userPath)

		os.Setenv("HELM_HOME", path.Join(userPath, ".helm"))

		_, err = tryDownloadHelm(userPath, clientArch, clientOS)
		if err != nil {
			return err
		}

		if updateRepo {
			err = updateHelmRepos()
			if err != nil {
				return err
			}
		}

		chartPath := path.Join(os.TempDir(), "charts")
		err = fetchChart(chartPath, "stable/minio")

		if err != nil {
			return err
		}

		overrides := map[string]string{}
		accessKey, _ := minio.Flags().GetString("access-key")
		secretKey, _ := minio.Flags().GetString("secret-key")

		if len(accessKey) == 0 {
			fmt.Printf("Access Key not provided, one will be generated for you\n")
			accessKey, err = password.Generate(20, 10, 0, false, true)
		}
		if len(secretKey) == 0 {
			fmt.Printf("Secret Key not provided, one will be generated for you\n")
			secretKey, err = password.Generate(40, 10, 5, false, true)
		}

		if err != nil {
			return err
		}

		overrides["accessKey"] = accessKey
		overrides["secretKey"] = secretKey

		if dist, _ := minio.Flags().GetBool("distributed"); dist {
			overrides["mode"] = "distributed"
		}

		outputPath := path.Join(chartPath, "minio/rendered")

		ns := "default"
		err = templateChart(chartPath,
			"minio",
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

		fmt.Println(`=======================================================================
= Minio has been installed.                                        =
=======================================================================

# Forward the minio port to your machine
kubectl port-forward -n default svc/minio 9000:9000 &

# Get the access and secret key to gain access to minio
ACCESSKEY=$(kubectl get secret -n default minio -o jsonpath="{.data.accesskey}" | base64 --decode; echo)
SECRETKEY=$(kubectl get secret -n default minio -o jsonpath="{.data.secretkey}" | base64 --decode; echo)

# Get the Minio Client
wget https://dl.min.io/client/mc/release/linux-amd64/mc \
&& chmod +x mc \
&& ./mc --help

# Find out more at:
# https://min.io

` + thanksForUsing)
		return nil
	}

	return minio
}
