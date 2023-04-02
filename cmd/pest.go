package cmd

import (
	"github.com/lthummus/protohackers/problems/pest"
	"github.com/spf13/cobra"
)

var pestCommand = &cobra.Command{
	Use: "pest",
	Run: func(cmd *cobra.Command, args []string) {
		pest.RunPest(port)
	},
}
