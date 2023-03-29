package cmd

import (
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

var reversalCommand = &cobra.Command{
	Use: "reversal",
	Run: func(cmd *cobra.Command, args []string) {
		log.Info().Msg("coming soon")
	},
}
