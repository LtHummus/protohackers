package cmd

import (
	"github.com/lthummus/protohackers/problems/reversal"
	"github.com/spf13/cobra"
)

var reversalCommand = &cobra.Command{
	Use: "reversal",
	Run: func(cmd *cobra.Command, args []string) {
		reversal.RunReversal(port)
	},
}
