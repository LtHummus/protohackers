package cmd

import (
	"github.com/lthummus/protohackers/problems/means"
	"github.com/spf13/cobra"
)

var meansCommand = &cobra.Command{
	Use: "means",
	Run: func(cmd *cobra.Command, args []string) {
		means.RunMeans(port)
	},
}
