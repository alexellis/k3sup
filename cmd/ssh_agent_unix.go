//go:build !windows

package cmd

import (
	"net"
	"os"
)

func dialSystemSSHAgent() (net.Conn, error) {
	return net.Dial("unix", os.Getenv("SSH_AUTH_SOCK"))
}
