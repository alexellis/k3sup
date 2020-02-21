package apps

import (
	"fmt"
	"log"
	"os"
	"path"
	"strings"

	execute "github.com/alexellis/go-execute/pkg/v1"
	"github.com/alexellis/k3sup/pkg/env"
)

func fetchChart(path, chart, version string, helm3 bool) error {
	versionStr := ""

	if len(version) > 0 {
		// Issue in helm where adding a space to the command makes it think that it's another chart of " " we want to template,
		// So we add the space before version here rather than on the command
		versionStr = " --version " + version
	}
	subdir := ""
	if helm3 {
		subdir = "helm3"
	}

	mkErr := os.MkdirAll(path, 0700)

	if mkErr != nil {
		return mkErr
	}

 println(fmt.Sprintf("%s fetch %s --untar=true --untardir %s %s", env.LocalBinary("helm", subdir), chart, path, versionStr))
	task := execute.ExecTask{
		Command:     fmt.Sprintf("%s fetch %s --untar=true --untardir %s%s", env.LocalBinary("helm", subdir), chart, path, versionStr),
		Env:         os.Environ(),
		StreamStdio: true,
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

func getNodeArchitecture() string {
	res, _ := kubectlTask("get", "nodes", `--output`, `jsonpath={range $.items[0]}{.status.nodeInfo.architecture}`)

	arch := strings.TrimSpace(string(res.Stdout))

	return arch
}

func helm3Upgrade(basePath, chart, namespace, values, version string, overrides map[string]string, wait bool) error {

	chartName := chart
	if index := strings.Index(chartName, "/"); index > -1 {
		chartName = chartName[index+1:]
	}

	chartRoot := basePath



	args := []string{"upgrade", "--install", chartName, chart, "--namespace", namespace,}
	if len(version) > 0 {
		args = append(args, "--version", version)
	}

	if wait {
		args = append(args, "--wait")
	}

	fmt.Println("VALUES", values)
	if len(values) > 0 {
		args = append(args, "--values")
		if !strings.HasPrefix(values, "/") {
			args = append(args, path.Join(chartRoot, values))
		} else {
			args = append(args, values)
		}
	}

	for k, v := range overrides {
		args = append(args, "--set")
		args = append(args, fmt.Sprintf("%s=%s", k, v))
	}

	task := execute.ExecTask{
		Command:     env.LocalBinary("helm", "helm3"),
		Args:        args,
		Env:         os.Environ(),
		Cwd:         basePath,
		StreamStdio: true,
	}

	fmt.Printf("Command: %s %s\n", task.Command, task.Args)
	res, err := task.Execute()

	if err != nil {
		return err
	}

	if res.ExitCode != 0 {
		return fmt.Errorf("exit code %d, stderr: %s", res.ExitCode, res.Stderr)
	}

	if len(res.Stderr) > 0 {
		log.Printf("stderr: %s\n", res.Stderr)
	}

	return nil
}

func templateChart(basePath, chart, namespace, outputPath, values string, overrides map[string]string) error {

	rmErr := os.RemoveAll(outputPath)

	if rmErr != nil {
		log.Printf("Error cleaning up: %s, %s\n", outputPath, rmErr.Error())
	}

	mkErr := os.MkdirAll(outputPath, 0700)
	if mkErr != nil {
		return mkErr
	}

	overridesStr := ""
	for k, v := range overrides {
		overridesStr += fmt.Sprintf(" --set %s=%s", k, v)
	}

	chartRoot := path.Join(basePath, chart)

	valuesStr := ""
	if len(values) > 0 {
		valuesStr = "--values " + path.Join(chartRoot, values)
	}

	task := execute.ExecTask{
		Command: fmt.Sprintf("%s template %s --name %s --namespace %s --output-dir %s %s %s",
			env.LocalBinary("helm", ""), chart, chart, namespace, outputPath, valuesStr, overridesStr),
		Env:         os.Environ(),
		Cwd:         basePath,
		StreamStdio: true,
	}

	res, err := task.Execute()

	if err != nil {
		return err
	}

	if res.ExitCode != 0 {
		return fmt.Errorf("exit code %d, stderr: %s", res.ExitCode, res.Stderr)
	}

	if len(res.Stderr) > 0 {
		log.Printf("stderr: %s\n", res.Stderr)
	}

	return nil
}

func addHelmRepo(name, url string, helm3 bool) error {
	subdir := ""
	if helm3 {
		subdir = "helm3"
	}

	task := execute.ExecTask{
		Command:     fmt.Sprintf("%s repo add %s %s", env.LocalBinary("helm", subdir), name, url),
		Env:         os.Environ(),
		StreamStdio: true,
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

func updateHelmRepos(helm3 bool) error {
	subdir := ""
	if helm3 {
		subdir = "helm3"
	}
	task := execute.ExecTask{
		Command:     fmt.Sprintf("%s repo update", env.LocalBinary("helm", subdir)),
		Env:         os.Environ(),
		StreamStdio: true,
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
		Command:     "kubectl",
		Args:        parts,
		StreamStdio: false,
	}

	res, err := task.Execute()

	return res, err
}

func kubectl(parts ...string) error {
	task := execute.ExecTask{
		Command:     "kubectl",
		Args:        parts,
		StreamStdio: true,
	}

	res, err := task.Execute()

	if err != nil {
		return err
	}

	if res.ExitCode != 0 {
		return fmt.Errorf("kubectl exit code %d, stderr: %s",
			res.ExitCode,
			res.Stderr)
	}
	return nil
}

func getDefaultKubeconfig() string {
	kubeConfigPath := path.Join(os.Getenv("HOME"), ".kube/config")

	if val, ok := os.LookupEnv("KUBECONFIG"); ok {
		kubeConfigPath = val
	}

	return kubeConfigPath
}
