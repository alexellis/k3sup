package download

import (
	"fmt"
	"github.com/pkg/errors"
	"net/http"
	"os"
	"path"
)

func FaaSCLI(version, outputLocation, arch, osType string) error {
	var err error
	if len(version) == 0 {
		version, err = findGithubRelease("https://github.com/openfaas/faas-cli/releases/latest")
		if err != nil {
			return err
		}
	}

	suffix, ext := buildFilename(arch, osType)
	url := fmt.Sprintf("https://github.com/openfaas/faas-cli/releases/download/%s/faas-cli%s%s", version, suffix, ext)

	tmpLocation := os.TempDir()
	if err := downloadBinary(http.DefaultClient, url, "faas-cli", tmpLocation, false); err != nil {
		return err
	}

	err, permissionErr := moveFile(path.Join(tmpLocation, "faas-cli"), outputLocation)
	if permissionErr {
		return errors.Wrap(err, "This command was run without enough permissions to move the file, please use sudo to move to this location")
	}

	if err != nil {
		return err
	}

	return nil
}
