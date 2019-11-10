package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

func makeInstallInletsOperator() *cobra.Command {
	var inletsOperator = &cobra.Command{
		Use:          "inlets-operator",
		Short:        "Install inlets-operator",
		Long:         `Install inlets-operator to get public IPs for your cluster`,
		Example:      `  k3sup app install inlets-operator --namespace default`,
		SilenceUsage: true,
	}

	inletsOperator.Flags().StringP("namespace", "n", "default", "The namespace used for installation")
	inletsOperator.Flags().StringP("token-file", "t", "", "Text file for your DigitalOcean token")

	inletsOperator.RunE = func(command *cobra.Command, args []string) error {
		kubeConfigPath := getDefaultKubeconfig()

		if command.Flags().Changed("kubeconfig") {
			kubeConfigPath, _ = command.Flags().GetString("kubeconfig")
		}

		fmt.Printf("Using kubeconfig: %s\n", kubeConfigPath)

		namespace, _ := command.Flags().GetString("namespace")

		if namespace != "default" {
			return fmt.Errorf(`to override the namespace, edit the YAML files on GitHub`)
		}
		secretFileName, _ := command.Flags().GetString("token-file")

		if len(secretFileName) == 0 {
			return fmt.Errorf(`--token-file is a required field for your cloud API token`)
		}

		res, err := kubectlTask("create", "secret", "generic",
			"inlets-access-key",
			"--from-file", "inlets-access-key="+secretFileName)

		if len(res.Stderr) > 0 {
			return fmt.Errorf("Error from kubectl\n%q", res.Stderr)
		}

		if err != nil {
			return err
		}

		yamls := []string{
			"https://raw.githubusercontent.com/inlets/inlets-operator/master/artifacts/crd.yaml",
			"https://raw.githubusercontent.com/inlets/inlets-operator/master/artifacts/operator-rbac.yaml",
			"https://raw.githubusercontent.com/inlets/inlets-operator/master/artifacts/operator.yaml",
		}

		for _, yaml := range yamls {
			err = kubectl("apply", "-f", yaml)

			if err != nil {
				return err
			}
		}

		fmt.Println(`=======================================================================
= inlets-operator has been installed.                                  =
=======================================================================

# The default configuration is for DigitalOcean and your secret is
# stored as "inlets-access-key" in the "default" namespace.

# To get your first Public IP run the following:
kubectl run nginx-1 --image=nginx --port=80 --restart=Always
kubectl expose deployment nginx-1 --port=80 --type=LoadBalancer

# Find your IP in the "EXTERNAL-IP" field, watch for "<pending>" to 
# change to an IP

kubectl get svc -w

# When you're done, remove the tunnel by deleting the service
kubectl delete svc/nginx-1

# Find out more at:
# https://github.com/inlets/inlets-operator

` + thanksForUsing)

		return nil
	}

	return inletsOperator
}
