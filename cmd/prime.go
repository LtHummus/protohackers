package cmd

import (
	"github.com/lthummus/protohackers/problems/prime"
	"github.com/spf13/cobra"
)

var primeCommand = &cobra.Command{
	Use: "prime",
	Run: func(cmd *cobra.Command, args []string) {
		prime.RunPrime(port)
	},
}
