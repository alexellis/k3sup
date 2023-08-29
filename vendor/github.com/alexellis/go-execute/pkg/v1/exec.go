package execute

import (
	"bytes"
	"fmt"
	"io"
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

	// Stdin connect a reader to stdin for the command
	// being executed.
	Stdin io.Reader

	// StreamStdio prints stdout and stderr directly to os.Stdout/err as
	// the command runs.
	StreamStdio bool

	// PrintCommand prints the command before executing
	PrintCommand bool
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

	if et.PrintCommand {
		fmt.Println("exec: ", et.Command, argsSt)
	}

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
		overrides := map[string]bool{}
		for _, env := range et.Env {
			key := strings.Split(env, "=")[0]
			overrides[key] = true
			cmd.Env = append(cmd.Env, env)
		}

		for _, env := range os.Environ() {
			key := strings.Split(env, "=")[0]

			if _, ok := overrides[key]; !ok {
				cmd.Env = append(cmd.Env, env)
			}
		}
	}
	if et.Stdin != nil {
		cmd.Stdin = et.Stdin
	}

	stdoutBuff := bytes.Buffer{}
	stderrBuff := bytes.Buffer{}

	var stdoutWriters io.Writer
	var stderrWriters io.Writer

	if et.StreamStdio {
		stdoutWriters = io.MultiWriter(os.Stdout, &stdoutBuff)
		stderrWriters = io.MultiWriter(os.Stderr, &stderrBuff)
	} else {
		stdoutWriters = &stdoutBuff
		stderrWriters = &stderrBuff
	}

	cmd.Stdout = stdoutWriters
	cmd.Stderr = stderrWriters

	startErr := cmd.Start()

	if startErr != nil {
		return ExecResult{}, startErr
	}

	exitCode := 0
	execErr := cmd.Wait()
	if execErr != nil {
		if exitError, ok := execErr.(*exec.ExitError); ok {

			exitCode = exitError.ExitCode()
		}
	}

	return ExecResult{
		Stdout:   string(stdoutBuff.Bytes()),
		Stderr:   string(stderrBuff.Bytes()),
		ExitCode: exitCode,
	}, nil
}
