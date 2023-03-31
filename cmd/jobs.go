package cmd

import (
	"github.com/lthummus/protohackers/problems/jobs"
	"github.com/spf13/cobra"
)

var jobsCommand = &cobra.Command{
	Use: "jobs",
	Run: func(cmd *cobra.Command, args []string) {
		jobs.RunJobs(port)
	},
}
