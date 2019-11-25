package ssh

import (
	"bytes"
	"io"
	"os"
	"sync"

	"golang.org/x/crypto/ssh"
)

type SSHOperator struct {
	conn *ssh.Client
}

func (s SSHOperator) Close() error {

	return s.conn.Close()
}

func NewSSHOperator(address string, config *ssh.ClientConfig) (*SSHOperator, error) {
	conn, err := ssh.Dial("tcp", address, config)
	if err != nil {
		return nil, err
	}

	operator := SSHOperator{
		conn: conn,
	}

	return &operator, nil
}

func (s SSHOperator) Execute(command string) (CommandRes, error) {

	sess, err := s.conn.NewSession()
	if err != nil {
		return CommandRes{}, err
	}

	defer sess.Close()

	sessStdOut, err := sess.StdoutPipe()
	if err != nil {
		return CommandRes{}, err
	}

	output := bytes.Buffer{}

	wg := sync.WaitGroup{}

	stdOutWriter := io.MultiWriter(os.Stdout, &output)
	wg.Add(1)
	go func() {
		io.Copy(stdOutWriter, sessStdOut)
		wg.Done()
	}()
	sessStderr, err := sess.StderrPipe()
	if err != nil {
		return CommandRes{}, err
	}

	errorOutput := bytes.Buffer{}
	stdErrWriter := io.MultiWriter(os.Stderr, &errorOutput)
	wg.Add(1)
	go func() {
		io.Copy(stdErrWriter, sessStderr)
		wg.Done()
	}()

	err = sess.Run(command)

	wg.Wait()

	if err != nil {
		return CommandRes{}, err
	}

	return CommandRes{
		StdErr: errorOutput.Bytes(),
		StdOut: output.Bytes(),
	}, nil
}

type CommandRes struct {
	StdOut []byte
	StdErr []byte
}

func executeCommand(cmd string) (CommandRes, error) {

	return CommandRes{}, nil
}
