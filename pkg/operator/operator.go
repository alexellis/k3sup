package ssh

import (
	goexecute "github.com/alexellis/go-execute/pkg/v1"
)

type CommandOperator interface {
	Execute(command string) (CommandRes, error)
}

type ExecOperator struct {
}

func (ex ExecOperator) Execute(command string) (CommandRes, error) {

	task := goexecute.ExecTask{
		Command:     command,
		Shell:       true,
		StreamStdio: true,
	}

	res, err := task.Execute()
	if err != nil {
		return CommandRes{}, err
	}

	return CommandRes{
		StdErr: []byte(res.Stderr),
		StdOut: []byte(res.Stdout),
	}, nil

}
