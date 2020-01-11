package cmd

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path"

	"github.com/alexellis/k3sup/pkg/config"
	"github.com/alexellis/k3sup/pkg/env"

	"github.com/spf13/cobra"
)

func makeInstallIstio() *cobra.Command {
	var istio = &cobra.Command{
		Use:          "istio",
		Short:        "Install istio",
		Long:         `Install istio`,
		Example:      `  k3sup app install istio --loadbalancer`,
		SilenceUsage: true,
	}
	istio.Flags().Bool("update-repo", true, "Update the helm repo")
	istio.Flags().String("namespace", "istio-system", "Namespace for the app")
	istio.Flags().Bool("init", true, "Run the Istio init to add CRDs etc")

	istio.Flags().StringArray("set", []string{},
		"Use custom flags or override existing flags \n(example --set=prometheus.enabled=false)")

	istio.RunE = func(command *cobra.Command, args []string) error {
		kubeConfigPath := getDefaultKubeconfig()

		if command.Flags().Changed("kubeconfig") {
			kubeConfigPath, _ = command.Flags().GetString("kubeconfig")
		}

		fmt.Printf("Using kubeconfig: %s\n", kubeConfigPath)

		namespace, _ := command.Flags().GetString("namespace")

		if namespace != "istio-system" {
			return fmt.Errorf(`to override the "istio-system" namespace, install Istio via helm manually`)
		}

		arch := getNodeArchitecture()
		fmt.Printf("Node architecture: %q\n", arch)

		userPath, err := config.InitUserDir()
		if err != nil {
			return err
		}

		clientArch, clientOS := env.GetClientArch()

		fmt.Printf("Client: %q, %q\n", clientArch, clientOS)

		log.Printf("User dir established as: %s\n", userPath)

		os.Setenv("HELM_HOME", path.Join(userPath, ".helm"))

		helm3 := true
		if helm3 {
			helm3Version := "v3.0.1"
			os.Setenv("HELM_VERSION", helm3Version)
		}

		_, err = tryDownloadHelm(userPath, clientArch, clientOS, helm3)
		if err != nil {
			return err
		}

		istioVer := "1.3.3"

		err = addHelmRepo("istio", "https://storage.googleapis.com/istio-release/releases/"+istioVer+"/charts", helm3)
		if err != nil {
			return fmt.Errorf("unable to add repo %s", err)
		}

		updateRepo, _ := istio.Flags().GetBool("update-repo")

		if updateRepo {
			err = updateHelmRepos(helm3)
			if err != nil {
				return fmt.Errorf("unable to update repos %s", err)
			}
		}

		_, err = kubectlTask("create", "ns", "istio-system")

		if err != nil {
			return fmt.Errorf("unable to create namespace %s", err)
		}

		chartPath := path.Join(os.TempDir(), "charts")

		err = fetchChart(chartPath, "istio/istio", helm3)

		if err != nil {
			return fmt.Errorf("unable fetch chart %s", err)
		}

		overrides := map[string]string{}

		valuesFile, writeErr := writeIstioValues()
		if writeErr != nil {
			return writeErr
		}

		outputPath := path.Join(chartPath, "istio")

		wait := true

		if initIstio, _ := command.Flags().GetBool("init"); initIstio {
			err = helm3Upgrade(outputPath, "istio/istio-init", namespace, "", overrides, wait)
			if err != nil {
				return fmt.Errorf("unable to istio-init install chart with helm %s", err)
			}

		}

		customFlags, customFlagErr := command.Flags().GetStringArray("set")
		if customFlagErr != nil {
			return fmt.Errorf("error with --set usage: %s", customFlagErr)
		}

		if err := mergeFlags(overrides, customFlags); err != nil {
			return err
		}

		err = helm3Upgrade(outputPath, "istio/istio", namespace, valuesFile, overrides, wait)
		if err != nil {
			return fmt.Errorf("unable to istio install chart with helm %s", err)
		}

		fmt.Println(istioPostInstallMsg)

		return nil
	}

	return istio
}

const istioInfoMsg = `# Find out more at:
# https://github.com/istio/`

const istioPostInstallMsg = `=======================================================================
= Istio has been installed.                                        =
=======================================================================` +
	"\n\n" + istioInfoMsg + "\n\n" + thanksForUsing

func writeIstioValues() (string, error) {
	out := `#
# Minimal Istio Configuration taken from https://github.com/weaveworks/flagger

# pilot configuration
pilot:
  enabled: true
  sidecar: true
  resources:
    requests:
      cpu: 10m
      memory: 128Mi

gateways:
  enabled: true
  istio-ingressgateway:
    autoscaleMax: 1

# sidecar-injector webhook configuration
sidecarInjectorWebhook:
  enabled: true

# galley configuration
galley:
  enabled: false

# mixer configuration
mixer:
  policy:
    enabled: false
  telemetry:
    enabled: true
    replicaCount: 1
    autoscaleEnabled: false
  resources:
    requests:
      cpu: 10m
      memory: 128Mi

# addon prometheus configuration
prometheus:
  enabled: true
  scrapeInterval: 5s

# addon jaeger tracing configuration
tracing:
  enabled: false

# Common settings.
global:
  proxy:
    # Resources for the sidecar.
    resources:
      requests:
        cpu: 10m
        memory: 64Mi
      limits:
        cpu: 1000m
        memory: 256Mi
  useMCP: false`

	writeTo := path.Join(os.TempDir(), "istio-values.yaml")
	return writeTo, ioutil.WriteFile(writeTo, []byte(out), 0600)
}
