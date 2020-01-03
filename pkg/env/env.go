package env

import (
	"log"
	"strings"

	execute "github.com/alexellis/go-execute/pkg/v1"
)

// GetClientArch returns a pair of arch and os
func GetClientArch() (string, string) {
	task := execute.ExecTask{Command: "uname", Args: []string{"-m"}, StreamStdio: true}
	res, err := task.Execute()
	if err != nil {
		log.Println(err)
	}

	arch := strings.TrimSpace(res.Stdout)

	taskOS := execute.ExecTask{Command: "uname", Args: []string{"-s"}, StreamStdio: true}
	resOS, errOS := taskOS.Execute()
	if errOS != nil {
		log.Println(errOS)
	}

	os := strings.TrimSpace(resOS.Stdout)

	return arch, os
}
