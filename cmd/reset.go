package cmd

import (
    "encoding/json"
    "fmt"
    "io/ioutil"
    "os"
    "os/exec"
    "strings"

    "github.com/spf13/cobra"
)

type Node struct {
    Hostname string `json:"hostname"`
    IP       string `json:"ip"`
}

func MakeReset() *cobra.Command {
    var user, ip, sshKey, plan string

    cmd := &cobra.Command{
        Use:   "reset",
        Short: "Uninstall k3s on specified nodes",
        RunE: func(cmd *cobra.Command, args []string) error {
            if user == "" || (ip == "" && plan == "") {
                return fmt.Errorf("Usage: %s", cmd.UsageString())
            }

            if plan != "" {
                return uninstallK3sFromPlan(user, sshKey, plan)
            }
            return uninstallK3s(user, sshKey, ip)
        },
    }

    cmd.Flags().StringVarP(&user, "user", "u", "", "Username for SSH connection")
    cmd.Flags().StringVarP(&ip, "ip", "i", "", "IP address of the host")
    cmd.Flags().StringVar(&sshKey, "ssh-key", os.Getenv("HOME")+"/.ssh/id_rsa", "Path to the private SSH key")
    cmd.Flags().StringVar(&plan, "plan", "", "JSON file containing the list of nodes")

    return cmd
}

func uninstallK3s(user, sshKey, ip string) error {
    fmt.Printf("Uninstalling k3s on host %s\n", ip)
    cmd := exec.Command("ssh", "-i", sshKey, "-o", "BatchMode=yes", "-o", "StrictHostKeyChecking=no", "-o", "ConnectTimeout=10", fmt.Sprintf("%s@%s", user, ip), "bash -s")
    cmd.Stdin = strings.NewReader(`
if [ -f /usr/local/bin/k3s-uninstall.sh ]; then
    /usr/local/bin/k3s-uninstall.sh
    echo "k3s server uninstalled successfully."
elif [ -f /usr/local/bin/k3s-agent-uninstall.sh ]; then
    /usr/local/bin/k3s-agent-uninstall.sh
    echo "k3s agent uninstalled successfully."
else
    echo "Neither k3s-uninstall.sh nor k3s-agent-uninstall.sh found."
    exit 1
fi
`)
    output, err := cmd.CombinedOutput()
    fmt.Printf("%s\n", output)
    if err != nil {
        return fmt.Errorf("failed to execute script on %s: %v", ip, err)
    }
    return nil
}

func uninstallK3sFromPlan(user, sshKey, plan string) error {
    data, err := ioutil.ReadFile(plan)
    if err != nil {
        return fmt.Errorf("unable to read JSON file %s: %v", plan, err)
    }

    var nodes []Node
    if err := json.Unmarshal(data, &nodes); err != nil {
        return fmt.Errorf("error parsing JSON file %s: %v", plan, err)
    }

    var successNodes []Node
    var failedNodes []Node

    for _, node := range nodes {
        fmt.Printf("Uninstalling k3s on %s (%s)\n", node.Hostname, node.IP)
        if err := uninstallK3s(user, sshKey, node.IP); err != nil {
            fmt.Printf("Error: %v\n", err)
            failedNodes = append(failedNodes, node)
        } else {
            fmt.Printf("k3s successfully uninstalled on %s (%s)\n", node.Hostname, node.IP)
            successNodes = append(successNodes, node)
        }
    }

    fmt.Println("\nSummary of uninstallation:")
    fmt.Println("Successful:")
    for _, node := range successNodes {
        fmt.Printf("- %s (%s)\n", node.Hostname, node.IP)
    }

    if len(failedNodes) > 0 {
        fmt.Println("Failed:")
        for _, node := range failedNodes {
            fmt.Printf("- %s (%s)\n", node.Hostname, node.IP)
        }
    }

    return nil
}
