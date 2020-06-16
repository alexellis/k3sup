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
		Example:      `  k3sup join --user root --server-ip 192.168.0.100 --ip 192.168.0.101`,
		SilenceUsage: true,
	}

	command.Flags().IP("server-ip", nil, "Public IP of existing k3s server")
	command.Flags().IP("ip", nil, "Public IP of node on which to install agent")

	command.Flags().String("user", "root", "Username for SSH login")
	command.Flags().String("server-user", "root", "Server username for SSH login (Default to --user)")
	command.Flags().String("k3s-envs", "", "k3s environment variables")
	command.Flags().String("ssh-key", "~/.ssh/id_rsa", "The ssh key to use for remote login")
	command.Flags().Int("ssh-port", 22, "The port on which to connect for ssh")
	command.Flags().Int("server-ssh-port", 22, "The port on which to connect to server for ssh (Default to --ssh-port)")
	command.Flags().Bool("skip-install", false, "Skip the k3s installer")
	command.Flags().Bool("sudo", true, "Use sudo for installation. e.g. set to false when using the root user and no sudo is available.")
	command.Flags().String("k3s-extra-args", "", "Optional extra arguments to pass to k3s installer, wrapped in quotes (e.g. --k3s-extra-args '--node-taint key=value:NoExecute')")
	command.Flags().String("k3s-version", config.K3sVersion, "Optional version to install, pinned at a default")
	command.Flags().String("k3s-install-url", "https://get.k3s.io", "k3s install shell url")
	command.Flags().Bool("server", false, "Join the cluster as a server rather than as an agent")

	command.RunE = func(command *cobra.Command, args []string) error {
		fmt.Printf("Running: k3sup join\n")

		ip, _ := command.Flags().GetIP("ip")

		serverIP, _ := command.Flags().GetIP("server-ip")

		fmt.Println("Server IP: " + serverIP.String())

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
		k3sInstallURL, _ := command.Flags().GetString("k3s-install-url")
		k3sEnvs, _ := command.Flags().GetString("k3s-envs")
		useSudo, _ := command.Flags().GetBool("sudo")
		sudoPrefix := ""
		if useSudo {
			sudoPrefix = "sudo "
		}

		sshKeyPath := expandPath(sshKey)
		fmt.Printf("ssh -i %s -p %v %s@%s\n", sshKeyPath, serverPort, serverUser, serverIP.String())

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

		address := fmt.Sprintf("%s:%d", serverIP.String(), serverPort)
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

		operator.Close()

		joinToken := string(res.StdOut)

		var boostrapErr error
		if server {
			boostrapErr = setupAdditionalServer(serverIP, ip, port, user, sshKeyPath, joinToken, k3sExtraArgs, k3sVersion, k3sEnvs, k3sInstallURL)
		} else {
			boostrapErr = setupAgent(serverIP, ip, port, user, sshKeyPath, joinToken, k3sExtraArgs, k3sVersion, k3sEnvs, k3sInstallURL)
		}

		return boostrapErr
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

func setupAdditionalServer(serverIP, ip net.IP, port int, user, sshKeyPath, joinToken, k3sExtraArgs, k3sVersion, k3sEnvs, k3sInstallURL string) error {

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
		return errors.Wrapf(err, "unable to connect to %s over ssh as %s", address, user)
	}

	defer operator.Close()

	getTokenCommand := fmt.Sprintf("curl -sfL %s | K3S_URL='https://%s:6443' INSTALL_K3S_EXEC='server --server https://%s:6443' K3S_TOKEN='%s' INSTALL_K3S_VERSION='%s' %s sh -s - %s",
		k3sInstallURL,
		serverIP.String(),
		serverIP.String(),
		strings.TrimSpace(joinToken),
		k3sVersion,
		k3sEnvs,
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

func setupAgent(serverIP, ip net.IP, port int, user, sshKeyPath, joinToken, k3sExtraArgs, k3sVersion, k3sEnvs, k3sInstallURL string) error {

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

	getTokenCommand := fmt.Sprintf("curl -sfL %s | K3S_URL='https://%s:6443' K3S_TOKEN='%s' INSTALL_K3S_VERSION='%s' %s sh -s - %s",
		k3sInstallURL,
		serverIP.String(),
		strings.TrimSpace(joinToken),
		k3sVersion,
		k3sEnvs,
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
