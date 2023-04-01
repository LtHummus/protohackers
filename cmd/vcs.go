package cmd

import (
	"github.com/lthummus/protohackers/problems/vcs"
	"github.com/spf13/cobra"
)

var vcsCommand = &cobra.Command{
	Use: "vcs",
	Run: func(cmd *cobra.Command, args []string) {
		vcs.RunVCS(port)
	},
}
