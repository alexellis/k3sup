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
	"github.com/sethvargo/go-password/password"
	"github.com/spf13/cobra"
)

var clientArch, clientOS = env.GetClientArch()

func MakeInstallMinio() *cobra.Command {
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
	minio.Flags().String("namespace", "default", "Kubernetes namespace for the application")
	minio.Flags().Bool("persistence", false, "Enable persistence")
	minio.Flags().StringArray("set", []string{},
		"Use custom flags or override existing flags \n(example --set persistence.enabled=true)")

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

		fmt.Printf("Client: %s, %s\n", clientArch, clientOS)
		log.Printf("User dir established as: %s\n", userPath)

		os.Setenv("HELM_HOME", path.Join(userPath, ".helm"))

		ns, _ := minio.Flags().GetString("namespace")

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
		err = fetchChart(chartPath, "stable/minio", false)

		if err != nil {
			return err
		}

		persistence, _ := minio.Flags().GetBool("persistence")

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

		overrides["persistence.enabled"] = strings.ToLower(strconv.FormatBool(persistence))

		if dist, _ := minio.Flags().GetBool("distributed"); dist {
			overrides["mode"] = "distributed"
		}

		customFlags, err := minio.Flags().GetStringArray("set")
		if err != nil {
			return fmt.Errorf("error with --set usage: %s", err)
		}

		if err := mergeFlags(overrides, customFlags); err != nil {
			return err
		}

		outputPath := path.Join(chartPath, "minio/rendered")

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

		fmt.Println(minioInstallMsg)
		return nil
	}

	return minio
}

var MinioInfoMsg = `# Forward the minio port to your machine
kubectl port-forward -n default svc/minio 9000:9000 &

# Get the access and secret key to gain access to minio
ACCESSKEY=$(kubectl get secret -n default minio -o jsonpath="{.data.accesskey}" | base64 --decode; echo)
SECRETKEY=$(kubectl get secret -n default minio -o jsonpath="{.data.secretkey}" | base64 --decode; echo)

# Get the Minio Client
curl -SLf https://dl.min.io/client/mc/release/` + strings.ToLower(clientOS) + `-amd64/mc \
  && chmod +x mc

# Add a host
mc config host add minio http://127.0.0.1:9000 $ACCESSKEY $SECRETKEY

# List buckets
mc ls minio

# Find out more at: https://min.io`

var minioInstallMsg = `=======================================================================
= Minio has been installed.                                           =
=======================================================================` +
	"\n\n" + MinioInfoMsg + "\n\n" + pkg.ThanksForUsing
