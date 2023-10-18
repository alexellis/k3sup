package cmd

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/alexellis/k3sup/pkg"
	"github.com/spf13/cobra"
)

func MakePlan() *cobra.Command {
	var command = &cobra.Command{
		Use:   "plan",
		Short: "Plan an installation of K3s.",
		Long: `Generate a plan of installation commands for K3s for a HA cluster.

Format:

[{
	"hostname": "node-1",
	"ip": "192.168.1.128"
}]

` + pkg.SupportMessageShort + `
`,
		Example: `  Generate an installation script where 3x of the
  available hosts are dedicated as servers.
  k3sup plan hosts.json --servers 3
`,
		SilenceUsage: true,
	}

	command.Flags().Int("servers", 3, "Number of servers to use from the pool of hosts")
	command.Flags().String("local-path", "kubeconfig", "Where to save the kubeconfig file")
	command.Flags().String("context", "default", "Name of the kubeconfig context to use")
	command.Flags().String("user", "root", "Username for SSH login")

	command.Flags().String("ssh-key", "", "Path to the private key for SSH login")
	command.Flags().String("tls-san", "", "SAN for TLS certificates, can be a comma-separated list")

	// Background
	command.Flags().Bool("background", false, "Run the installation in the background for all agents/nodes after the first server is up")

	command.Flags().Int("limit", 0, "Maximum number of nodes to use from the pool of hosts, 0 for all")

	command.RunE = func(cmd *cobra.Command, args []string) error {

		if len(args) == 0 {
			return fmt.Errorf("give a path to a JSON file containing a list of hosts")
		}

		nodeLimit, _ := cmd.Flags().GetInt("limit")
		name := args[0]
		data, err := os.ReadFile(name)
		if err != nil {
			return err
		}

		background, _ := cmd.Flags().GetBool("background")

		var hosts []Host
		if err = json.Unmarshal(data, &hosts); err != nil {
			return err
		}

		servers, _ := cmd.Flags().GetInt("servers")
		kubeconfig, _ := cmd.Flags().GetString("local-path")
		contextName, _ := cmd.Flags().GetString("context")
		user, _ := cmd.Flags().GetString("user")
		tlsSan, _ := cmd.Flags().GetString("tls-san")

		tlsSanStr := ""
		if len(tlsSan) > 0 {
			tlsSanStr = fmt.Sprintf(` \
--tls-san %s`, tlsSan)
		}
		// sshKey, _ := cmd.Flags().GetString("ssh-key")

		bgStr := ""
		if background {
			bgStr = " &"
		}

		serversAdded := 0
		var primaryServer Host
		script := "#!/bin/sh\n\n"

		for i, host := range hosts {
			if serversAdded == 0 {

				script += `echo ""Setting up primary server 1
`

				script += fmt.Sprintf(`k3sup install --host %s \
--user %s \
--cluster \
--local-path %s \
--context %s%s
`,
					host.IP,
					user,
					kubeconfig,
					contextName, tlsSanStr)

				script += fmt.Sprintf(`
echo "Saving the server's node-token to ./token.txt"

k3sup node-token --host %s \
--user %s > token.txt
`, host.IP, user)

				serversAdded = 1
				primaryServer = host
			} else if serversAdded < servers {
				script += fmt.Sprintf("\necho \"Setting up additional server: %d\"\n", serversAdded+1)

				script += fmt.Sprintf(`k3sup join --host %s \
--server-host %s \
--server \
--node-token-path ./token.txt \
--user %s%s%s
`, host.IP, primaryServer.IP, user, tlsSanStr, bgStr)

				serversAdded++
			} else {
				script += fmt.Sprintf("\necho \"Setting up worker: %d\"\n", (i+1)-serversAdded)

				script += fmt.Sprintf(`k3sup join --host %s \
--server-host %s \
--node-token-path ./token.txt \
--user %s%s
`, host.IP, primaryServer.IP, user, bgStr)
			}

			if nodeLimit > 0 && i+1 >= nodeLimit {
				break
			}
		}

		fmt.Printf("%s\n", script)

		return nil
	}

	return command
}

type Host struct {
	Hostname string `json:"hostname"`
	IP       string `json:"ip"`
}
