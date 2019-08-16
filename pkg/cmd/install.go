package cmd

import (
	"fmt"
	"io/ioutil"
	"path/filepath"
	"strings"

	kssh "github.com/alexellis/k3sup/pkg/ssh"

	homedir "github.com/mitchellh/go-homedir"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"golang.org/x/crypto/ssh"
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

		config := &ssh.ClientConfig{
			User: user,
			Auth: []ssh.AuthMethod{
				loadPublickey(sshKeyPath),
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

func loadPublickey(path string) ssh.AuthMethod {

	key, err := ioutil.ReadFile(path)
	if err != nil {
		panic(err)
	}
	signer, err := ssh.ParsePrivateKey(key)
	if err != nil {
		panic(err)
	}
	return ssh.PublicKeys(signer)
}
