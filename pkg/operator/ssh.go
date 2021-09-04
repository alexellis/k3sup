package ssh

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/pkg/sftp"
	"golang.org/x/crypto/ssh"
)

type SSHOperator struct {
	conn     *ssh.Client
	sftpConn *sftp.Client
}

func (s SSHOperator) Close() error {
	err := s.sftpConn.Close()
	if err != nil {
		return err
	}

	return s.conn.Close()
}

func NewSSHOperator(address string, config *ssh.ClientConfig) (*SSHOperator, error) {
	conn, err := ssh.Dial("tcp", address, config)
	if err != nil {
		return nil, err
	}

	sftpConn, err := sftp.NewClient(conn)
	if err != nil {
		return nil, err
	}

	operator := SSHOperator{
		conn:     conn,
		sftpConn: sftpConn,
	}

	return &operator, nil
}
func (s SSHOperator) ExecuteStdio(command string, stream bool) (CommandRes, error) {

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

	var stdOutWriter io.Writer
	if stream {
		stdOutWriter = io.MultiWriter(os.Stdout, &output)
	} else {
		stdOutWriter = &output
	}

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
	var stdErrWriter io.Writer
	if stream {
		stdErrWriter = io.MultiWriter(os.Stderr, &errorOutput)
	} else {
		stdErrWriter = &errorOutput
	}

	wg.Add(1)
	go func() {
		io.Copy(stdErrWriter, sessStderr)
		wg.Done()
	}()

	err = sess.Run(command)
	if err != nil {
		return CommandRes{}, err
	}

	wg.Wait()

	return CommandRes{
		StdErr: errorOutput.Bytes(),
		StdOut: output.Bytes(),
	}, nil
}

func (s SSHOperator) Execute(command string) (CommandRes, error) {
	return s.ExecuteStdio(command, true)
}

func (s SSHOperator) OpenFile(path string, flags int) (*sftp.File, error) {
	return s.sftpConn.OpenFile(path, flags)
}

func (s SSHOperator) UploadFile(localFilePath, remoteFilePath string) error {
	fmt.Printf("Uploading file %s to %s\n", localFilePath, remoteFilePath)
	srcFile, err := os.Open(localFilePath)
	if err != nil {
		return err
	}
	defer srcFile.Close()

	parent := filepath.Dir(remoteFilePath)
	path := string(filepath.Separator)
	dirs := strings.Split(parent, path)
	sshFxFailure := uint32(4)
	for _, dir := range dirs {
		path = filepath.Join(path, dir)
		err := s.sftpConn.Mkdir(path)
		if status, ok := err.(*sftp.StatusError); ok {
			if status.Code == sshFxFailure {
				var fi os.FileInfo
				fi, err = s.sftpConn.Stat(path)
				if err == nil {
					if !fi.IsDir() {
						return fmt.Errorf("file exists: %s", path)
					}
				}
			}
		}
		if err != nil {
			return err
		}
	}

	dstFile, err := s.OpenFile(remoteFilePath, (os.O_WRONLY | os.O_CREATE | os.O_TRUNC))
	if err != nil {
		return err
	}
	defer dstFile.Close()

	_, err = io.Copy(dstFile, srcFile)
	if err != nil {
		return err
	}

	return nil
}

func (s SSHOperator) Chmod(path string, mode os.FileMode) error {
	return s.sftpConn.Chmod(path, mode)
}

type CommandRes struct {
	StdOut []byte
	StdErr []byte
}

func executeCommand(cmd string) (CommandRes, error) {

	return CommandRes{}, nil
}
