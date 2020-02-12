package apps

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
)

func Test_buildYAML_SubsitutesDomainEmailAndIngress(t *testing.T) {
	templBytes, _ := buildYAML("openfaas.subdomain.example.com", "openfaas@subdomain.example.com", "traefik")

	got := string(templBytes)
	if want != got {
		t.Errorf("suffix, want: %q, got: %q", want, got)
	}
}

var want = `
apiVersion: extensions/v1beta1 
kind: Ingress
metadata:
  name: openfaas-gateway
  namespace: openfaas
  annotations:
    cert-manager.io/cluster-issuer: letsencrypt-prod
    kubernetes.io/ingress.class: traefik
spec:
  rules:
  - host: openfaas.subdomain.example.com
    http:
      paths:
      - backend:
          serviceName: gateway
          servicePort: 8080
        path: /
  tls:
  - hosts:
    - openfaas.subdomain.example.com
    secretName: openfaas-gateway
---
apiVersion: cert-manager.io/v1alpha2
kind: ClusterIssuer
metadata:
  name: letsencrypt-prod
spec:
  acme:
    email: openfaas@subdomain.example.com
    server: https://acme-v02.api.letsencrypt.org/directory
    privateKeySecretRef:
      name: example-issuer-account-key
    solvers:
    - http01:
        ingress:
          class: traefik`

func Test_writeTempFile_writes_to_tmp(t *testing.T) {
	var want = "some input string"
	tmpLocation, _ := writeTempFile([]byte(want), "tmp_file_name.yaml")

	got, _ := ioutil.ReadFile(tmpLocation)
	if string(got) != want {
		t.Errorf("suffix, want: %q, got: %q", want, got)
	}
}

func Test_createTempDirectory_creates(t *testing.T) {
	var want = filepath.Join(os.TempDir(), ".k3sup")

	got, _ := createTempDirectory(".k3sup")

	if got != want {
		t.Errorf("suffix, want: %q, got: %q", want, got)
	}
}
