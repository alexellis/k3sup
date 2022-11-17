package cmd

import (
	"errors"
	"io/ioutil"
	"os"
	"regexp"
	"strings"
	"testing"

	"golang.org/x/crypto/ssh"
)

// To regenerate:
// openssl genrsa -des3 -out /tmp/id_rsa_encrypted 2048
const privateKey = `-----BEGIN RSA PRIVATE KEY-----
Proc-Type: 4,ENCRYPTED
DEK-Info: DES-EDE3-CBC,4FA02C824DA7DA35

0MxJJqh8FGEciV4fkZzq7bCfKmPy5a3x9eJ+8sY+ssNGG8cLMdV2uDPMrarssjBK
QtaEFUMu2f2lxuXYvPzZtSQNkcUUj2kBJCxgdrs7mGLgqnLOYkbkWA3rUiiNYf2S
UofpsAO4gVWcQ4HBtnWW6skQp4fa0fdfg8elKlOrM0wcRX890attpyCCbfEAtn/v
6z31ezObnGQOFSZ9kM7icbAM8pPgjex3kno5kxzVfpyL+5pq36AyFoFBN1wzYzMf
gpwtr/m+Kw/KUWVGFKXLKgFSe9aF/dIXisdVCJze+2uQEBDXZo1OiJ3rflrTKwlB
t2NNLPdd23MOHK8B8dBWhitppBqloy68Thfw0cFq2E4qtUIl9SqtBhcQ118E0Z/7
UGVg6Ki+sgBO5fHcQUnDn7DGV8/Gawl+ZOhvGkD9C2Q/vK/SaKEg3X7ap4Oo2/em
NGVVnxCpgd3GfVHsZFHjRvt/YYQtBHdAhH8cU45WlRaLbUyKSyfq7TIAzDa04bl+
VMytKkfwYoPG8E3POABZ2lOgDWeBQeK3eP7EkxTkv3sSahWIwFE1HaBZmhoL27vH
necstfLEHqKkONvaXzqSbKk7e0GXKooSgZS2NJP7wJSX4e5CbOBrM4hf69CIG5bm
rPPYs9mhxsa+iP4X5EVxdr1IEUTzwqeLB+/e/C/+mbs37L7tv3yvKC8UG8gXr1jC
qzm+V3SSH9W5tgqDx97ljuDqLXgZl158W5NbYIwXB7FazU1DEJAOGSgu21w8XqlF
SmxKXJHAjLwGzkygNGZYRGllq8GppZxLeUZHmlL+F490BclIdxCaOir//Bqd+a8a
bs3Q7D57kuo003x9z0e044anmANdmEFSjfPG7ajHUfm7EqsQ4pZOYp4twllOJkY5
Yffoe94wdYbMGtrBKY9xeZPgecDZpjMv1g6pB5Gt6p4VLz/U2rstTkjqiHTZfIej
tphVIzOTpsfVNMG4As3WOapz+9MH2kzEKORAHpQpZenyvcAfhJJa404riZ3HJ++O
Nmc7ASSirGNty1BTJKKQtN/QDvVbM011jUpuQxbEwfUDAUlQU4g5YElfMw3l9tDo
jWM4jimYxGaeaTI2C6hjy7pLMWCywkOGrKVKuii8EI8vd4Mw9jTIMRQzBotzEBFn
qAy3PMlnGd/CDs/HPAWqPWEloU9bcY8oP954EEfNZoNz95u6VMJkqfM/ynu1yBEl
FjG6pf31NEqjYTeFmJROozLGLxdPTrchn/MYU60oG/eJfY+eZ02h8J58yC67aG/f
7tCUfB8UrQH1s16BY2j9EM6KPbX3Hh8VXiKb7/UzIPtD9aD5HKzl7K3fIbi+aQcX
ySQXENXiPpieDZj7kKp9VskNjpLyXyR1BN7Tf3eIZ6N1gK1d6esZMhXhPR5S4LT9
F8ZD5KeHVWB6hOaodWr/bhfEVb8E67/OcnLQM8iKdBfqkoPDInVIXGkt8FXQfiB0
I+rSXfppnf7bhQK3HLeU27Ca6zxQYZ7TI6bXTRBjozFakKkQ+8xcfCVzZ/0/oZgu
kfFJfrUjElq6Bx9oPPxc2vD40gqnYL57A+Y+X+A0kL4fO7pfh2VxOw==
-----END RSA PRIVATE KEY-----
`

func Test_loadPublickeyEncrypted(t *testing.T) {
	want := &ssh.PassphraseMissingError{}

	tmpfile, err := ioutil.TempFile("", "key")
	if err != nil {
		t.Error(err)
	}

	fileName := tmpfile.Name()
	defer os.Remove(fileName)
	if _, err := tmpfile.Write([]byte(privateKey)); err != nil {
		t.Fatalf("unable to write test file %s, %s", fileName, err)
	}

	tmpfile.Close()
	_, _, err = loadPublickey(fileName)
	if errors.Is(err, want) {
		t.Fatalf("want: %q, but got: %q", want, err.Error())
	}
}

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

func Test_RewriteKubeconfig(t *testing.T) {
	var ip = "192.168.0.25"
	var context = "context-test"

	// Test master ip rewrite
	kubeconfig = rewriteKubeconfig(kubeconfigExample, ip, context)

	re := regexp.MustCompile(`server:\s?https://(.*):\d+`)
	group := re.FindSubmatch(kubeconfig)

	if len(group) == 0 || string(group[1]) != ip {
		t.Fatalf("unexpected error, got: %q, want: %q.", string(group[1]), ip)
	}

	kubeconfigExampleIPLocal := strings.Replace(kubeconfigExample, "localhost", "127.0.0.1", -1)
	kubeconfig = rewriteKubeconfig(kubeconfigExampleIPLocal, ip, context)

	group = re.FindSubmatch(kubeconfig)
	if len(group) == 0 || string(group[1]) != ip {
		t.Fatalf("unexpected error, got: %q, want: %q.", string(group[1]), ip)
	}

	// Test context
	re = regexp.MustCompile(`default`)
	expectedContextsToReplace := re.FindAllStringIndex(kubeconfigExample, -1)

	kubeconfig = rewriteKubeconfig(kubeconfigExample, ip, "")
	match := re.FindAllIndex(kubeconfig, -1)

	if len(match) != len(expectedContextsToReplace) {
		t.Fatalf("unexpected error, got: %q, want: %q.", len(match), len(expectedContextsToReplace))
	}

	kubeconfig = rewriteKubeconfig(kubeconfigExample, ip, context)

	re = regexp.MustCompile(`context-test`)
	match = re.FindAllIndex(kubeconfig, -1)

	if len(match) != len(expectedContextsToReplace) {
		t.Fatalf("unexpected error, got: %q, want: %q.", len(match), len(expectedContextsToReplace))
	}
}

func Test_makeInstallExec(t *testing.T) {
	cluster := false
	datastore := ""
	flannelIPSec := false
	k3sNoExtras := false
	k3sExtraArgs := ""
	ip := "raspberrypi.local"
	tlsSAN := ""
	got := makeInstallExec(cluster, ip, tlsSAN,
		k3sExecOptions{
			Datastore:    datastore,
			FlannelIPSec: flannelIPSec,
			NoExtras:     k3sNoExtras,
			ExtraArgs:    k3sExtraArgs,
		})
	want := "INSTALL_K3S_EXEC='server --tls-san raspberrypi.local'"
	if got != want {
		t.Errorf("want: %q, got: %q", want, got)
	}
}

func Test_makeInstallExec_Cluster(t *testing.T) {
	cluster := true
	datastore := ""
	flannelIPSec := false
	k3sNoExtras := false
	k3sExtraArgs := ""
	ip := "127.0.0.1"
	tlsSAN := ""
	got := makeInstallExec(cluster, ip, tlsSAN,
		k3sExecOptions{
			Datastore:    datastore,
			FlannelIPSec: flannelIPSec,
			NoExtras:     k3sNoExtras,
			ExtraArgs:    k3sExtraArgs,
		})
	want := "INSTALL_K3S_EXEC='server --cluster-init --tls-san 127.0.0.1'"
	if got != want {
		t.Errorf("want: %q, got: %q", want, got)
	}
}

func Test_makeInstallExec_SAN(t *testing.T) {
	cluster := false
	datastore := ""
	flannelIPSec := false
	k3sNoExtras := false
	k3sExtraArgs := ""
	ip := "127.0.0.1"
	tlsSAN := "192.168.0.1"
	got := makeInstallExec(cluster, ip, tlsSAN,
		k3sExecOptions{
			Datastore:    datastore,
			FlannelIPSec: flannelIPSec,
			NoExtras:     k3sNoExtras,
			ExtraArgs:    k3sExtraArgs,
		})
	want := "INSTALL_K3S_EXEC='server --tls-san 192.168.0.1'"
	if got != want {
		t.Errorf("want: %q, got: %q", want, got)
	}
}

func Test_makeInstallExec_IPSec(t *testing.T) {
	cluster := false
	datastore := ""
	flannelIPSec := true
	k3sNoExtras := false
	k3sExtraArgs := ""
	ip := "127.0.0.1"
	tlsSAN := ""
	got := makeInstallExec(cluster, ip, tlsSAN,
		k3sExecOptions{
			Datastore:    datastore,
			FlannelIPSec: flannelIPSec,
			NoExtras:     k3sNoExtras,
			ExtraArgs:    k3sExtraArgs,
		})
	want := "INSTALL_K3S_EXEC='server --tls-san 127.0.0.1 --flannel-backend ipsec'"
	if got != want {
		t.Errorf("want: %q, got: %q", want, got)
	}
}

func Test_makeInstallExec_Datastore(t *testing.T) {
	cluster := false
	datastore := "mysql://doadmin:show-password@tcp(db-mysql-lon1-40939-do-user-2197152-0.b.db.ondigitalocean.com:25060)/defaultdb"
	flannelIPSec := false
	k3sNoExtras := false
	k3sExtraArgs := ""
	token := "this-token"
	ip := "127.0.0.1"
	tlsSAN := "192.168.0.1"
	got := makeInstallExec(cluster, ip, tlsSAN,
		k3sExecOptions{
			Datastore:    datastore,
			Token:        token,
			FlannelIPSec: flannelIPSec,
			NoExtras:     k3sNoExtras,
			ExtraArgs:    k3sExtraArgs,
		})
	want := "INSTALL_K3S_EXEC='server --tls-san 192.168.0.1 --datastore-endpoint mysql://doadmin:show-password@tcp(db-mysql-lon1-40939-do-user-2197152-0.b.db.ondigitalocean.com:25060)/defaultdb --token this-token'"
	if got != want {
		t.Errorf("want: %q, got: %q", want, got)
	}
}

func Test_makeInstallExec_Datastore_NoExtras(t *testing.T) {
	cluster := false
	datastore := "mysql://doadmin:show-password@tcp(db-mysql-lon1-40939-do-user-2197152-0.b.db.ondigitalocean.com:25060)/defaultdb"
	flannelIPSec := false
	k3sNoExtras := true
	token := "this-token"
	k3sExtraArgs := ""
	ip := "raspberrypi.local"
	tlsSAN := "192.168.0.1"
	got := makeInstallExec(cluster, ip, tlsSAN,
		k3sExecOptions{
			Datastore:    datastore,
			Token:        token,
			FlannelIPSec: flannelIPSec,
			NoExtras:     k3sNoExtras,
			ExtraArgs:    k3sExtraArgs,
		})
	want := "INSTALL_K3S_EXEC='server --tls-san 192.168.0.1 --datastore-endpoint mysql://doadmin:show-password@tcp(db-mysql-lon1-40939-do-user-2197152-0.b.db.ondigitalocean.com:25060)/defaultdb --token this-token --disable servicelb --disable traefik'"
	if got != want {
		t.Errorf("want: %q, got: %q", want, got)
	}
}
