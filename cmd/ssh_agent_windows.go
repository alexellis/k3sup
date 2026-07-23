//go:build windows

package cmd

import (
	"net"

	"github.com/Microsoft/go-winio"
)

const windowsOpenSSHAgentPipe = `\\.\pipe\openssh-ssh-agent`

func dialSystemSSHAgent() (net.Conn, error) {
	return dialWindowsPipe(windowsOpenSSHAgentPipe)
}

func dialWindowsPipe(path string) (net.Conn, error) {
	return winio.DialPipe(path, nil)
}
