package cmd

import (
	"github.com/lthummus/protohackers/problems/speeding"
	"github.com/spf13/cobra"
)

var speedCommand = &cobra.Command{
	Use: "speed",
	Run: func(cmd *cobra.Command, args []string) {
		//s := speeding.NewSystem()
		//s.RegisterRoad(7455, 10000)
		//s.RecordObservation("E228XX", 7455, 511, 30687215)
		//s.RecordObservation("E228XX", 7455, 739, 30695423)

		speeding.RunSpeeding(port)
	},
}
