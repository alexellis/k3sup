package cmd

import (
	"fmt"
	"net"

	operator "github.com/alexellis/k3sup/pkg/operator"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"golang.org/x/crypto/ssh"
)

func MakeRemove() *cobra.Command {
	var command = &cobra.Command{
		Use:          "remove",
		Short:        "remove k3s",
		Long:         `remove k3s.`,
		Example:      `  k3sup remove --ip 192.168.0.100 --user ubuntu`,
		SilenceUsage: true,
	}
	command.Flags().IP("ip", net.ParseIP("127.0.0.1"), "Public IP of node")
	command.Flags().String("user", "root", "Username for SSH login")

	command.Flags().String("ssh-key", "~/.ssh/id_rsa", "The ssh key to use for remote login")
	command.Flags().Int("ssh-port", 22, "The port on which to connect for ssh")
	command.Flags().Bool("local", false, "Perform a local install without using ssh")
	command.Flags().Bool("sudo", true, "Use sudo for installation. e.g. set to false when using the root user and no sudo is available.")

	command.RunE = func(command *cobra.Command, args []string) error {

		fmt.Printf("Running: k3sup remove\n")

		local, _ := command.Flags().GetBool("local")

		ip, _ := command.Flags().GetIP("ip")

		useSudo, _ := command.Flags().GetBool("sudo")
		sudoPrefix := ""
		if useSudo {
			sudoPrefix = "sudo "
		}

		removeK3scommand := fmt.Sprintf(sudoPrefix + "/usr/local/bin/k3s-uninstall.sh")
		if local {
			operator := operator.ExecOperator{}

			fmt.Printf("Executing: %s\n", removeK3scommand)

			res, err := operator.Execute(removeK3scommand)
			if err != nil {
				return err
			}

			if len(res.StdErr) > 0 {
				fmt.Printf("stderr: %q", res.StdErr)
			}
			if len(res.StdOut) > 0 {
				fmt.Printf("stdout: %q", res.StdOut)
			}

			return nil
		}

		port, _ := command.Flags().GetInt("ssh-port")

		fmt.Println("Public IP: " + ip.String())

		user, _ := command.Flags().GetString("user")
		sshKey, _ := command.Flags().GetString("ssh-key")

		sshKeyPath := expandPath(sshKey)
		fmt.Printf("ssh -i %s -p %d %s@%s\n", sshKeyPath, port, user, ip.String())

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
		operator, err := operator.NewSSHOperator(address, config)

		if err != nil {
			return errors.Wrapf(err, "unable to connect to %s over ssh", address)
		}

		defer operator.Close()

		fmt.Printf("ssh: %s\n", removeK3scommand)
		res, err := operator.Execute(removeK3scommand)

		if err != nil {
			return fmt.Errorf("Error received processing command: %s", err)
		}

		fmt.Printf("Result: %s %s\n", string(res.StdOut), string(res.StdErr))

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
