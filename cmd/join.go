package cmd

import (
	"fmt"
	"net"
	"strings"

	config "github.com/alexellis/k3sup/pkg/config"
	operator "github.com/alexellis/k3sup/pkg/operator"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"golang.org/x/crypto/ssh"
)

func MakeJoin() *cobra.Command {
	var command = &cobra.Command{
		Use:          "join",
		Short:        "Install the k3s agent on a remote host and join it to an existing server",
		Long:         `Install the k3s agent on a remote host and join it to an existing server`,
		Example:      `  k3sup join --user root --server-host 192.168.0.100 --host 192.168.0.101`,
		SilenceUsage: true,
	}

	command.Flags().IP("server-ip", net.ParseIP("127.0.0.1"), "Public IP of existing k3s server")
	command.Flags().String("server-host", "", "Public IP or hostname of existing k3s server")

	command.Flags().IP("ip", net.ParseIP("127.0.0.1"), "Public IP of node on which to install agent")
	command.Flags().String("host", "", "Public IP or hostname of node on which to install agent")

	command.Flags().String("user", "root", "Username for SSH login")
	command.Flags().String("server-user", "root", "Server username for SSH login (Default to --user)")

	command.Flags().String("ssh-key", "~/.ssh/id_rsa", "The ssh key to use for remote login")
	command.Flags().Int("ssh-port", 22, "The port on which to connect for ssh")
	command.Flags().Int("server-ssh-port", 22, "The port on which to connect to server for ssh (Default to --ssh-port)")
	command.Flags().Bool("skip-install", false, "Skip the k3s installer")
	command.Flags().Bool("sudo", true, "Use sudo for installation. e.g. set to false when using the root user and no sudo is available.")
	command.Flags().String("k3s-extra-args", "", "Optional extra arguments to pass to k3s installer, wrapped in quotes (e.g. --k3s-extra-args '--node-taint key=value:NoExecute')")
	command.Flags().String("k3s-version", config.K3sVersion, "Optional version to install, pinned at a default")

	command.Flags().Bool("server", false, "Join the cluster as a server rather than as an agent")

	command.Flags().MarkDeprecated("server-ip", "please use --server-host instead")
	command.Flags().MarkDeprecated("ip", "please use --host instead")

	command.RunE = func(command *cobra.Command, args []string) error {
		fmt.Printf("Running: k3sup join\n")

		ip, _ := command.Flags().GetIP("ip")
		host, _ := command.Flags().GetString("host")
		if host == "" {
			host = ip.String()
		}

		serverIP, _ := command.Flags().GetIP("server-ip")
		serverHost, _ := command.Flags().GetString("server-host")
		if serverHost == "" {
			serverHost = serverIP.String()
		}

		fmt.Println("Server Host: " + serverHost)

		user, _ := command.Flags().GetString("user")

		serverUser := user
		if command.Flags().Changed("server-user") {
			serverUser, _ = command.Flags().GetString("server-user")
		}

		sshKey, _ := command.Flags().GetString("ssh-key")
		server, getServerErr := command.Flags().GetBool("server")
		if getServerErr != nil {
			return getServerErr
		}

		port, _ := command.Flags().GetInt("ssh-port")
		serverPort := port
		if command.Flags().Changed("server-ssh-port") {
			serverPort, _ = command.Flags().GetInt("server-ssh-port")
		}

		k3sExtraArgs, _ := command.Flags().GetString("k3s-extra-args")
		k3sVersion, _ := command.Flags().GetString("k3s-version")

		useSudo, _ := command.Flags().GetBool("sudo")
		sudoPrefix := ""
		if useSudo {
			sudoPrefix = "sudo "
		}

		sshKeyPath := expandPath(sshKey)
		fmt.Printf("ssh -i %s -p %v %s@%s\n", sshKeyPath, serverPort, serverUser, serverHost)

		authMethod, closeSSHAgent, err := loadPublickey(sshKeyPath)
		if err != nil {
			return errors.Wrapf(err, "unable to load the ssh key with path %q", sshKeyPath)
		}

		defer closeSSHAgent()

		config := &ssh.ClientConfig{
			User: serverUser,
			Auth: []ssh.AuthMethod{
				authMethod,
			},
			HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		}

		address := fmt.Sprintf("%s:%d", serverHost, serverPort)
		operator, err := operator.NewSSHOperator(address, config)

		if err != nil {
			return errors.Wrapf(err, "unable to connect to (server) %s over ssh", address)
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

		closeSSHAgent()
		operator.Close()

		joinToken := string(res.StdOut)

		var boostrapErr error
		if server {
			boostrapErr = setupAdditionalServer(serverHost, host, port, user, sshKeyPath, joinToken, k3sExtraArgs, k3sVersion)
		} else {
			boostrapErr = setupAgent(serverHost, host, port, user, sshKeyPath, joinToken, k3sExtraArgs, k3sVersion)
		}

		return boostrapErr
	}

	command.PreRunE = func(command *cobra.Command, args []string) error {
		_, ipErr := command.Flags().GetIP("ip")
		if ipErr != nil {
			return ipErr
		}

		_, hostErr := command.Flags().GetString("host")
		if hostErr != nil {
			return hostErr
		}

		_, sshPortErr := command.Flags().GetInt("ssh-port")
		if sshPortErr != nil {
			return sshPortErr
		}
		return nil
	}

	return command
}

func setupAdditionalServer(serverHost, host string, port int, user, sshKeyPath, joinToken, k3sExtraArgs, k3sVersion string) error {

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

	address := fmt.Sprintf("%s:%d", host, port)
	operator, err := operator.NewSSHOperator(address, config)

	if err != nil {
		return errors.Wrapf(err, "unable to connect to %s over ssh as %s", address, user)
	}

	defer operator.Close()
	getTokenCommand := fmt.Sprintf("curl -sfL https://get.k3s.io/ | K3S_URL='https://%s:6443' INSTALL_K3S_EXEC='server --server https://%s:6443' K3S_TOKEN='%s' INSTALL_K3S_VERSION='%s' sh -s - %s",
		serverHost,
		serverHost,
		strings.TrimSpace(joinToken),
		k3sVersion,
		k3sExtraArgs)

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

func setupAgent(serverHost, host string, port int, user, sshKeyPath, joinToken, k3sExtraArgs, k3sVersion string) error {

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

	address := fmt.Sprintf("%s:%d", host, port)
	operator, err := operator.NewSSHOperator(address, config)

	if err != nil {
		return errors.Wrapf(err, "unable to connect to %s over ssh", address)
	}

	defer operator.Close()

	getTokenCommand := fmt.Sprintf("curl -sfL https://get.k3s.io/ | K3S_URL='https://%s:6443' K3S_TOKEN='%s' INSTALL_K3S_VERSION='%s' sh -s - %s",
		serverHost,
		strings.TrimSpace(joinToken),
		k3sVersion,
		k3sExtraArgs)

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
