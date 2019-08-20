package cmd

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net"
	"os"
	"path/filepath"
	"strings"

	kssh "github.com/alexellis/k3sup/pkg/ssh"

	homedir "github.com/mitchellh/go-homedir"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/agent"
	"golang.org/x/crypto/ssh/terminal"
)

func MakeInstall() *cobra.Command {
	var command = &cobra.Command{
		Use:          "install",
		Short:        "Install k3s on a server via SSH",
		Long:         `Install k3s on a server via SSH.`,
		Example:      `  k3sup install --ip 192.168.0.100 --user root`,
		SilenceUsage: true,
	}

	command.Flags().IP("ip", nil, "Public IP of node")
	command.Flags().String("user", "root", "Username for SSH login")

	command.Flags().String("ssh-key", "~/.ssh/id_rsa", "The ssh key to use for remote login")
	command.Flags().Int("ssh-port", 22, "The port on which to connect for ssh")
	command.Flags().Bool("skip-install", false, "Skip the k3s installer")
	command.Flags().String("local-path", "kubeconfig", "Local path to save the kubeconfig file")

	command.RunE = func(command *cobra.Command, args []string) error {

		localKubeconfig, _ := command.Flags().GetString("local-path")

		skipInstall, _ := command.Flags().GetBool("skip-install")

		port, _ := command.Flags().GetInt("ssh-port")

		ip, _ := command.Flags().GetIP("ip")
		fmt.Println("Public IP: " + ip.String())

		user, _ := command.Flags().GetString("user")
		sshKey, _ := command.Flags().GetString("ssh-key")

		sshKeyPath := expandPath(sshKey)
		fmt.Printf("ssh -i %s %s@%s\n", sshKeyPath, user, ip.String())

		authMethod, closeSSHAgent, err := loadPublickey(sshKeyPath)
		if err != nil {
			return errors.Wrapf(err, "unable to load the ssh key with path %q", sshKeyPath)
		}

		defer closeSSHAgent()

		config := &ssh.ClientConfig{
			User: user,
			Auth: []ssh.AuthMethod{
				authMethod,
			},
			HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		}

		address := fmt.Sprintf("%s:%d", ip.String(), port)
		operator, err := kssh.NewSSHOperator(address, config)

		if err != nil {
			return errors.Wrapf(err, "unable to connect to %s over ssh", address)
		}

		defer operator.Close()

		if !skipInstall {
			installK3scommand := fmt.Sprintf("curl -sLS https://get.k3s.io | INSTALL_K3S_EXEC='server --tls-san %s' sh -\n", ip)
			fmt.Printf("ssh: %s\n", installK3scommand)
			res, err := operator.Execute(installK3scommand)

			if err != nil {
				return fmt.Errorf("Error received processing command: %s", err)
			}

			fmt.Printf("Result: %s %s\n", string(res.StdOut), string(res.StdErr))
		}

		getConfigcommand := fmt.Sprintf("sudo cat /etc/rancher/k3s/k3s.yaml\n")
		fmt.Printf("ssh: %s\n", getConfigcommand)

		res, err := operator.Execute(getConfigcommand)

		if err != nil {
			return fmt.Errorf("Error received processing command: %s", err)
		}

		fmt.Printf("Result: %s %s\n", string(res.StdOut), string(res.StdErr))

		absPath, _ := filepath.Abs(localKubeconfig)
		fmt.Printf("Saving file to: %s\n", absPath)

		kubeconfig := strings.Replace(string(res.StdOut), "localhost", ip.String(), -1)

		writeErr := ioutil.WriteFile(absPath, []byte(kubeconfig), 0600)
		if writeErr != nil {
			return writeErr
		}

		return nil
	}

	command.PreRunE = func(command *cobra.Command, args []string) error {
		_, ipErr := command.Flags().GetIP("ip")
		if ipErr != nil {
			return ipErr
		}

		_, sshPortErr := command.Flags().GetInt("ssh-port")
		if sshPortErr != nil {
			return sshPortErr
		}
		return nil
	}
	return command
}

func expandPath(path string) string {
	res, _ := homedir.Expand(path)
	return res
}

func sshAgent(publicKeyPath string) (ssh.AuthMethod, func() error) {
	if sshAgentConn, err := net.Dial("unix", os.Getenv("SSH_AUTH_SOCK")); err == nil {
		sshAgent := agent.NewClient(sshAgentConn)

		keys, _ := sshAgent.List()
		if len(keys) == 0 {
			return nil, sshAgentConn.Close
		}

		pubkey, err := ioutil.ReadFile(publicKeyPath)
		if err != nil {
			return nil, sshAgentConn.Close
		}

		authkey, _, _, _, err := ssh.ParseAuthorizedKey(pubkey)
		if err != nil {
			return nil, sshAgentConn.Close
		}
		parsedkey := authkey.Marshal()

		for _, key := range keys {
			if bytes.Equal(key.Blob, parsedkey) {
				return ssh.PublicKeysCallback(sshAgent.Signers), sshAgentConn.Close
			}
		}
	}
	return nil, func() error { return nil }
}

func loadPublickey(path string) (ssh.AuthMethod, func() error, error) {
	noopCloseFunc := func() error { return nil }

	key, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, noopCloseFunc, err
	}

	signer, err := ssh.ParsePrivateKey(key)
	if err != nil {
		if err.Error() != "ssh: cannot decode encrypted private keys" {
			return nil, noopCloseFunc, err
		}

		agent, close := sshAgent(path + ".pub")
		if agent != nil {
			return agent, close, nil
		}

		defer close()

		fmt.Printf("Enter passphrase for '%s': ", path)
		bytePassword, err := terminal.ReadPassword(int(os.Stdin.Fd()))
		fmt.Println()

		signer, err = ssh.ParsePrivateKeyWithPassphrase(key, bytePassword)
		if err != nil {
			return nil, noopCloseFunc, err
		}
	}

	return ssh.PublicKeys(signer), noopCloseFunc, nil
}
