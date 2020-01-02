package cmd

import (
	"github.com/spf13/cobra"
	"testing"
)

var command = &cobra.Command{
	Use:          "kafka-connector",
	Short:        "Install kafka-connector for OpenFaaS",
	Long:         `Install kafka-connector for OpenFaaS`,
	Example:      `  k3sup app install kafka-connector`,
	SilenceUsage: true,
}

func indexOf(element string, data []string) int {
	for k, v := range data {
		if element == v {
			return k
		}
	}
	return -1 //not found.
}
func Test_kubectlArgs_1(t *testing.T) {
	task := kubectl(command, "apply")
	want := -1
	got := indexOf("--kubeconfig", task.Args)
	if want != got {
		t.Errorf("unneeded argument --kubeconfig passed")
	}
	if "apply" != task.Args[0] {
		t.Errorf("suffix, want: %s, got: %s", "apply", task.Args[0])
	}
}

func Test_kubectlArgs_2(t *testing.T) {
	command.Flags().String("kubeconfig", "", "Local path for your kubeconfig file")
	configPath:="~/.kube/test"
	command.Flags().Set("kubeconfig",  configPath)
	task := kubectl(command, "apply")
	got := indexOf("--kubeconfig", task.Args)

	if configPath != task.Args[got+1] && got != -1 {
		t.Errorf("Args order is wrong, want: %s, got:%s", configPath, task.Args[got+1])
	}
	if "apply" != task.Args[0] {
		t.Errorf("suffix, want: %s, got: %s", "apply", task.Args[0])
	}
}
