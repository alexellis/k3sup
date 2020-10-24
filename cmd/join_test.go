package cmd

import (
	"testing"
)

type test struct {
	serverIP       string
	joinToken      string
	installStr     string
	installk3sExec string
	k3sExtraArgs   string
}

func Test_makeJoinExec(t *testing.T) {
	tests := []test{
		{
			serverIP:       "172.27.251.164",
			joinToken:      "K10c8bc21f68fef3f56d431a08df2e894481ab0a61a3c84cbd639b56449ad15523c::server:9d30861e1ba54177b8e4dd1426076e5d",
			installStr:     "INSTALL_K3S_VERSION=1.18",
			installk3sExec: "K3S_URL='https://172.27.251.164:6443' INSTALL_K3S_EXEC='server --server https://172.27.251.164:6443' K3S_TOKEN='K10c8bc21f68fef3f56d431a08df2e894481ab0a61a3c84cbd639b56449ad15523c::server:9d30861e1ba54177b8e4dd1426076e5d' INSTALL_K3S_VERSION=1.18 sh -s -",
			k3sExtraArgs:   "",
		},

		{
			serverIP:       "172.27.251.164",
			joinToken:      "K10c8bc21f68fef3f56d431a08df2e894481ab0a61a3c84cbd639b56449ad15523c::server:9d30861e1ba54177b8e4dd1426076e5d",
			installStr:     "INSTALL_K3S_VERSION=1.18",
			installk3sExec: "K3S_URL='https://172.27.251.164:6443' INSTALL_K3S_EXEC='server --server https://172.27.251.164:6443' K3S_TOKEN='K10c8bc21f68fef3f56d431a08df2e894481ab0a61a3c84cbd639b56449ad15523c::server:9d30861e1ba54177b8e4dd1426076e5d' INSTALL_K3S_VERSION=1.18 sh -s - --node-taint key=value:NoExecute",
			k3sExtraArgs:   "--node-taint key=value:NoExecute",
		},
	}

	for _, tc := range tests {
		got := makeJoinServerExec(tc.serverIP, tc.joinToken, tc.installStr, tc.k3sExtraArgs)

		if got != tc.installk3sExec {
			t.Errorf("want: %s, got: %s", tc.installk3sExec, got)
		}
	}

}
