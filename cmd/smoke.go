package cmd

import (
	"github.com/lthummus/protohackers/problems/smoke"
	"github.com/spf13/cobra"
)

var smokeTestCmd = &cobra.Command{
	Use: "smoke",
	Run: func(cmd *cobra.Command, args []string) {
		smoke.RunSmoke(port)
	},
}
