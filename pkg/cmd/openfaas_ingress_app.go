package cmd

import (
	"bytes"
	"fmt"
	"log"

	"text/template"

	"github.com/spf13/cobra"
)

type InputData struct {
	IngressDomain 	string
	CertmanagerEmail	string
}


func makeInstallOpenFaaSIngress() *cobra.Command {
	var openfaasIngress = &cobra.Command{
		Use:          "openfaas-ingress",
		Short:        "Install openfaas ingress with TLS",
		Long:         `Install openfaas ingress. Requires Certmanager installation in the cluster. Please set --domain to your custom domain and set --email to your email - this email is used by letsencrypt for domain expiry etc.`,
		Example:      `  k3sup app install openfaas-ingress --domain openfaas.example.com --email openfaas@example.com`,
		SilenceUsage: true,
	}

	openfaasIngress.Flags().StringP("domain", "d", "openfaas.example.com", "Custom Ingress Domain")
	openfaasIngress.Flags().StringP("email", "e", "openfaas@example.com", "Letsencrypt Email")

	openfaasIngress.RunE = func(command *cobra.Command, args []string) error {
		kubeConfigPath := getDefaultKubeconfig()

		if command.Flags().Changed("kubeconfig") {
			kubeConfigPath, _ = command.Flags().GetString("kubeconfig")
		}

		fmt.Printf("Using kubeconfig: %s\n", kubeConfigPath)

		// Check if cert-manager and OpenFaaS are installed into the cluster? And in correct namespaces
		namespaceErr := kubectl("get", "namespace", "openfaas", "cert-manager")
		if namespaceErr != nil {
			log.Panic("openfaas and cert-manager namespaces not found on this cluster.")
			return namespaceErr
		}

		email, _ := command.Flags().GetString("email")
		domain, _ := command.Flags().GetString("domain")
		yamlString := buildYaml(domain, email)


		err := kubectlStdIn(yamlString, "apply", "-f", "-")

		if err != nil {
			log.Print(err)
			return err
		}



		fmt.Println(`=======================================================================
= OpenFaaS Ingress and CertManager ClusterIssuer have been installed  =
=======================================================================

# You will need to ensure that your domain points to your cluster and is
# accessible through ports 80 and 443. 
#
# This is used to validate your ownership of this domain by LetsEncrypt
# and then you can use https with your installation. Without a valid DNS

# Ingress to your domain has been installed for OpenFaas
# to see the ingress record run

kubectl get -n openfaas Ingress openfaas-gateway

# A Certmanager ClusterIssuer has been installed into the default
# namespace - to see the resource run
kubectl describe ClusterIssuer letsencrypt-prod

# To check the status of your certificate you can run
kubectl describe -n openfaas Certificate gw-openfaas

#It may take a while to be issued by LetsEncrypt, in the meantime a SelfSigned cert will be installed


Thank you for using k3sup!`)

		return nil
	}

	return openfaasIngress
}

func buildYaml(domain string, email string) []byte {
	tmpl, err := template.New("yaml").Parse(yamlTemplate)

	if err != nil {
		log.Panic("Error loading Yaml Template: ", err)
	}

	inputData := InputData{
		IngressDomain: 		domain,
		CertmanagerEmail:	email,
	}
	var tpl bytes.Buffer

	err = tmpl.Execute(&tpl, inputData)

	if err != nil {
		log.Panic("Error executing template: ", err)
	}

	return tpl.Bytes()
}


var yamlTemplate = `
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
  - host: {{.IngressDomain}}
    http:
      paths:
      - backend:
          serviceName: gateway
          servicePort: 8080
        path: /
  tls:
  - hosts:
    - {{.IngressDomain}}
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
    email: {{.CertmanagerEmail}}
    server: https://acme-v02.api.letsencrypt.org/directory
    privateKeySecretRef:
      # Secret resource used to store the account's private key.
      name: example-issuer-account-key
    # Add a single challenge solver, HTTP01 using nginx
    solvers:
    - http01:
        ingress:
          class: nginx`
