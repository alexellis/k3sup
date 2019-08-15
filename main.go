package main

import (
	"fmt"
	"io/ioutil"
	"path/filepath"

	homedir "github.com/mitchellh/go-homedir"
	"github.com/spf13/cobra"
	"golang.org/x/crypto/ssh"
)

func main() {

	cmdInstall := makeInstallCmd()

	cmdVersion := makeVersionCmd()

	var rootCmd = &cobra.Command{Use: "app"}

	rootCmd.AddCommand(cmdInstall)
	rootCmd.AddCommand(cmdVersion)
	rootCmd.Execute()
}

func makeVersionCmd() *cobra.Command {
	var cmd = &cobra.Command{
		Use:          "version",
		Short:        "Print the version",
		Example:      `  k3sup version`,
		SilenceUsage: false,
	}
	cmd.Run = func(cmd *cobra.Command, args []string) {
		fmt.Printf("Welcome to k3sup!\n")
	}
	return cmd
}

func makeInstallCmd() *cobra.Command {
	var cmd = &cobra.Command{
		Use:          "install k3s",
		Short:        "Install k3s on a server via SSH",
		Long:         `Install k3s on a server via SSH.`,
		Example:      `  k3sup install --ip 192.168.0.100 --user root`,
		SilenceUsage: true,
	}

	cmd.Flags().IP("ip", nil, "Public IP of node")
	cmd.Flags().String("user", "root", "Username for SSH login")

	cmd.Flags().String("ssh-key", "~/.ssh/id_rsa", "The ssh key to use for remote login")
	cmd.Flags().Bool("skip-install", false, "Skip the k3s installer")
	cmd.Flags().String("local-path", "kubeconfig", "Local path to save the kubeconfig file")

	cmd.Run = func(cmd *cobra.Command, args []string) {

		localKubeconfig, _ := cmd.Flags().GetString("local-path")

		skipInstall, _ := cmd.Flags().GetBool("skip-install")

		ip, _ := cmd.Flags().GetIP("ip")
		fmt.Println("Public SAN: " + ip.String())

		user, _ := cmd.Flags().GetString("user")
		sshKey, _ := cmd.Flags().GetString("ssh-key")
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
			installK3sCmd := fmt.Sprintf("curl -sLS https://get.k3s.io | INSTALL_K3S_EXEC='server --tls-san %s' sh -\n", ip)
			fmt.Printf("ssh: %s\n", installK3sCmd)
		}

		getConfigCmd := fmt.Sprintf("sudo cat /etc/rancher/k3s/k3s.yaml\n")
		fmt.Printf("ssh: %s\n", getConfigCmd)

		absPath, _ := filepath.Abs(localKubeconfig)
		fmt.Printf("Saving file to: %s\n", absPath)

	}

	cmd.PreRunE = func(cmd *cobra.Command, args []string) error {
		// if val, _ := cmd.Flags().getip .GetString("ip"); len(val) == 0 {
		// 	return fmt.Errorf(`give --ip or install --help`)
		// }
		_, ipErr := cmd.Flags().GetIP("ip")
		if ipErr != nil {
			return ipErr
		}

		return nil
	}
	return cmd
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
