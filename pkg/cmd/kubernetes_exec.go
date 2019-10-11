package cmd

import (
	"fmt"
	"os"
	"path"
	"strings"

	"github.com/alexellis/go-execute"
)

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
