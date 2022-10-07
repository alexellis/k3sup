package ssh

// CommandOperator executes a command on a machine to install k3sup
type CommandOperator interface {
	Execute(command string) (CommandRes, error)
	ExecuteStdio(command string, stream bool) (CommandRes, error)
}

// CommandRes contains the STDIO output from running a command
type CommandRes struct {
	StdOut   []byte
	StdErr   []byte
	ExitCode int
}
