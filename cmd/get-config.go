package cmd

import (
	"fmt"
	"log"
	"net"

	"github.com/alexellis/k3sup/pkg"
	operator "github.com/alexellis/k3sup/pkg/operator"

	"github.com/spf13/cobra"
)

// MakeGetConfig creates the get-config command
func MakeGetConfig() *cobra.Command {
	var command = &cobra.Command{
		Use:   "get-config",
		Short: "Get kubeconfig from an existing K3s installation",
		Long: `Create a local kubeconfig for use with kubectl from your local machine.

` + pkg.SupportMessageShort + `
`,
		Example: `  # Get the kubeconfig and save it to ./kubeconfig in the local
  # directory under the default context
  k3sup get-config --host HOST \
    --local-path ./kubeconfig

  # Merge kubeconfig into local file under custom context
  k3sup get-config \
    --host HOST \
    --merge \
    --local-path $HOME/.kube/kubeconfig \
    --context k3s-prod-eu-1

  # Get kubeconfig from local installation directly on a server
  # where you ran "k3sup install --local"
  k3sup get-config --local`,
		SilenceUsage: true,
	}

	command.Flags().IP("ip", net.ParseIP("127.0.0.1"), "Public IP of node")
	command.Flags().String("user", "root", "Username for SSH login")
	command.Flags().String("host", "", "Public hostname of node")
	command.Flags().String("ssh-key", "~/.ssh/id_rsa", "The ssh key to use for remote login")
	command.Flags().Int("ssh-port", 22, "The port on which to connect for ssh")
	command.Flags().Bool("sudo", true, "Use sudo for kubeconfig retrieval. e.g. set to false when using the root user and no sudo is available.")
	command.Flags().String("local-path", "kubeconfig", "Local path to save the kubeconfig file")
	command.Flags().String("context", "default", "Set the name of the kubeconfig context.")
	command.Flags().Bool("merge", false, `Merge the config with existing kubeconfig if it already exists.
Provide the --local-path flag with --merge if a kubeconfig already exists in some other directory`)
	command.Flags().Bool("print-command", false, "Print a command that you can use with SSH to manually recover from an error")
	command.Flags().Bool("local", false, "Perform a local get-config without using ssh")

	command.PreRunE = func(command *cobra.Command, args []string) error {
		local, err := command.Flags().GetBool("local")
		if err != nil {
			return err
		}

		if !local {
			_, err = command.Flags().GetString("host")
			if err != nil {
				return err
			}

			if _, err := command.Flags().GetIP("ip"); err != nil {
				return err
			}

			if _, err := command.Flags().GetInt("ssh-port"); err != nil {
				return err
			}
		}
		return nil
	}

	command.RunE = func(command *cobra.Command, args []string) error {
		localKubeconfig, _ := command.Flags().GetString("local-path")
		useSudo, err := command.Flags().GetBool("sudo")
		if err != nil {
			return err
		}

		sudoPrefix := ""
		if useSudo {
			sudoPrefix = "sudo "
		}

		local, _ := command.Flags().GetBool("local")

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

		log.Println(host)

		printCommand, err := command.Flags().GetBool("print-command")
		if err != nil {
			return err
		}

		merge, err := command.Flags().GetBool("merge")
		if err != nil {
			return err
		}
		context, err := command.Flags().GetString("context")
		if err != nil {
			return err
		}

		getConfigcommand := fmt.Sprintf(sudoPrefix + "cat /etc/rancher/k3s/k3s.yaml\n")

		if local {
			operator := operator.ExecOperator{}

			if err = obtainKubeconfig(operator, getConfigcommand, host, context, localKubeconfig, merge); err != nil {
				return err
			}

			return nil
		}

		fmt.Println("Public IP: " + host)

		port, _ := command.Flags().GetInt("ssh-port")
		user, _ := command.Flags().GetString("user")
		sshKey, _ := command.Flags().GetString("ssh-key")

		sshKeyPath := expandPath(sshKey)
		address := fmt.Sprintf("%s:%d", host, port)

		sshOperator, sshOperatorDone, errored, err := connectOperator(user, address, sshKeyPath)
		if errored {
			return err
		}

		if sshOperatorDone != nil {
			defer sshOperatorDone()
		}

		if printCommand {
			fmt.Printf("ssh: %s\n", getConfigcommand)
		}

		if err = obtainKubeconfig(sshOperator, getConfigcommand, host, context, localKubeconfig, merge); err != nil {
			return err
		}

		return nil
	}

	return command
}
