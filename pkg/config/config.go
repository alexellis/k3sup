package config

import (
	"fmt"
	"os"
)

// K3sVersion default version
const K3sVersion = "v0.8.1"

func InitUserDir() (string, error) {
	home := os.Getenv("HOME")
	fullPath := fmt.Sprintf("%s/k3sup/.bin/", home)

	if len(home) == 0 {
		return fullPath, fmt.Errorf("env-var HOME, not set")
	}

	err := os.MkdirAll(fullPath, 0700)
	if err != nil {
		return fullPath, err
	}

	return fullPath, nil
}
