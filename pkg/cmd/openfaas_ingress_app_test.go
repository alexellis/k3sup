package cmd

import (
	"io/ioutil"
	"testing"
)

func Test_build_yaml_returns_correct_substitutions(t *testing.T) {
	got := string(buildYaml("openfaas.subdomain.example.com", "openfaas@subdomain.example.com"))
	if want != got {
		t.Errorf("suffix, want:%s, got:%s", want, got)
	}
}


var want = `
apiVersion: extensions/v1beta1 
kind: Ingress
metadata:
  name: openfaas-gateway
  namespace: openfaas
  annotations:
    certmanager.k8s.io/acme-challenge-type: http01
    cert-manager.io/cluster-issuer: letsencrypt-prod
    kubernetes.io/ingress.class: nginx
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
    secretName: gw-openfaas
---
apiVersion: cert-manager.io/v1alpha2
kind: ClusterIssuer
metadata:
  name: letsencrypt-prod
spec:
  acme:
    # You must replace this email address with your own.
    # Let's Encrypt will use this to contact you about expiring
    # certificates, and issues related to your account.
    email: openfaas@subdomain.example.com
    server: https://acme-v02.api.letsencrypt.org/directory
    privateKeySecretRef:
      # Secret resource used to store the account's private key.
      name: example-issuer-account-key
    # Add a single challenge solver, HTTP01 using nginx
    solvers:
    - http01:
        ingress:
          class: nginx`

func Test_writeTempFile_writes_to_tmp(t *testing.T) {
	var expected = ("some input string")
	tmpLocation := writeTempFile([]byte(expected))

	got, _:= ioutil.ReadFile(tmpLocation)
	if string(got) != expected {
		t.Errorf("suffix, want:%s, got:%s", want, got)
	}
}