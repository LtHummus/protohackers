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
var verbose bool

func init() {
	log.Logger = log.Output(zerolog.ConsoleWriter{
		Out:        os.Stderr,
		TimeFormat: time.RFC3339,
	})

	rootCmd.PersistentFlags().IntVarP(&port, "port", "p", 28172, "port to listen on")
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "verbose logging")

	rootCmd.AddCommand(smokeTestCmd, primeCommand, meansCommand, chatCommand, databaseCommand, mobCommand, speedCommand, speedTestCmd, reversalCommand, insecureCommand, jobsCommand)

	cobra.OnInitialize(func() {
		if verbose {
			zerolog.SetGlobalLevel(zerolog.TraceLevel)
		} else {
			zerolog.SetGlobalLevel(zerolog.WarnLevel)
		}
		log.Trace().Msg("verbose logging enabled")

	})
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
