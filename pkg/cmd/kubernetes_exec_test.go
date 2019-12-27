package cmd

import "testing"

func indexOf(element string, data []string) int {
	for k, v := range data {
		if element == v {
			return k
		}
	}
	return -1 //not found.
}
func Test_kubectlArgs_1(t *testing.T) {
	task := kubectl("", "", "apply")
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
	configPath := "~/.kube/test"
	task := kubectl(configPath, "", "apply")
	got := indexOf("--kubeconfig", task.Args)

	if configPath != task.Args[got+1] {
		t.Errorf("Args order is wrong, want: %s, got:%s", configPath, task.Args[got+1])
	}
	if "apply" != task.Args[0] {
		t.Errorf("suffix, want: %s, got: %s", "apply", task.Args[0])
	}
}
