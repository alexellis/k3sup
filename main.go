package main

import (
	"github.com/alexellis/k3sup/pkg/cmd"
	"github.com/spf13/cobra"
)

func main() {

	cmdInstall := cmd.MakeInstall()

	cmdVersion := cmd.MakeVersion()

	cmdJoin := cmd.MakeJoin()

	var rootCmd = &cobra.Command{Use: "k3sup"}

	rootCmd.AddCommand(cmdInstall)
	rootCmd.AddCommand(cmdVersion)
	rootCmd.AddCommand(cmdJoin)

	rootCmd.Execute()
}
