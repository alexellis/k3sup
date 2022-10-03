package cmd

import (
	"fmt"
	"os"
	"strings"
	"time"

	execute "github.com/alexellis/go-execute/pkg/v1"
	"github.com/alexellis/k3sup/pkg"
	"github.com/spf13/cobra"
)

func MakeReady() *cobra.Command {
	var command = &cobra.Command{
		Use:   "ready",
		Short: "Check if a cluster is ready using kubectl.",
		Long: `Check if the K3s cluster is ready using kubectl to query the nodes.

` + pkg.SupportMessageShort + `
`,
		Example: `  # Check from a local file, with the context "default"
  k3sup ready \
    --context default \
    --kubeconfig ./kubeconfig

  # Check a merged kubeconfig with a custom context
  k3sup ready \
    --context e2e \
    --kubeconfig $HOME/.kube/config
`,
		SilenceUsage: true,
	}

	command.Flags().Int("attempts", 25, "Number of attempts to check for readiness")
	command.Flags().Duration("pause", time.Second*2, "Pause between checking cluster for readiness")
	command.Flags().String("kubeconfig", "$HOME/.kube/config", "Path to the kubeconfig file")
	command.Flags().String("context", "default", "Name of the kubeconfig context to use")
	command.Flags().Bool("quiet", false, "Suppress output from each attempt")

	command.RunE = func(cmd *cobra.Command, args []string) error {

		attempts, _ := cmd.Flags().GetInt("attempts")
		pause, _ := cmd.Flags().GetDuration("pause")
		kubeconfig, _ := cmd.Flags().GetString("kubeconfig")
		contextName, _ := cmd.Flags().GetString("context")
		quiet, _ := cmd.Flags().GetBool("quiet")

		if len(kubeconfig) == 0 {
			return fmt.Errorf("kubeconfig cannot be empty")
		}

		if len(contextName) == 0 {
			return fmt.Errorf("context cannot be empty")
		}

		kubeconfig = os.ExpandEnv(kubeconfig)

		// Inspired by Kind: https://github.com/kubernetes-sigs/kind/blob/master/pkg/cluster/internal/create/actions/waitforready/waitforready.go
		for i := 0; i < attempts; i++ {
			if !quiet {
				fmt.Printf("Checking cluster status: %d/%d \n", i+1, attempts)
			}

			task := execute.ExecTask{
				Command: "kubectl",
				Args: []string{
					"get",
					"nodes",
					"--kubeconfig=" + kubeconfig,
					"--context=" + contextName,
					"-o=jsonpath='{.items..status.conditions[-1:].status}'",
				},
				StreamStdio: false,
			}

			res, err := task.Execute()
			if err != nil {
				return err
			}

			if strings.Contains(res.Stderr, "context was not found") {
				return fmt.Errorf("context %s not found in %s", contextName, kubeconfig)
			}

			if res.ExitCode == 0 {
				parts := strings.Split(strings.TrimSpace(res.Stdout), " ")

				ready := true
				for _, part := range parts {
					trimmed := strings.TrimSpace(part)

					// Note: The command is returning a single quoted string
					if len(trimmed) > 0 && trimmed != "'True'" {
						ready = false
						break
					}
				}

				if ready {
					if !quiet {
						fmt.Printf("All node(s) are ready\n")
					}
					break
				}
			}
			time.Sleep(pause)
		}

		return nil
	}
	return command
}
