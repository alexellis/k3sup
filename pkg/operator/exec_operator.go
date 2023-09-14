package ssh

import (
	"context"

	goexecute "github.com/alexellis/go-execute/v2"
)

// ExecOperator executes commands on the local machine directly
type ExecOperator struct {
}

func (ex ExecOperator) ExecuteStdio(command string, stream bool) (CommandRes, error) {
	task := goexecute.ExecTask{
		Command:     command,
		Shell:       true,
		StreamStdio: stream,
	}

	res, err := task.Execute(context.Background())
	if err != nil {
		return CommandRes{}, err
	}

	return CommandRes{
		StdErr:   []byte(res.Stderr),
		StdOut:   []byte(res.Stdout),
		ExitCode: res.ExitCode,
	}, nil
}

func (ex ExecOperator) Execute(command string) (CommandRes, error) {
	return ex.ExecuteStdio(command, true)
}
