package cmd

import (
	"github.com/lthummus/protohackers/problems/chat"
	"github.com/spf13/cobra"
)

var chatCommand = &cobra.Command{
	Use: "chat",
	Run: func(cmd *cobra.Command, args []string) {
		chat.RunChat(port)
	},
}
