package cmd

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"path"
	"strings"

	"github.com/spf13/cobra"
)

func MakeApps() *cobra.Command {
	var command = &cobra.Command{
		Use:          "app",
		Short:        "Manage Kubernetes apps",
		Long:         `Manage Kubernetes apps`,
		Example:      `  k3sup app install openfaas`,
		SilenceUsage: false,
	}

	var install = &cobra.Command{
		Use:          "install",
		Short:        "Install a Kubernetes app",
		Long:         `Install a Kubernetes app`,
		Example:      `  k3sup app install [APP]`,
		SilenceUsage: true,
	}

	install.Flags().String("kubeconfig", "kubeconfig", "Local path to save the kubeconfig file")

	openfaas := makeInstallOpenFaaS()

	install.RunE = func(command *cobra.Command, args []string) error {

		if len(args) == 0 {
			fmt.Printf("You can install: %s\n", strings.TrimRight(strings.Join(getApps(), ", "), ", "))
			return nil
		}

		return nil
	}

	command.AddCommand(install)
	install.AddCommand(openfaas)

	return command
}

func getApps() []string {
	return []string{"openfaas"}
}

func installOpenFaaS(kubeconfigPath string, loadBalancer bool) error {

	res, err := http.Get("https://raw.githubusercontent.com/openfaas/faas-netes/master/install.sh")
	if err != nil {
		return err
	}
	defer res.Body.Close()

	out, _ := ioutil.ReadAll(res.Body)

	val := string(out)
	if !loadBalancer {
		val = strings.Replace(val, "LoadBalancer", "NodePort", -1)
	}

	script := path.Join(os.TempDir(), "install.sh")

	err = ioutil.WriteFile(script, []byte(val), 0600)
	if err != nil {
		return err
	}

	fmt.Printf("Wrote script to: %s\n", script)

	cmd1 := exec.Command("/bin/bash", script)
	cmd1.Env = append(os.Environ(), "KUBECONFIG="+kubeconfigPath)
	cmd1.Stderr = os.Stderr
	cmd1.Stdout = os.Stdout

	startErr := cmd1.Start()
	if startErr != nil {
		return startErr
	}

	cmd1.Wait()

	return nil
}

func makeInstallOpenFaaS() *cobra.Command {
	var openfaas = &cobra.Command{
		Use:          "openfaas",
		Short:        "Install openfaas",
		Long:         `Install openfaas`,
		Example:      `  k3sup app install openfaas --loadbalancer`,
		SilenceUsage: true,
	}

	openfaas.Flags().Bool("loadbalancer", false, "add a loadbalancer")

	openfaas.RunE = func(command *cobra.Command, args []string) error {
		kubeConfigPath := path.Join(os.Getenv("HOME"), ".kube/config")

		if val, ok := os.LookupEnv("KUBECONFIG"); ok {
			kubeConfigPath = val
		}
		if command.Flags().Changed("kubeconfig") {
			kubeConfigPath, _ = command.Flags().GetString("kubeconfig")
		}
		fmt.Printf("Using context: %s\n", kubeConfigPath)

		lb, _ := command.Flags().GetBool("loadbalancer")
		err := installOpenFaaS(kubeConfigPath, lb)

		if err != nil {
			return err
		}

		return nil
	}

	return openfaas
}
