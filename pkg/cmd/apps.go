package cmd

import (
	"fmt"
	"os"
	"path"
	"strings"

	"github.com/sethvargo/go-password/password"

	"github.com/alexellis/go-execute"
	"github.com/alexellis/k3sup/pkg/config"

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

func makeInstallOpenFaaS() *cobra.Command {
	var openfaas = &cobra.Command{
		Use:          "openfaas",
		Short:        "Install openfaas",
		Long:         `Install openfaas`,
		Example:      `  k3sup app install openfaas --loadbalancer`,
		SilenceUsage: true,
	}

	openfaas.Flags().Bool("loadbalancer", false, "add a loadbalancer")
	openfaas.Flags().StringP("namespace", "n", "openfaas", "Namespace for core services")

	openfaas.RunE = func(command *cobra.Command, args []string) error {
		arch := getArchitecture()
		fmt.Printf("Node architecture: %s\n", arch)

		valuesSuffix := ""
		switch arch {
		case "arm":
			valuesSuffix = "-armhf"
		}

		err := config.InitUserDir()
		if err != nil {
			return err
		}

		kubeConfigPath := path.Join(os.Getenv("HOME"), ".kube/config")

		if val, ok := os.LookupEnv("KUBECONFIG"); ok {
			kubeConfigPath = val
		}
		if command.Flags().Changed("kubeconfig") {
			kubeConfigPath, _ = command.Flags().GetString("kubeconfig")
		}
		fmt.Printf("Using context: %s\n", kubeConfigPath)

		// lb, _ := command.Flags().GetBool("loadbalancer")

		namespace, _ := command.Flags().GetString("namespace")

		err = addHelmRepo("openfaas", "https://openfaas.github.io/faas-netes/")
		if err != nil {
			return err
		}

		err = updateHelmRepos()

		if err != nil {
			return err
		}

		err = kubectl("apply", "-f", "https://raw.githubusercontent.com/openfaas/faas-netes/master/namespaces.yml")

		if err != nil {
			return err
		}

		pass, err := password.Generate(25, 10, 0, false, true)
		if err != nil {
			return err
		}

		_, err = kubectlTask("-n", namespace, "create", "secret", "generic", "basic-auth", "--from-literal=basic-auth-user=admin", `--from-literal=basic-auth-password=`+pass)

		if err != nil {
			return err
		}

		chartPath := path.Join(os.TempDir(), "charts")

		err = fetchChart(chartPath, "openfaas/openfaas")

		if err != nil {
			return err
		}

		outputPath := path.Join(chartPath, "openfaas/rendered")
		err = templateChart(chartPath, "openfaas", namespace, outputPath, "values"+valuesSuffix+".yaml")
		if err != nil {
			return err
		}

		err = kubectl("apply", "-R", "-f", outputPath)

		return err
	}

	return openfaas
}

func fetchChart(path, chart string) error {
	mkErr := os.MkdirAll(path, 0700)

	if mkErr != nil {
		return mkErr
	}
	task := execute.ExecTask{
		Command: fmt.Sprintf("helm fetch %s --untar --untardir %s", chart, path),
	}
	res, err := task.Execute()

	if err != nil {
		return err
	}

	if res.ExitCode != 0 {
		return fmt.Errorf("exit code %d", res.ExitCode)
	}
	return nil
}

func getArchitecture() string {
	res, _ := kubectlTask("get", "nodes", `--output`, `jsonpath={range $.items[0]}{.status.nodeInfo.architecture}`)

	arch := strings.TrimSpace(string(res.Stdout))

	return arch
}

func templateChart(basePath, chart, namespace, outputPath, values string) error {

	mkErr := os.MkdirAll(outputPath, 0700)
	if mkErr != nil {
		return mkErr
	}

	chartRoot := path.Join(basePath, chart)
	task := execute.ExecTask{
		Command: fmt.Sprintf("helm template %s --output-dir %s --values %s --namespace %s",
			chart, outputPath, path.Join(chartRoot, values), namespace),
		Cwd: basePath,
	}

	res, err := task.Execute()

	if err != nil {
		return err
	}

	if res.ExitCode != 0 {
		return fmt.Errorf("exit code %d", res.ExitCode)
	}
	return nil
}

func addHelmRepo(name, url string) error {
	task := execute.ExecTask{
		Command: fmt.Sprintf("helm repo add %s %s", name, url),
	}
	res, err := task.Execute()

	if err != nil {
		return err
	}

	if res.ExitCode != 0 {
		return fmt.Errorf("exit code %d", res.ExitCode)
	}
	return nil
}

func updateHelmRepos() error {
	task := execute.ExecTask{
		Command: fmt.Sprintf("helm repo update"),
	}
	res, err := task.Execute()

	if err != nil {
		return err
	}

	if res.ExitCode != 0 {
		return fmt.Errorf("exit code %d", res.ExitCode)
	}
	return nil
}

func kubectlTask(parts ...string) (execute.ExecResult, error) {
	task := execute.ExecTask{
		Command: "kubectl",
		Args:    parts,
	}

	res, err := task.Execute()

	return res, err
}

func kubectl(parts ...string) error {
	task := execute.ExecTask{
		Command: "kubectl",
		Args:    parts,
	}

	res, err := task.Execute()

	if err != nil {
		return err
	}

	if res.ExitCode != 0 {
		return fmt.Errorf("exit code %d", res.ExitCode)
	}
	return nil
}
