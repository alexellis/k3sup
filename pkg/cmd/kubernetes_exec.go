package cmd

import (
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"path"
	"strings"

	execute "github.com/alexellis/go-execute/pkg/v1"
)

const helmVersion = "v2.15.2"

func fetchChart(path, chart string) error {
	mkErr := os.MkdirAll(path, 0700)

	if mkErr != nil {
		return mkErr
	}

	task := execute.ExecTask{
		Command: fmt.Sprintf("%s fetch %s --untar --untardir %s", localBinary("helm"), chart, path),
		Env:     os.Environ(),
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
			localBinary("helm"), chart, chart, namespace, outputPath, valuesStr, overridesStr),
		Env: os.Environ(),
		Cwd: basePath,
	}

	res, err := task.Execute()

	if err != nil {
		return err
	}

	if res.ExitCode != 0 {
		return fmt.Errorf("exit code %d", res.ExitCode)
	}

	if len(res.Stderr) > 0 {
		log.Printf("stderr: %s\n", res.Stderr)
	}

	return nil
}

func localBinary(name string) string {
	home := os.Getenv("HOME")
	return path.Join(path.Join(home, ".k3sup/.bin/"), name)
}

func addHelmRepo(name, url string) error {
	task := execute.ExecTask{
		Command: fmt.Sprintf("%s repo add %s %s", localBinary("helm"), name, url),
		Env:     os.Environ(),
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
		Command: fmt.Sprintf("%s repo update", localBinary("helm")),
		Env:     os.Environ(),
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

func helmInit() error {
	task := execute.ExecTask{
		Command: fmt.Sprintf("%s", localBinary("helm")),
		Env:     os.Environ(),
		Args:    []string{"init", "--client-only"},
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

func getDefaultKubeconfig() string {
	kubeConfigPath := path.Join(os.Getenv("HOME"), ".kube/config")

	if val, ok := os.LookupEnv("KUBECONFIG"); ok {
		kubeConfigPath = val
	}

	return kubeConfigPath
}

func tryDownloadHelm(userPath, clientArch, clientOS string) (string, error) {
	helmBinaryPath := path.Join(path.Join(userPath, ".bin"), "helm")
	if _, statErr := os.Stat(helmBinaryPath); statErr != nil {
		downloadHelm(userPath, clientArch, clientOS)

		err := helmInit()
		if err != nil {
			return "", err
		}
	}
	return helmBinaryPath, nil
}

// getClientArch returns a pair of arch and os
func getClientArch() (string, string) {
	task := execute.ExecTask{Command: "uname", Args: []string{"-m"}}
	res, err := task.Execute()
	if err != nil {
		log.Println(err)
	}

	arch := strings.TrimSpace(res.Stdout)

	taskOS := execute.ExecTask{Command: "uname", Args: []string{"-s"}}
	resOS, errOS := taskOS.Execute()
	if errOS != nil {
		log.Println(errOS)
	}

	os := strings.TrimSpace(resOS.Stdout)

	return arch, os
}

func getHelmURL(arch, os, version string) string {
	archSuffix := "amd64"
	osSuffix := strings.ToLower(os)

	if strings.HasPrefix(arch, "armv7") {
		archSuffix = "arm"
	} else if strings.HasPrefix(arch, "aarch64") {
		archSuffix = "arm64"
	}

	return fmt.Sprintf("https://get.helm.sh/helm-%s-%s-%s.tar.gz", version, osSuffix, archSuffix)
}

func downloadHelm(userPath, clientArch, clientOS string) error {
	helmURL := getHelmURL(clientArch, clientOS, helmVersion)
	fmt.Println(helmURL)
	parsedURL, _ := url.Parse(helmURL)

	res, err := http.DefaultClient.Get(parsedURL.String())
	if err != nil {
		return err
	}

	defer res.Body.Close()
	r := ioutil.NopCloser(res.Body)
	untarErr := Untar(r, path.Join(userPath, ".bin"))
	if untarErr != nil {
		return untarErr
	}

	return nil
}

func getLinkerdUrl(os, version string) string {
	osSuffix := strings.ToLower(os)
	return fmt.Sprintf("https://github.com/linkerd/linkerd2/releases/download/%s/linkerd2-cli-%s-%s", version, version, osSuffix)
}

func downloadLinkerd(userPath, clientOS string) error {
	filePath := path.Join(path.Join(userPath, ".bin"), "linkerd")
	if _, statErr := os.Stat(filePath); statErr != nil {
		linkerdUrl := getLinkerdUrl(clientOS, "stable-2.6.0")
		fmt.Println(linkerdUrl)
		parsedURL, _ := url.Parse(linkerdUrl)

		res, err := http.DefaultClient.Get(parsedURL.String())
		if err != nil {
			return err
		}

		defer res.Body.Close()
		out, err := os.Create(filePath)
		if err != nil {
			return err
		}
		defer out.Close()

		// Write the body to file
		_, err = io.Copy(out, res.Body)

		err = os.Chmod(filePath, 0755)
		if err != nil {
			return err
		}
	}
	return nil
}

func linkerdCli(parts ...string) (execute.ExecResult, error) {
	task := execute.ExecTask{
		Command: fmt.Sprintf("%s", localBinary("linkerd")),
		Args:    parts,
		Env:     os.Environ(),
	}
	res, err := task.Execute()

	if err != nil {
		return res, err
	}

	if res.ExitCode != 0 {
		return res, fmt.Errorf("exit code %d", res.ExitCode)
	}

	return res, nil
}
