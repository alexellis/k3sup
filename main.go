package main

import (
	"os"

	"github.com/alexellis/k3sup/cmd"
	"github.com/alexellis/k3sup/pkg"
	"github.com/spf13/cobra"
)

func main() {

	cmdInstall := cmd.MakeInstall()
	cmdVersion := cmd.MakeVersion()
	cmdJoin := cmd.MakeJoin()
	cmdUpdate := cmd.MakeUpdate()
	cmdReady := cmd.MakeReady()
	cmdPlan := cmd.MakePlan()
	cmdNodeToken := cmd.MakeNodeToken()
	cmdGetConfig := cmd.MakeGetConfig()
	cmdGet := cmd.MakeGet()
	cmdGetPro := cmd.MakeGetPro()
	cmdPro := cmd.MakePro()

	printk3supASCIIArt := cmd.PrintK3supASCIIArt

	var rootCmd = &cobra.Command{
		Use: "k3sup",
		Run: func(cmd *cobra.Command, args []string) {
			printk3supASCIIArt()
			cmd.Help()
		},
		Long: pkg.SupportMessageShort,
	}

	rootCmd.AddCommand(cmdInstall)
	rootCmd.AddCommand(cmdVersion)
	rootCmd.AddCommand(cmdJoin)
	rootCmd.AddCommand(cmdUpdate)
	rootCmd.AddCommand(cmdReady)
	rootCmd.AddCommand(cmdPlan)
	rootCmd.AddCommand(cmdNodeToken)
	rootCmd.AddCommand(cmdGetConfig)

	cmdGet.AddCommand(cmdGetPro)
	rootCmd.AddCommand(cmdGet)
	rootCmd.AddCommand(cmdPro)

	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
