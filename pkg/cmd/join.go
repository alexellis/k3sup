package cmd

import (
	"fmt"
	"net"
	"strings"

	config "github.com/alexellis/k3sup/pkg/config"
	kssh "github.com/alexellis/k3sup/pkg/ssh"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"golang.org/x/crypto/ssh"
)

func MakeJoin() *cobra.Command {
	var command = &cobra.Command{
		Use:          "join",
		Short:        "Install the k3s agent on a remote host and join it to an existing server",
		Long:         `Install the k3s agent on a remote host and join it to an existing server`,
		Example:      `  k3sup join --user root --server-ip 192.168.0.100 --ip 192.168.0.101`,
		SilenceUsage: true,
	}

	command.Flags().IP("server-ip", nil, "Public IP of existing k3s server")
	command.Flags().IP("ip", nil, "Public IP of node on which to install agent")

	command.Flags().String("user", "root", "Username for SSH login")

	command.Flags().String("ssh-key", "~/.ssh/id_rsa", "The ssh key to use for remote login")
	command.Flags().Int("ssh-port", 22, "The port on which to connect for ssh")
	command.Flags().Bool("skip-install", false, "Skip the k3s installer")
	command.Flags().Bool("sudo", true, "Use sudo for installation. e.g. set to false when using the root user and no sudo is available.")
	command.Flags().String("k3s-extra-args", "", "Optional extra arguments to pass to k3s installer, wrapped in quotes (e.g. --k3s-extra-args '--node-taint key=value:NoExecute')")
	command.Flags().String("k3s-version", config.K3sVersion, "Optional version to install, pinned at a default")

	command.RunE = func(command *cobra.Command, args []string) error {

		ip, _ := command.Flags().GetIP("ip")

		serverIP, _ := command.Flags().GetIP("server-ip")

		fmt.Println("Server IP: " + serverIP.String())

		user, _ := command.Flags().GetString("user")
		sshKey, _ := command.Flags().GetString("ssh-key")

		port, _ := command.Flags().GetInt("ssh-port")

		k3sExtraArgs, _ := command.Flags().GetString("k3s-extra-args")

		k3sVersion, _ := command.Flags().GetString("k3s-version")

		useSudo, _ := command.Flags().GetBool("sudo")
		sudoPrefix := ""
		if useSudo {
			sudoPrefix = "sudo "
		}

		sshKeyPath := expandPath(sshKey)
		fmt.Printf("ssh -i %s %s@%s\n", sshKeyPath, user, serverIP.String())

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

		address := fmt.Sprintf("%s:%d", serverIP.String(), port)
		operator, err := kssh.NewSSHOperator(address, config)

		if err != nil {
			return errors.Wrapf(err, "unable to connect to %s over ssh", address)
		}

		defer operator.Close()

		getTokenCommand := fmt.Sprintf(sudoPrefix + "cat /var/lib/rancher/k3s/server/node-token\n")
		fmt.Printf("ssh: %s\n", getTokenCommand)

		res, err := operator.Execute(getTokenCommand)

		if err != nil {
			return errors.Wrap(err, "unable to get join-token from server")
		}

		if len(res.StdErr) > 0 {
			fmt.Printf("Logs: %s", res.StdErr)
		}

		joinToken := string(res.StdOut)

		setupAgent(serverIP, ip, port, user, sshKeyPath, joinToken, k3sExtraArgs, k3sVersion)

		return nil
	}

	command.PreRunE = func(command *cobra.Command, args []string) error {
		_, ipErr := command.Flags().GetIP("ip")
		if ipErr != nil {
			return ipErr
		}

		_, ipErr = command.Flags().GetIP("server-ip")
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

func setupAgent(serverIP, ip net.IP, port int, user, sshKeyPath, joinToken, k3sExtraArgs, k3sVersion string) error {

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

	getTokenCommand := fmt.Sprintf("curl -sfL https://get.k3s.io/ | K3S_URL='https://%s:6443' K3S_TOKEN='%s' INSTALL_K3S_VERSION='%s' sh -s - %s", serverIP.String(), strings.TrimSpace(joinToken), k3sVersion, k3sExtraArgs)
	fmt.Printf("ssh: %s\n", getTokenCommand)

	res, err := operator.Execute(getTokenCommand)

	if err != nil {
		return errors.Wrap(err, "unable to setup agent")
	}

	if len(res.StdErr) > 0 {
		fmt.Printf("Logs: %s", res.StdErr)
	}

	joinRes := string(res.StdOut)
	fmt.Printf("Output: %s", string(joinRes))

	return nil
}
