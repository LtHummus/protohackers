package cmd

import (
	"github.com/lthummus/protohackers/problems/database"
	"github.com/spf13/cobra"
)

var databaseCommand = &cobra.Command{
	Use: "database",
	Run: func(cmd *cobra.Command, args []string) {
		database.RunDatabase(port)
	},
}
