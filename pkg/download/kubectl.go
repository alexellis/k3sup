package download

import (
	"fmt"
	"github.com/pkg/errors"
	"net/http"
	"os"
	"path"
)

func DownloadKubectl(version, clientOs, clientArch, outputLocation string) error {
	downloadURL := fmt.Sprintf("https://storage.googleapis.com/kubernetes-release/release/%s/bin/%s/%s/kubectl",
		version, clientOs, clientArch)

	tmpLocation := os.TempDir()
	if err := downloadBinary(http.DefaultClient, downloadURL, "kubectl", tmpLocation, false); err != nil {
		return err
	}

	err, permissionErr := moveFile(path.Join(tmpLocation, "kubectl"), outputLocation)
	if permissionErr {
		return errors.Wrap(err, "This command was run without enough permissions to move the file, please use sudo to move to this location")
	}

	if err != nil {
		return err
	}

	return nil

}
