package main

import (
	"github.com/alexellis/k3sup/pkg/cmd"
	"github.com/spf13/cobra"
)

func main() {

	cmdInstall := cmd.MakeInstall()

	cmdVersion := cmd.MakeVersion()

	var rootCmd = &cobra.Command{Use: "app"}

	rootCmd.AddCommand(cmdInstall)
	rootCmd.AddCommand(cmdVersion)
	rootCmd.Execute()
}
