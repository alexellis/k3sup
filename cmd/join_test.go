package cmd

import (
	"testing"
)

type test struct {
	title          string
	serverIP       string
	joinToken      string
	installStr     string
	k3sExtraArgs   string
	serverAgent    bool
	installk3sExec string
	tlsSAN         string
}

func Test_makeJoinServerExec(t *testing.T) {
	tests := []test{
		{
			title:          "Join Server without k3sExtraArgs",
			serverIP:       "172.27.251.164",
			joinToken:      "K10c8bc21f68fef3f56d431a08df2e894481ab0a61a3c84cbd639b56449ad15523c::server:9d30861e1ba54177b8e4dd1426076e5d",
			installStr:     "INSTALL_K3S_VERSION=1.18",
			installk3sExec: "K3S_URL='https://172.27.251.164:6443' K3S_TOKEN='K10c8bc21f68fef3f56d431a08df2e894481ab0a61a3c84cbd639b56449ad15523c::server:9d30861e1ba54177b8e4dd1426076e5d' INSTALL_K3S_VERSION=1.18 INSTALL_K3S_EXEC='server --server https://172.27.251.164:6443' sh -s -",
			k3sExtraArgs:   "",
			serverAgent:    true,
		},

		{
			title:          "Join Server with K3sExtraArgs",
			serverIP:       "172.27.251.164",
			joinToken:      "K10c8bc21f68fef3f56d431a08df2e894481ab0a61a3c84cbd639b56449ad15523c::server:9d30861e1ba54177b8e4dd1426076e5d",
			installStr:     "INSTALL_K3S_VERSION=1.18",
			installk3sExec: "K3S_URL='https://172.27.251.164:6443' K3S_TOKEN='K10c8bc21f68fef3f56d431a08df2e894481ab0a61a3c84cbd639b56449ad15523c::server:9d30861e1ba54177b8e4dd1426076e5d' INSTALL_K3S_VERSION=1.18 INSTALL_K3S_EXEC='server --server https://172.27.251.164:6443' sh -s - --node-taint key=value:NoExecute",
			k3sExtraArgs:   "--node-taint key=value:NoExecute",
			serverAgent:    true,
		},

		{
			title:          "Join Server with K3sExtraArgs and TLS SAN",
			serverIP:       "172.27.251.164",
			joinToken:      "K10c8bc21f68fef3f56d431a08df2e894481ab0a61a3c84cbd639b56449ad15523c::server:9d30861e1ba54177b8e4dd1426076e5d",
			installStr:     "INSTALL_K3S_VERSION=1.18",
			installk3sExec: "K3S_URL='https://172.27.251.164:6443' K3S_TOKEN='K10c8bc21f68fef3f56d431a08df2e894481ab0a61a3c84cbd639b56449ad15523c::server:9d30861e1ba54177b8e4dd1426076e5d' INSTALL_K3S_VERSION=1.18 INSTALL_K3S_EXEC='server --server https://172.27.251.164:6443 --tls-san 127.0.0.1' sh -s - --node-taint key=value:NoExecute",
			k3sExtraArgs:   "--node-taint key=value:NoExecute",
			serverAgent:    true,
			tlsSAN:         "127.0.0.1",
		},
		{
			title:          "Join agent with K3sExtraArgs",
			serverIP:       "172.27.251.164",
			joinToken:      "K10c8bc21f68fef3f56d431a08df2e894481ab0a61a3c84cbd639b56449ad15523c::server:9d30861e1ba54177b8e4dd1426076e5d",
			installStr:     "INSTALL_K3S_VERSION=1.18",
			installk3sExec: "K3S_URL='https://172.27.251.164:6443' K3S_TOKEN='K10c8bc21f68fef3f56d431a08df2e894481ab0a61a3c84cbd639b56449ad15523c::server:9d30861e1ba54177b8e4dd1426076e5d' INSTALL_K3S_VERSION=1.18 sh -s - --node-ip=192.0.3.4 --node-external-ip=85.159.215.50",
			k3sExtraArgs:   "--node-ip=192.0.3.4 --node-external-ip=85.159.215.50",
			serverAgent:    false,
			tlsSAN:         "127.0.0.1",
		},
	}
	for _, tc := range tests {
		t.Run(tc.title, func(t *testing.T) {
			got := makeJoinExec(tc.serverIP, tc.joinToken, tc.installStr, tc.k3sExtraArgs, tc.serverAgent, "", tc.tlsSAN)

			if got != tc.installk3sExec {
				t.Errorf("want:\n%s\n, got:\n%s\n", tc.installk3sExec, got)
			}
		})
	}

}

func Test_makeJoinAgentExec(t *testing.T) {
	tests := []test{
		{
			title:          "Join Agent without K3sExtraArgs",
			serverIP:       "172.27.251.164",
			joinToken:      "K10c8bc21f68fef3f56d431a08df2e894481ab0a61a3c84cbd639b56449ad15523c::server:9d30861e1ba54177b8e4dd1426076e5d",
			installStr:     "INSTALL_K3S_VERSION=1.18",
			k3sExtraArgs:   "",
			serverAgent:    false,
			installk3sExec: "K3S_URL='https://172.27.251.164:6443' K3S_TOKEN='K10c8bc21f68fef3f56d431a08df2e894481ab0a61a3c84cbd639b56449ad15523c::server:9d30861e1ba54177b8e4dd1426076e5d' INSTALL_K3S_VERSION=1.18 sh -s -",
		},
		{
			title:          "Join Agent with K3sExtraArgs",
			serverIP:       "172.27.251.164",
			joinToken:      "K10c8bc21f68fef3f56d431a08df2e894481ab0a61a3c84cbd639b56449ad15523c::server:9d30861e1ba54177b8e4dd1426076e5d",
			installStr:     "INSTALL_K3S_VERSION=1.18",
			installk3sExec: "K3S_URL='https://172.27.251.164:6443' K3S_TOKEN='K10c8bc21f68fef3f56d431a08df2e894481ab0a61a3c84cbd639b56449ad15523c::server:9d30861e1ba54177b8e4dd1426076e5d' INSTALL_K3S_VERSION=1.18 sh -s - --node-taint key=value:NoExecute",
			k3sExtraArgs:   "--node-taint key=value:NoExecute",
			serverAgent:    false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.title, func(t *testing.T) {
			got := makeJoinExec(tc.serverIP, tc.joinToken, tc.installStr, tc.k3sExtraArgs, tc.serverAgent, "", "")

			if got != tc.installk3sExec {
				t.Errorf("want: %s, got: %s", tc.installk3sExec, got)
			}
		})
	}
}
