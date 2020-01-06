package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/willscott/talirc/services"
)

var (
	port        int
	talekConfig string
	rootCmd     = &cobra.Command{
		Use:   "talirc",
		Short: "Talirc is an IRC proxy for Talek",
		Long: `An IRC proxy for the Talek PIR system.
		IRC channels are mapped to a collection of
		mutually subscribed logs by each participant.
		nickserv and chanserv interfaces are locally resolved
		to help with management of contacts and channels.`,
		Run: func(cmd *cobra.Command, args []string) {
			if err := services.Serve(port, talekConfig); err != nil {
				fmt.Fprintf(os.Stderr, "%v\n", err)
				os.Exit(1)
			}
		},
	}
)

func init() {
	rootCmd.PersistentFlags().IntVar(&port, "port", 5222, "listening port for talirc")
	rootCmd.PersistentFlags().StringVarP(&talekConfig, "config", "c", "", "Talek config file")
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
