package cmd

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	kssh "github.com/alexellis/k3sup/pkg/ssh"

	homedir "github.com/mitchellh/go-homedir"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"golang.org/x/crypto/ssh"
)

var kubeconfig []byte

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

	command.Flags().Bool("merge", false, "Merge the config with existing kubeconfig if it already exists.\nProvide the --local-path flag with --merge if a kubeconfig already exists in some other directory")
	command.Flags().String("ssh-key", "~/.ssh/id_rsa", "The ssh key to use for remote login")
	command.Flags().Int("ssh-port", 22, "The port on which to connect for ssh")
	command.Flags().Bool("skip-install", false, "Skip the k3s installer")
	command.Flags().String("local-path", "kubeconfig", "Local path to save the kubeconfig file")

	command.RunE = func(command *cobra.Command, args []string) error {
		ip, _ := command.Flags().GetIP("ip")
		fmt.Println("Public IP: " + ip.String())

		user, _ := command.Flags().GetString("user")

		merge, _ := command.Flags().GetBool("merge")

		sshKey, _ := command.Flags().GetString("ssh-key")

		sshKeyPath := expandPath(sshKey)
		fmt.Printf("ssh -i %s %s@%s\n", sshKeyPath, user, ip.String())

		port, _ := command.Flags().GetInt("ssh-port")

		skipInstall, _ := command.Flags().GetBool("skip-install")

		localKubeconfig, _ := command.Flags().GetString("local-path")

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

		kubeconfig = []byte(strings.Replace(string(res.StdOut), "localhost", ip.String(), -1))

		if merge {
			// Create a merged kubeconfig
			kubeconfig, err = mergeConfigs(localKubeconfig, []byte(kubeconfig))
			if err != nil {
				return err
			}
		}
		// Create a new kubeconfig
		if writeErr := writeConfig(localKubeconfig, []byte(kubeconfig)); writeErr != nil {
			return writeErr
		}

		// Switch context
		fmt.Println("Switching to the current context: default")
		cmd := exec.Command("kubectl", "config", "set", "current-context", "default")
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("Could not switch to 'default' context")
		}
		fmt.Println("Context switched to 'default'")
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

// Generates config files give the path to file: string and the data: []byte
func writeConfig(path string, data []byte) error {
	absPath, _ := filepath.Abs(path)
	fmt.Printf("Saving file to: %s\n", absPath)
	writeErr := ioutil.WriteFile(absPath, []byte(data), 0600)
	if writeErr != nil {
		return writeErr
	}
	return nil
}

func mergeConfigs(localKubeconfigPath string, k3sconfig []byte) ([]byte, error) {
	// Create a temporary kubeconfig to store the config of the newly create k3s cluster
	file, err := ioutil.TempFile(os.TempDir(), "k3s-temp-*")
	if err != nil {
		return nil, fmt.Errorf("Could not generate a temporary file to store the kuebeconfig: %s", err)
	}
	defer file.Close()

	if writeErr := writeConfig(file.Name(), []byte(k3sconfig)); writeErr != nil {
		return nil, writeErr
	}

	fmt.Println("Merging with existing kubeconfig")

	// Append KUBECONFIGS in ENV Vars
	appendKubeConfigENV := fmt.Sprintf("KUBECONFIG=%s:%s", localKubeconfigPath, file.Name())

	// Merge the two kubeconfigs and read the output into 'data'
	cmd := exec.Command("kubectl", "config", "view", "--merge", "--flatten")
	cmd.Env = append(os.Environ(), appendKubeConfigENV)
	data, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("Could not merge kubeconfigs: %s", err)
	}

	// Remove the temporarily generated file
	err = os.Remove(file.Name())
	if err != nil {
		return nil, errors.Wrapf(err, "Could not remove temporary kubeconfig file: %s", file.Name())
	}

	return data, nil
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
