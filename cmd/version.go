package cmd

import (
	"fmt"

	"github.com/morikuni/aec"
	"github.com/spf13/cobra"
)

var (
	// Version stores the version of the build
	Version string
	// GitCommit stores the git commit of the build
	GitCommit string
)

// PrintK3supASCIIArt prints the ascii art of k3sup
func PrintK3supASCIIArt() {
	k3supLogo := aec.RedF.Apply(k3supFigletStr)
	fmt.Print(k3supLogo)
}

// MakeVersion returns the version sub command of k3sup
func MakeVersion() *cobra.Command {
	var command = &cobra.Command{
		Use:          "version",
		Short:        "Print the version",
		Example:      `  k3sup version`,
		SilenceUsage: false,
	}
	command.Run = func(cmd *cobra.Command, args []string) {
		PrintK3supASCIIArt()
		if len(Version) == 0 {
			fmt.Println("Version: dev")
		} else {
			fmt.Println("Version:", Version)
		}
		fmt.Println("Git Commit:", GitCommit)
	}
	return command
}

const k3supFigletStr = ` _    _____                 
| | _|___ / ___ _   _ _ __  
| |/ / |_ \/ __| | | | '_ \ 
|   < ___) \__ \ |_| | |_) |
|_|\_\____/|___/\__,_| .__/ 
                     |_|    
`
