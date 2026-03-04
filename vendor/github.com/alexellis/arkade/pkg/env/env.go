// Copyright (c) arkade author(s) 2022. All rights reserved.
// Licensed under the MIT license. See LICENSE file in the project root for full license information.

package env

import (
	"context"
	"log"
	"os"
	"path"
	"runtime"
	"strings"

	execute "github.com/alexellis/go-execute/v2"
)

// GetClientArch returns the architecture and OS of the client machine
// on Windows, that's a direct passthrough from runtime.GOARCH and runtime.GOOS
// On Linux and Darwin, it uses `uname -m` and `uname -s` to get more specific values
func GetClientArch() (arch string, os string) {
	if runtime.GOOS == "windows" {
		return runtime.GOARCH, runtime.GOOS
	}

	return getClientArchFromUname()
}

func getClientArchFromUname() (arch string, os string) {
	task := execute.ExecTask{
		Command:     "uname",
		Args:        []string{"-m"},
		StreamStdio: false}
	res, err := task.Execute(context.Background())
	if err != nil {
		log.Println(err)
	}

	archResult := strings.TrimSpace(res.Stdout)

	taskOS := execute.ExecTask{Command: "uname",
		Args:        []string{"-s"},
		StreamStdio: false}
	resOS, errOS := taskOS.Execute(context.Background())
	if errOS != nil {
		log.Println(errOS)
	}

	osResult := strings.TrimSpace(resOS.Stdout)

	return archResult, osResult
}

func LocalBinary(name, subdir string) string {
	home := os.Getenv("HOME")
	val := path.Join(home, ".arkade/bin/")
	if len(subdir) > 0 {
		val = path.Join(val, subdir)
	}

	return path.Join(val, name)
}
