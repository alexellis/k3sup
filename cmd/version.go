package cmd

import (
	"fmt"

	"github.com/alexellis/k3sup/pkg"
	"github.com/morikuni/aec"
	"github.com/spf13/cobra"
)

var (
	Version   string
	GitCommit string
)

func PrintK3supASCIIArt() {
	k3supLogo := aec.RedF.Apply(k3supFigletStr)
	support := aec.CyanF.Apply(pkg.SupportMessageShort)

	fmt.Print(k3supLogo)

	fmt.Printf("%s\n\n", support)
}

func MakeVersion() *cobra.Command {
	var command = &cobra.Command{
		Use:   "version",
		Short: "Print the version",
		Example: `  k3sup version
` + pkg.SupportMessageShort + `
`,
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

bootstrap K3s over SSH in < 60s 🚀
`
