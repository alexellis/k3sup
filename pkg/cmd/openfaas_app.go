package cmd

import (
	"fmt"
	"log"
	"os"
	"path"
	"strconv"
	"strings"

	"github.com/sethvargo/go-password/password"

	"github.com/alexellis/k3sup/pkg/env"
	"github.com/alexellis/k3sup/pkg/config"

	"github.com/spf13/cobra"
)

func makeInstallOpenFaaS() *cobra.Command {
	var openfaas = &cobra.Command{
		Use:          "openfaas",
		Short:        "Install openfaas",
		Long:         `Install openfaas`,
		Example:      `  k3sup app install openfaas --loadbalancer`,
		SilenceUsage: true,
	}

	openfaas.Flags().BoolP("basic-auth", "a", true, "Enable authentication")
	openfaas.Flags().BoolP("load-balancer", "l", false, "Add a loadbalancer")
	openfaas.Flags().StringP("namespace", "n", "openfaas", "The namespace for the core services")
	openfaas.Flags().Bool("update-repo", true, "Update the helm repo")
	openfaas.Flags().String("pull-policy", "IfNotPresent", "Pull policy for OpenFaaS core services")
	openfaas.Flags().String("function-pull-policy", "Always", "Pull policy for functions")

	openfaas.Flags().Bool("operator", false, "Create OpenFaaS Operator")
	openfaas.Flags().Bool("clusterrole", false, "Create a ClusterRole for OpenFaaS instead of a limited scope Role")
	openfaas.Flags().Bool("direct-functions", true, "Invoke functions directly from the gateway")

	openfaas.Flags().Int("queue-workers", 1, "Replicas of queue-worker")
	openfaas.Flags().Int("gateways", 1, "Replicas of gateway")

	openfaas.Flags().Bool("helm3", false, "Use helm3 instead of the default helm2")

	openfaas.RunE = func(command *cobra.Command, args []string) error {
		kubeConfigPath := getDefaultKubeconfig()

		if command.Flags().Changed("kubeconfig") {
			kubeConfigPath, _ = command.Flags().GetString("kubeconfig")
		}

		fmt.Printf("Using kubeconfig: %s\n", kubeConfigPath)

		helm3, _ := command.Flags().GetBool("helm3")

		if helm3 {
			fmt.Println("Using helm3")
		}
		namespace, _ := command.Flags().GetString("namespace")

		if namespace != "openfaas" {
			return fmt.Errorf(`to override the "openfaas", install OpenFaaS via helm manually`)
		}

		arch := getNodeArchitecture()
		fmt.Printf("Node architecture: %q\n", arch)

		valuesSuffix := getValuesSuffix(arch)

		userPath, err := config.InitUserDir()
		if err != nil {
			return err
		}

		clientArch, clientOS := env.GetClientArch()

		fmt.Printf("Client: %q, %q\n", clientArch, clientOS)

		log.Printf("User dir established as: %s\n", userPath)

		os.Setenv("HELM_HOME", path.Join(userPath, ".helm"))

		if helm3 {
			helm3Version := "v3.0.1"
			os.Setenv("HELM_VERSION", helm3Version)
		}

		_, err = tryDownloadHelm(userPath, clientArch, clientOS, helm3)
		if err != nil {
			return err
		}

		err = addHelmRepo("openfaas", "https://openfaas.github.io/faas-netes/", helm3)
		if err != nil {
			return err
		}

		updateRepo, _ := openfaas.Flags().GetBool("update-repo")

		if updateRepo {
			err = updateHelmRepos(helm3)
			if err != nil {
				return err
			}
		}

		if err != nil {
			return err
		}

		_, err = kubectlTask("apply", "-f",
			"https://raw.githubusercontent.com/openfaas/faas-netes/master/namespaces.yml")

		if err != nil {
			return err
		}

		pass, err := password.Generate(25, 10, 0, false, true)
		if err != nil {
			return err
		}

		res, secretErr := kubectlTask("-n", namespace, "create", "secret", "generic",
			"basic-auth",
			"--from-literal=basic-auth-user=admin",
			`--from-literal=basic-auth-password=`+pass)

		if secretErr != nil {
			return secretErr
		}

		if res.ExitCode != 0 {
			fmt.Printf("[Warning] unable to create secret %s, may already exist: %s", "basic-auth", res.Stderr)
		}

		chartPath := path.Join(os.TempDir(), "charts")

		err = fetchChart(chartPath, "openfaas/openfaas", helm3)

		if err != nil {
			return err
		}

		overrides := map[string]string{}

		pullPolicy, _ := command.Flags().GetString("pull-policy")
		if len(pullPolicy) == 0 {
			return fmt.Errorf("you must give a value for pull-policy such as IfNotPresent or Always")
		}

		functionPullPolicy, _ := command.Flags().GetString("function-pull-policy")
		if len(pullPolicy) == 0 {
			return fmt.Errorf("you must give a value for function-pull-policy such as IfNotPresent or Always")
		}

		createOperator, _ := command.Flags().GetBool("operator")
		createOperatorVal := "false"
		if createOperator {
			createOperatorVal = "true"
		}

		clusterRole, _ := command.Flags().GetBool("clusterrole")

		clusterRoleVal := "false"
		if clusterRole {
			clusterRoleVal = "true"
		}

		directFunctions, _ := command.Flags().GetBool("direct-functions")
		directFunctionsVal := "true"
		if !directFunctions {
			directFunctionsVal = "false"
		}
		gateways, _ := command.Flags().GetInt("gateways")
		queueWorkers, _ := command.Flags().GetInt("queue-workers")

		overrides["clusterRole"] = clusterRoleVal
		overrides["gateway.directFunctions"] = directFunctionsVal
		overrides["operator.create"] = createOperatorVal
		overrides["openfaasImagePullPolicy"] = pullPolicy
		overrides["faasnetes.imagePullPolicy"] = functionPullPolicy
		overrides["basicAuthPlugin.replicas"] = "1"
		overrides["gateway.replicas"] = fmt.Sprintf("%d", gateways)
		overrides["queueWorker.replicas"] = fmt.Sprintf("%d", queueWorkers)

		basicAuth, _ := command.Flags().GetBool("basic-auth")

		// the value in the template is "basic_auth" not the more usual basicAuth
		overrides["basic_auth"] = strings.ToLower(strconv.FormatBool(basicAuth))

		overrides["serviceType"] = "NodePort"
		lb, _ := command.Flags().GetBool("load-balancer")
		if lb {
			overrides["serviceType"] = "LoadBalancer"
		}

		if helm3 {
			outputPath := path.Join(chartPath, "openfaas")

			err := helm3Upgrade(outputPath, "openfaas/openfaas", namespace, "values"+valuesSuffix+".yaml", overrides)
			if err != nil {
				return err
			}

		} else {
			outputPath := path.Join(chartPath, "openfaas/rendered")
			err = templateChart(chartPath, "openfaas",
				namespace,
				outputPath,
				"values"+valuesSuffix+".yaml",
				overrides)

			if err != nil {
				return err
			}

			applyRes, applyErr := kubectlTask("apply", "-R", "-f", outputPath)
			if applyErr != nil {
				return applyErr
			}

			if applyRes.ExitCode > 0 {
				return fmt.Errorf("Error applying templated YAML files, error: %s", applyRes.Stderr)
			}
		}

		fmt.Println(`=======================================================================
= OpenFaaS has been installed.                                        =
=======================================================================

# Get the faas-cli
curl -SLsf https://cli.openfaas.com | sudo sh

# Forward the gateway to your machine
kubectl rollout status -n openfaas deploy/gateway
kubectl port-forward -n openfaas svc/gateway 8080:8080 &

# If basic auth is enabled, you can now log into your gateway:
PASSWORD=$(kubectl get secret -n openfaas basic-auth -o jsonpath="{.data.basic-auth-password}" | base64 --decode; echo)
echo -n $PASSWORD | faas-cli login --username admin --password-stdin

faas-cli store deploy figlet
faas-cli list

# For Raspberry Pi
faas-cli store list \
 --platform armhf

faas-cli store deploy figlet \
 --platform armhf

# Find out more at:
# https://github.com/openfaas/faas

` + thanksForUsing)

		return nil
	}

	return openfaas
}

func getValuesSuffix(arch string) string {
	var valuesSuffix string
	switch arch {
	case "arm":
		valuesSuffix = "-armhf"
		break
	case "arm64", "aarch64":
		valuesSuffix = "-arm64"
		break
	default:
		valuesSuffix = ""
	}
	return valuesSuffix
}
