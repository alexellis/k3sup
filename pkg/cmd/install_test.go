package cmd

import (
	"io/ioutil"
	"os"
	"regexp"
	"strings"
	"testing"
)

const privateKey = `-----BEGIN RSA PRIVATE KEY-----
Proc-Type: 4,ENCRYPTED
DEK-Info: DES-EDE3-CBC,6015676AEB0A96A9

A3HKcrAWcupUNvWZ+zfoMFPuaI76PF2XV2QXXOuz7mh3QQRlygVtMNDJZckZAJkd
Sn28TdfmFOhJ/owJElcxKRBrRE+JbKEIgyAUaKiRrAMPlqvDu2kPn5Jan5HhQfnk
K8Y+WI5dnR2cS3uoB7PkRlZjiJtSJzT3Qw2hO0KoZftWKNuQfBRkrY5+c94veb3X
kX+Ym4H3dHUXcIaYjHrTK+tuC36bzF0sdPQRf94JjtpGP3XkdVvWnmbL3i9XKZ/s
niaqfBleWT/EqfjIaex1JAj7XTlvau4AjKLCOaLZe1BkHEViL1lNQX0PoBVfFNK9
o8oGx8EBdmtxBpL6vSLMJSqEIyv2j+ziTUCjUkRa1O5S0lmWFoEXhz8hZ1GiVg7u
GmM0qN6tv7S9hiPx3x8jeTxaTyeGVs2O4Se3Y5bzdXoxWj0FcRh6DMR8SP/AeUDJ
IWFBbr3vD6nMWKYF4Ego9QRBsyIUL2oQfJk2j65dry+VMeVxcAlt9eQSOlRuxBg5
ySfAwn0bof4uY/I1u53ObnZvUZ1/AtuwK8K5mYDkNUchnoZiUC+v1PuyDowmJJxC
ds/3e4Opcs/T+3dJJ6MDO1STGJwsGd3aUWIeJX/E8USs/D20tLdYdJjiH/ijjp8K
lSTBND/n5CH417m/ta/QMy1e1zRgAKcc0WbdyrAFv5P9E4dZuMa0Ppq/1QjhoY48
WBDTI4J6Jw0muGSRQGIO9FCCH2mU/l/JOQ8+dzeMspYq9CY0tqRI6HweDyKR7nII
9QdL0fOnltgsNyziC6AUOhlDGKVuorIyHiYhOLVY6No4K+RbNE5/Tw==
-----END RSA PRIVATE KEY-----
`

const kubeconfigExample = `
apiVersion: v1
clusters:
- cluster:
    certificate-authority-data: DATA+OMITTED
    server: https://localhost:6443
  name: default
contexts:
- context:
    cluster: default
    user: default
  name: default
current-context: default
kind: Config
preferences: {}
users:
- name: default
  user:
    password: 5ceb3a3e93621d265fd147929f3ace84
    username: admin
`

func Test_loadPublickeyEncrypted(t *testing.T) {
	expected := "x509: decryption password incorrect"

	tmpfile, err := ioutil.TempFile("", "key")
	if err != nil {
		t.Error(err)
	}

	defer os.Remove(tmpfile.Name())
	if _, err := tmpfile.Write([]byte(privateKey)); err != nil {
		t.Error(err)
	}

	tmpfile.Close()
	_, _, err = loadPublickey(tmpfile.Name())
	if err.Error() != expected {
		t.Errorf("Unexpected error, got: %q, want: %q.", err.Error(), expected)
	}
}

func Test_RewriteKubeconfig(t *testing.T) {
	var ip = "192.168.0.25"
	var context = "context-test"

	// Test master ip rewrite
	kubeconfig = rewriteKubeconfig(kubeconfigExample, ip, context)

	re := regexp.MustCompile(`server:\s?https://(.*):\d+`)
	group := re.FindSubmatch(kubeconfig)

	if len(group) == 0 || string(group[1]) != ip {
		t.Errorf("Unexpected error, got: %q, want: %q.", string(group[1]), ip)
	}

	kubeconfigExampleIPLocal := strings.Replace(kubeconfigExample, "localhost", "127.0.0.1", -1)
	kubeconfig = rewriteKubeconfig(kubeconfigExampleIPLocal, ip, context)

	group = re.FindSubmatch(kubeconfig)
	if len(group) == 0 || string(group[1]) != ip {
		t.Errorf("Unexpected error, got: %q, want: %q.", string(group[1]), ip)
	}

	// Test context
	re = regexp.MustCompile(`default`)
	expectedContextsToReplace := re.FindAllStringIndex(kubeconfigExample, -1)

	kubeconfig = rewriteKubeconfig(kubeconfigExample, ip, "")
	match := re.FindAllIndex(kubeconfig, -1)

	if len(match) != len(expectedContextsToReplace) {
		t.Errorf("Unexpected error, got: %q, want: %q.", len(match), len(expectedContextsToReplace))
	}

	kubeconfig = rewriteKubeconfig(kubeconfigExample, ip, context)

	re = regexp.MustCompile(`context-test`)
	match = re.FindAllIndex(kubeconfig, -1)

	if len(match) != len(expectedContextsToReplace) {
		t.Errorf("Unexpected error, got: %q, want: %q.", len(match), len(expectedContextsToReplace))
	}
}

func Test_getHelmURL(t *testing.T) {
	got := getHelmURL("amd64", "darwin", "v2.14.3")
	want := "https://get.helm.sh/helm-v2.14.3-darwin-amd64.tar.gz"

	if want != got {
		t.Errorf("want %s, got %s", want, got)
	}
}
