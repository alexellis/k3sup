package execute

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"strings"
)

type ExecTask struct {
	Command string
	Args    []string
	Shell   bool
	Env     []string
	Cwd     string
}

type ExecResult struct {
	Stdout   string
	Stderr   string
	ExitCode int
}

func (et ExecTask) Execute() (ExecResult, error) {
	argsSt := ""
	if len(et.Args) > 0 {
		argsSt = strings.Join(et.Args, " ")
	}

	fmt.Println("exec: ", et.Command, argsSt)

	var cmd *exec.Cmd

	if et.Shell {
		var args []string
		if len(et.Args) == 0 {
			startArgs := strings.Split(et.Command, " ")
			script := strings.Join(startArgs, " ")
			args = append([]string{"-c"}, fmt.Sprintf("%s", script))

		} else {
			script := strings.Join(et.Args, " ")
			args = append([]string{"-c"}, fmt.Sprintf("%s %s", et.Command, script))

		}

		cmd = exec.Command("/bin/bash", args...)
	} else {
		if strings.Index(et.Command, " ") > 0 {
			parts := strings.Split(et.Command, " ")
			command := parts[0]
			args := parts[1:]
			cmd = exec.Command(command, args...)

		} else {
			cmd = exec.Command(et.Command, et.Args...)
		}
	}

	cmd.Dir = et.Cwd

	if len(et.Env) > 0 {
		cmd.Env = os.Environ()
		for _, env := range et.Env {
			cmd.Env = append(cmd.Env, env)
		}
	}

	stdoutPipe, stdoutPipeErr := cmd.StdoutPipe()
	if stdoutPipeErr != nil {
		return ExecResult{}, stdoutPipeErr
	}

	stderrPipe, stderrPipeErr := cmd.StderrPipe()
	if stderrPipeErr != nil {
		return ExecResult{}, stderrPipeErr
	}

	startErr := cmd.Start()

	if startErr != nil {
		return ExecResult{}, startErr
	}

	stdoutBytes, err := ioutil.ReadAll(stdoutPipe)
	if err != nil {
		return ExecResult{}, err
	}

	stderrBytes, err := ioutil.ReadAll(stderrPipe)

	if err != nil {
		return ExecResult{}, err
	}

	fmt.Println("res: " + string(stdoutBytes))

	exitCode := 0
	execErr := cmd.Wait()
	if execErr != nil {
		if exitError, ok := execErr.(*exec.ExitError); ok {

			exitCode = exitError.ExitCode()
		}
	}

	return ExecResult{
		Stdout:   string(stdoutBytes),
		Stderr:   string(stderrBytes),
		ExitCode: exitCode,
	}, nil
}
