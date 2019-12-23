package config

import (
	"fmt"
	"os"
	"path"
)

// K3sVersion default version
const K3sVersion = "v1.0.1"

func InitUserDir() (string, error) {
	home := os.Getenv("HOME")
	root := fmt.Sprintf("%s/.k3sup/", home)

	if len(home) == 0 {
		return home, fmt.Errorf("env-var HOME, not set")
	}

	binPath := path.Join(root, "/bin/")
	err := os.MkdirAll(binPath, 0700)
	if err != nil {
		return binPath, err
	}

	helmPath := path.Join(root, "/.helm/")
	helmErr := os.MkdirAll(helmPath, 0700)
	if helmErr != nil {
		return helmPath, helmErr
	}

	return root, nil
}
