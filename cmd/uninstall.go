package cmd

import (
	"fmt"
	"net"

	operator "github.com/alexellis/k3sup/pkg/operator"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"golang.org/x/crypto/ssh"
)

func MakeUninstall() *cobra.Command {
	var command = &cobra.Command{
		Use:          "uninstall",
		Short:        "Uninstall k3s from a node",
		Long:         "Uninstall k3s from a node",
		Example:      `k3sup uninstall --ip 192.168.0.100 --user root`,
		SilenceUsage: true,
	}
	command.Flags().IP("ip", net.ParseIP("127.0.0.1"), "Public IP of node")
	command.Flags().String("user", "root", "Username for SSH login")
	command.Flags().String("ssh-key", "~/.ssh/id_rsa", "The ssh key to use for remote login")
	command.Flags().Int("ssh-port", 22, "The port on which to connect for ssh")
	command.Flags().Bool("server", false, "Remove a server node")

	command.RunE = func(command *cobra.Command, args []string) error {
		fmt.Printf("Running: k3sup uninstall\n")

		ip, _ := command.Flags().GetIP("ip")
		fmt.Println("IP: " + ip.String())

		user, _ := command.Flags().GetString("user")

		port, _ := command.Flags().GetInt("ssh-port")
		serverPort := port

		sshKey, _ := command.Flags().GetString("ssh-key")

		server, getServerErr := command.Flags().GetBool("server")
		if getServerErr != nil {
			return getServerErr
		}

		sshKeyPath := expandPath(sshKey)
		fmt.Println(sshKeyPath)
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

		address := fmt.Sprintf("%s:%d", ip.String(), serverPort)
		operator, err := operator.NewSSHOperator(address, config)

		if err != nil {
			return errors.Wrapf(err, "unable to connect to (server) %s over ssh", address)
		}

		defer operator.Close()

		var uninstallCommand string
		if server {
			uninstallCommand = fmt.Sprintf("sh /usr/local/bin/k3s-uninstall.sh\n")
		} else {
			uninstallCommand = fmt.Sprintf("sh /usr/local/bin/k3s-agent-uninstall.sh\n")
		}

		res, err := operator.Execute(uninstallCommand)
		if err != nil {
			return errors.Wrap(err, "unable to uninstall k3s from the node")
		}

		if len(res.StdErr) > 0 {
			fmt.Printf("Logs: %s", res.StdErr)
		}

		closeSSHAgent()
		operator.Close()

		fmt.Println(UninstallInfoMsg)

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

const UninstallInfoMsg = `
# k3s has been uninstalled from the node

## To delete the node from your cluster, run:

kubectl delete node <node-name>
`
