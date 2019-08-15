package cmd

import (
	"fmt"
	"io/ioutil"
	"path/filepath"

	homedir "github.com/mitchellh/go-homedir"
	"github.com/spf13/cobra"
	"golang.org/x/crypto/ssh"
)

func MakeInstall() *cobra.Command {
	var command = &cobra.Command{
		Use:          "install k3s",
		Short:        "Install k3s on a server via SSH",
		Long:         `Install k3s on a server via SSH.`,
		Example:      `  k3sup install --ip 192.168.0.100 --user root`,
		SilenceUsage: true,
	}

	command.Flags().IP("ip", nil, "Public IP of node")
	command.Flags().String("user", "root", "Username for SSH login")

	command.Flags().String("ssh-key", "~/.ssh/id_rsa", "The ssh key to use for remote login")
	command.Flags().Bool("skip-install", false, "Skip the k3s installer")
	command.Flags().String("local-path", "kubeconfig", "Local path to save the kubeconfig file")

	command.Run = func(command *cobra.Command, args []string) {

		localKubeconfig, _ := command.Flags().GetString("local-path")

		skipInstall, _ := command.Flags().GetBool("skip-install")

		ip, _ := command.Flags().GetIP("ip")
		fmt.Println("Public SAN: " + ip.String())

		user, _ := command.Flags().GetString("user")
		sshKey, _ := command.Flags().GetString("ssh-key")
		sshKeyPath := expandPath(sshKey)
		fmt.Printf("ssh -i %s %s@%s\n", sshKeyPath, user, ip.String())

		clientConfig := ssh.ClientConfig{
			Auth: []ssh.AuthMethod{
				loadPublickey(sshKeyPath),
			},
			HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		}
		fmt.Println(clientConfig)

		if !skipInstall {
			installK3scommand := fmt.Sprintf("curl -sLS https://get.k3s.io | INSTALL_K3S_EXEC='server --tls-san %s' sh -\n", ip)
			fmt.Printf("ssh: %s\n", installK3scommand)
		}

		getConfigcommand := fmt.Sprintf("sudo cat /etc/rancher/k3s/k3s.yaml\n")
		fmt.Printf("ssh: %s\n", getConfigcommand)

		absPath, _ := filepath.Abs(localKubeconfig)
		fmt.Printf("Saving file to: %s\n", absPath)

	}

	command.PreRunE = func(command *cobra.Command, args []string) error {
		// if val, _ := command.Flags().getip .GetString("ip"); len(val) == 0 {
		// 	return fmt.Errorf(`give --ip or install --help`)
		// }
		_, ipErr := command.Flags().GetIP("ip")
		if ipErr != nil {
			return ipErr
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
