package download

import (
	"fmt"
	"github.com/pkg/errors"
	"net/http"
	"net/url"
	"os"
	"path"
	"strings"

	execute "github.com/alexellis/go-execute/pkg/v1"
	"github.com/alexellis/k3sup/pkg/env"
)

const helmVersion = "v2.16.0"
const helm3Version = "v3.0.1"

func DownloadHelm(userPath, clientArch, clientOS string, helm3 bool) (string, error) {
	if helm3 {
		return DownloadHelmVersion(userPath, clientArch, clientOS, helm3Version, false)
	}
	return DownloadHelmVersion(userPath, clientArch, clientOS, helmVersion, false)
}

func DownloadHelmVersion(userPath, clientArch, clientOS, version string, force bool) (string, error) {
	helm3 := strings.HasPrefix(version, "v3.")
	helmVal := "helm"

	if helm3 {
		helmVal = "helm3"
	}

	helmBinaryPath := path.Join(userPath, helmVal)
	if _, statErr := os.Stat(helmBinaryPath); statErr != nil && !force {
		subdir := ""
		if helm3 {
			subdir = "helm3"
		}
		if err := tryDownloadHelm(userPath, clientArch, clientOS, subdir, version); err != nil {
			return "", err
		}

		if !helm3 {
			err := HelmInit()
			if err != nil {
				return "", err
			}
		}
	}
	return helmBinaryPath, nil
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

func tryDownloadHelm(userPath, clientArch, clientOS, version, name string) error {

	helmURL := getHelmURL(clientArch, clientOS, version)
	fmt.Println(helmURL)
	parsedURL, _ := url.Parse(helmURL)

	tmpLocation := os.TempDir()
	if err := downloadBinary(http.DefaultClient, parsedURL.String(), name, tmpLocation, true); err != nil {
		return err
	}

	err, permissionErr := moveFile(path.Join(tmpLocation, name), path.Join(userPath, name))

	if permissionErr {
		return errors.Wrap(err, "This command was run without enough permissions to move the file, please use sudo to move to this location")
	}
	if err != nil {
		return err
	}

	return nil
}

func HelmInit() error {
	fmt.Printf("Running helm init.\n")
	subdir := ""

	task := execute.ExecTask{
		Command:     fmt.Sprintf("%s", env.LocalBinary("helm", subdir)),
		Env:         os.Environ(),
		Args:        []string{"init", "--client-only"},
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
