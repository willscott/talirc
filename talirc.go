package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "talirc",
	Short: "Talirc is an IRC proxy for Talek",
	Long: `An IRC proxy for the Talek PIR system.
		IRC channels are mapped to a collection of
		mutually subscribed logs by each participant.
		nickserv and chanserv interfaces are locally resolved
		to help with management of contacts and channels.`,
	Run: func(cmd *cobra.Command, args []string) {
		// Do Stuff Here
	},
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
