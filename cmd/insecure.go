package cmd

import (
	"github.com/lthummus/protohackers/problems/insecure"
	"github.com/spf13/cobra"
)

var insecureCommand = &cobra.Command{
	Use: "insecure",
	Run: func(cmd *cobra.Command, args []string) {
		insecure.RunInsecure(port)
	},
}
