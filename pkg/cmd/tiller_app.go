package cmd

import (
	"fmt"
	"log"
	"os"
	"path"
	"strconv"

	execute "github.com/alexellis/go-execute/pkg/v1"

	"github.com/alexellis/k3sup/pkg/config"

	"github.com/spf13/cobra"
)

func addTillerSA(accountName string, namespace string) error {
        if namespace != "kube-system" {
                if _, err := kubectlTask("create", "namespace", namespace); err != nil {
                        return err
                }
        }
        if _, err := kubectlTask("-n", namespace, "create", "sa", accountName); err != nil {
                return err
        } else if _, err = kubectlTask("-n", namespace, "create", "clusterrolebinding", accountName, "--clusterrole",
                "cluster-admin", "--serviceaccount="+namespace+":"+accountName); err != nil {
                return err
        }
        return nil
}

func addTillerRestrictedSA(ns string, otherns string) error {
        var argsList [][]string
        if len(otherns) == 0 {
                argsList = [][]string{
                        //Create ns
                        {"create", "namespace", ns},
                        //Create serviceaccount
                        {"-n", ns, "create", "sa", "tiller"},
                        //Create role
                        {"-n", ns, "create", "role", "tiller-manager", "--verb=*",
                                "--resource=*.,*.apps,*.batch,*.extensions"},
                        //Create rolebinding
                        {"-n", ns, "create", "rolebinding", "tiller-binding", "--role=tiller-manager",
                                "--serviceaccount="+ns+":tiller"},
                }
        } else {
                argsList = [][]string{
                        //Create nses
                        {"create", "namespace", ns},
                        {"create", "namespace", otherns},
                        //Create serviceaccount
                        {"-n", ns, "create", "sa", "tiller"},
                        //Create role
                        {"-n", otherns, "create", "role", "tiller-manager", "--verb=*",
                                "--resource=*.,*.apps,*.batch,*.extensions"},
                        //Bind it
                        {"-n", otherns, "create", "rolebinding", "tiller-binding", "--role=tiller-manager",
                                "--serviceaccount="+ns+":tiller"},
                        //Create configmap access role
                        {"-n", ns, "create", "role", "tiller-manager", "--verb=*",
                                "--resource=configmaps"},
                        //Bind it
                        {"-n", ns, "create", "rolebinding", "tiller-binding", "--role=tiller-manager",
                                "--serviceaccount="+ns+":tiller"},
                }
        }
        for _, args := range(argsList) {
                if _, err := kubectlTask(args...); err != nil {
                        return err
                }
        }
        return nil
}


func makeInstallTiller() *cobra.Command {
	var tiller = &cobra.Command{
		Use:          "tiller",
		Short:        "Install tiller",
		Long:         `Install tiller`,
		Example:      `  k3sup app install tiller --insecure`,
		SilenceUsage: true,
	}

	tiller.Flags().Bool("insecure", true, "Deploy tiller in kube-system namespace with cluster-admin role, no TLS and plaintext configmap storage (ignores all other flags)")

	tiller.Flags().Bool("restricted", false, "Deploy tiller in a namespace with restricted RBAC")

        tiller.Flags().String("namespace", "ketchup", "Namespace of tiller")

        tiller.Flags().Bool("same-ns", true, "Set RBAC permissions to restrict tiller to be able to deploy in the same namespace it is deployed in")
        tiller.Flags().String("other-ns", "", "Set RBAC permissions to restrict tiller to deploy only in the specified namespace")

        tiller.Flags().Bool("secret-storage", true, "Use secret storage for tiller")
        tiller.Flags().Int("history-max", 200, "limit the maximum number of revisions saved per release. Use 0 for no limit")

        tiller.RunE = func(cmd *cobra.Command, args []string) error {
		kubeConfigPath := getDefaultKubeconfig()

		if cmd.Flags().Changed("kubeconfig") {
			kubeConfigPath, _ = cmd.Flags().GetString("kubeconfig")
		}

		fmt.Printf("Using kubeconfig: %s\n", kubeConfigPath)

		arch := getArchitecture()
		fmt.Printf("Node architecture: %q\n", arch)

		if arch != "x86_64" && arch != "amd64" {
			return fmt.Errorf("This app is not known to work with the %s architecture", arch)
		}

		userPath, err := config.InitUserDir()
		if err != nil {
			return err
		}

		clientArch, clientOS := getClientArch()

		fmt.Printf("Client: %q, %q\n", clientArch, clientOS)

		log.Printf("User dir established as: %s\n", userPath)

		os.Setenv("HELM_HOME", path.Join(userPath, ".helm"))

		var helmInitFlags []string

		if cmd.Flags().Changed("restricted") {
			exec_args := []string{"init"}

			ns, err := cmd.Flags().GetString("namespace")
			if err != nil {
				return err
			}
			exec_args = append(exec_args, "--tiller-namespace", ns)

			if ok, err := cmd.Flags().GetBool("secret-storage"); err != nil && ok {
				exec_args = append(exec_args, "--override", "'spec.template.spec.containers[0].command'='{/tiller,--storage=secret}'")
			}

			if arg, err := cmd.Flags().GetInt("history-max"); err != nil {
				return err
			} else {
				exec_args = append(exec_args, "--history-max", strconv.Itoa(arg))
			}

			if cmd.Flags().Changed("other-ns") {
				if otherns, err := cmd.Flags().GetString("other-ns"); err != nil {
					return err
				} else {
					if err := addTillerRestrictedSA(ns, otherns); err != nil {
						return err
					}
				}
			} else {
				ok, err := cmd.Flags().GetBool("same-ns")
				if err != nil {
					return err
				}
				if !ok {
					fmt.Println("You need to specify --same-ns (default) or --other-ns <some-namespace>")
					return nil
				}
				if err := addTillerRestrictedSA(ns, ""); err != nil {
					return err
				}
			}
			helmInitFlags = []string{"init", "--skip-refresh", "--upgrade", "--service-account", "tiller", "--tiller-namespace", ns}
		} else if ! cmd.Flags().Changed("insecure") {
			if err := addTillerSA("tiller", "kube-system"); err != nil {
				return err
			}
			helmInitFlags = []string{"init", "--skip-refresh", "--upgrade", "--service-account", "tiller"}
		} else {
			fmt.Println("You must choose --insecure or --restricted deployment !")
			return nil
		}

		helmBinary, err := tryDownloadHelm(userPath, clientArch, clientOS)
		if err != nil {
			return err
		}

		k3supBin := path.Join(userPath, ".bin")
		helmInit := execute.ExecTask{
			Command: path.Join(k3supBin, "helm"),
			Args: helmInitFlags,
		}
		res, err := helmInit.Execute()
		if err != nil {
			return err
		}

		fmt.Println(res.Stdout, res.Stderr)

		fmt.Println(`=======================================================================
tiller has been installed
=======================================================================

# You can now use helm with tiller from the installation directory

` + helmBinary + `

Thank you for using k3sup!`)

		return nil
        }

	return tiller
}
