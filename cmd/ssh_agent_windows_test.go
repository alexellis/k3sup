//go:build windows

package cmd

import (
	"fmt"
	"net"
	"os"
	"testing"
	"time"

	"github.com/Microsoft/go-winio"
)

func TestDialWindowsPipe(t *testing.T) {
	pipePath := fmt.Sprintf(`\\.\pipe\k3sup-ssh-agent-test-%d`, os.Getpid())
	pipeListener, err := winio.ListenPipe(pipePath, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer pipeListener.Close()

	serverConnCh := make(chan net.Conn, 1)
	acceptErrCh := make(chan error, 1)
	go func() {
		serverConn, err := pipeListener.Accept()
		if err != nil {
			acceptErrCh <- err
			return
		}
		serverConnCh <- serverConn
	}()

	clientConn, err := dialWindowsPipe(pipePath)
	if err != nil {
		t.Fatal(err)
	}
	defer clientConn.Close()

	select {
	case serverConn := <-serverConnCh:
		serverConn.Close()
	case err := <-acceptErrCh:
		t.Fatal(err)
	case <-time.After(2 * time.Second):
		t.Fatal("timed out waiting for pipe connection")
	}
}
