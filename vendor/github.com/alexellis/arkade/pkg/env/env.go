// Copyright (c) arkade author(s) 2022. All rights reserved.
// Licensed under the MIT license. See LICENSE file in the project root for full license information.

package env

import (
	"context"
	"log"
	"os"
	"path"
	"strings"

	execute "github.com/alexellis/go-execute/v2"
)

// GetClientArch returns a pair of arch and os
func GetClientArch() (arch string, os string) {
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
