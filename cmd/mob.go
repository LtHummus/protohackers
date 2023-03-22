package cmd

import (
	"github.com/lthummus/protohackers/problems/mob"
	"github.com/spf13/cobra"
)

var upstream string

func init() {
	mobCommand.Flags().StringVarP(&upstream, "upstream", "u", "chat.protohackers.com:16963", "upstream server to proxy")
}

var mobCommand = &cobra.Command{
	Use: "mob",
	Run: func(cmd *cobra.Command, args []string) {
		mob.RunMob(port, upstream)
	},
}
