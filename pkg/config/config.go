package config

import (
	"fmt"
	"os"
)

// K3sVersion default version
const K3sVersion = "v0.8.1"

func InitUserDir() error {
	home := os.Getenv("HOME")
	if len(home) == 0 {
		return fmt.Errorf("env-var HOME, not set")
	}

	err := os.MkdirAll(fmt.Sprintf("%s/k3sup/.bin/", home), 0700)
	if err != nil {
		return err
	}

	return nil
}
