package cmd

import (
	"fmt"
	"net"
	"strings"

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

	command.Flags().IP("ip", net.ParseIP("127.0.0.1"), "Public IP of node on which to install agent")
	command.Flags().IP("server-ip", net.ParseIP("127.0.0.1"), "Public IP of an existing k3s server")

	command.Flags().String("host", "", "Public hostname of node on which to install agent")
	command.Flags().String("server-host", "", "Public hostname of an existing k3s server")

	command.Flags().String("user", "root", "Username for SSH login")
	command.Flags().String("server-user", "root", "Server username for SSH login (Default to --user)")

	command.Flags().String("ssh-key", "~/.ssh/id_rsa", "The ssh key to use for remote login")
	command.Flags().Int("ssh-port", 22, "The port on which to connect for ssh")
	command.Flags().Int("server-ssh-port", 22, "The port on which to connect to server for ssh (Default to --ssh-port)")
	command.Flags().Bool("skip-install", false, "Skip the k3s installer")
	command.Flags().Bool("sudo", true, "Use sudo for installation. e.g. set to false when using the root user and no sudo is available.")

	command.Flags().Bool("server", false, "Join the cluster as a server rather than as an agent")
	command.Flags().Bool("print-command", false, "Print a command that you can use with SSH to manually recover from an error")

	command.Flags().String("k3s-extra-args", "", "Optional extra arguments to pass to k3s installer, wrapped in quotes (e.g. --k3s-extra-args '--node-taint key=value:NoExecute')")
	command.Flags().String("k3s-version", "", "Optional: set a version to install, overrides k3s-channel")
	command.Flags().String("k3s-channel", PinnedK3sChannel, "Optional release channel: stable, latest, or i.e. v1.18")

	command.RunE = func(command *cobra.Command, args []string) error {
		fmt.Printf("Running: k3sup join\n")

		ip, err := command.Flags().GetIP("ip")
		if err != nil {
			return err
		}

		host, err := command.Flags().GetString("host")
		if err != nil {
			return err
		}
		if len(host) == 0 {
			host = ip.String()
		}

		serverIP, err := command.Flags().GetIP("server-ip")
		if err != nil {
			return err
		}

		serverHost, err := command.Flags().GetString("server-host")
		if err != nil {
			return err
		}
		if len(serverHost) == 0 {
			serverHost = serverIP.String()
		}

		fmt.Println("Server IP: " + serverHost)

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

		k3sVersion, err := command.Flags().GetString("k3s-version")
		if err != nil {
			return err
		}
		k3sExtraArgs, err := command.Flags().GetString("k3s-extra-args")
		if err != nil {
			return err
		}
		k3sChannel, err := command.Flags().GetString("k3s-channel")
		if err != nil {
			return err
		}

		if len(k3sVersion) == 0 && len(k3sChannel) == 0 {
			return fmt.Errorf("give a value for --k3s-version or --k3s-channel")
		}

		printCommand, err := command.Flags().GetBool("print-command")
		if err != nil {
			return err
		}

		useSudo, err := command.Flags().GetBool("sudo")
		if err != nil {
			return err
		}
		sudoPrefix := ""
		if useSudo {
			sudoPrefix = "sudo "
		}

		sshKeyPath := expandPath(sshKey)

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
		if printCommand {
			fmt.Printf("ssh: %s\n", getTokenCommand)
		}

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
			boostrapErr = setupAdditionalServer(serverHost, host, port, user, sshKeyPath, joinToken, k3sExtraArgs, k3sVersion, k3sChannel, printCommand)
		} else {
			boostrapErr = setupAgent(serverHost, host, port, user, sshKeyPath, joinToken, k3sExtraArgs, k3sVersion, k3sChannel, printCommand)
		}

		return boostrapErr
	}

	command.PreRunE = func(command *cobra.Command, args []string) error {
		_, err := command.Flags().GetIP("ip")
		if err != nil {
			return err
		}
		_, err = command.Flags().GetIP("server-ip")
		if err != nil {
			return err
		}
		_, err = command.Flags().GetString("host")
		if err != nil {
			return err
		}
		_, err = command.Flags().GetString("server-host")
		if err != nil {
			return err
		}
		_, err = command.Flags().GetInt("ssh-port")
		if err != nil {
			return err
		}
		return nil
	}

	return command
}

func setupAdditionalServer(serverHost, host string, port int, user, sshKeyPath, joinToken, k3sExtraArgs, k3sVersion, k3sChannel string, printCommand bool) error {

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

	installStr := createVersionStr(k3sVersion, k3sChannel)
	serverAgent := true

	defer operator.Close()

	installk3sExec := makeJoinExec(
		serverHost,
		strings.TrimSpace(joinToken),
		installStr,
		k3sExtraArgs,
		serverAgent,
	)

	installAgentServerCommand := fmt.Sprintf("%s | %s", getScript, installk3sExec)

	if printCommand {
		fmt.Printf("ssh: %s\n", installAgentServerCommand)
	}

	res, err := operator.Execute(installAgentServerCommand)
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

func setupAgent(serverHost, host string, port int, user, sshKeyPath, joinToken, k3sExtraArgs, k3sVersion, k3sChannel string, printCommand bool) error {

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

	installStr := createVersionStr(k3sVersion, k3sChannel)

	serverAgent := false

	installK3sExec := makeJoinExec(
		serverHost,
		strings.TrimSpace(joinToken),
		installStr,
		k3sExtraArgs,
		serverAgent,
	)

	installAgentCommand := fmt.Sprintf("%s | %s", getScript, installK3sExec)

	if printCommand {
		fmt.Printf("ssh: %s\n", installAgentCommand)
	}

	res, err := operator.Execute(installAgentCommand)

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

func createVersionStr(k3sVersion, k3sChannel string) string {
	installStr := ""
	if len(k3sVersion) > 0 {
		installStr = fmt.Sprintf("INSTALL_K3S_VERSION='%s'", k3sVersion)
	} else {
		installStr = fmt.Sprintf("INSTALL_K3S_CHANNEL='%s'", k3sChannel)
	}
	return installStr
}

func makeJoinExec(serverIP, joinToken, installStr, k3sExtraArgs string, serverAgent bool) string {

	installEnvVar := []string{}
	installEnvVar = append(installEnvVar, fmt.Sprintf("K3S_URL='https://%s:6443'", serverIP))
	installEnvVar = append(installEnvVar, fmt.Sprintf("K3S_TOKEN='%s'", joinToken))
	installEnvVar = append(installEnvVar, installStr)

	if serverAgent {
		installEnvVar = append(installEnvVar, fmt.Sprintf("INSTALL_K3S_EXEC='server --server https://%s:6443'", serverIP))
	}

	joinExec := strings.Join(installEnvVar, " ")
	joinExec += " sh -s -"

	if len(k3sExtraArgs) > 0 {
		installEnvVar = append(installEnvVar, k3sExtraArgs)
		joinExec += fmt.Sprintf(" %s", k3sExtraArgs)
	}

	return joinExec
}
