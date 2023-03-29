package cmd

import (
	"fmt"
	"os"
	"time"

	"github.com/spf13/cobra"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

var port int

func init() {
	log.Logger = log.Output(zerolog.ConsoleWriter{
		Out:        os.Stderr,
		TimeFormat: time.RFC3339,
	})
	zerolog.SetGlobalLevel(zerolog.TraceLevel)

	rootCmd.PersistentFlags().IntVarP(&port, "port", "p", 28172, "port to listen on")

	rootCmd.AddCommand(smokeTestCmd, primeCommand, meansCommand, chatCommand, databaseCommand, mobCommand, speedCommand, speedTestCmd, reversalCommand)
}

var rootCmd = &cobra.Command{
	Use:   "protohackers",
	Short: "Run protohackers servers",
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err.Error())
		os.Exit(1)
	}
}
