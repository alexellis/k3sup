package cmd

import (
	"github.com/alexellis/k3sup/pkg"
	"github.com/spf13/cobra"
)

// MakeGet creates the get parent command
func MakeGet() *cobra.Command {
	var command = &cobra.Command{
		Use:   "get",
		Short: "Helper for downloading K3sup Pro",
		Long: `Helper for downloading K3sup Pro.

` + pkg.SupportMessageShort + `
`,
		SilenceUsage: true,
	}

	return command
}
