package cmd

import (
	"fmt"

	"github.com/FreeSign-io/fcli/internal/version"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(&cobra.Command{
		Use:   "version",
		Short: "Print fcli version, commit, build date",
		Run: func(cmd *cobra.Command, _ []string) {
			fmt.Printf("fcli %s\ncommit: %s\nbuilt:  %s\n", version.Version, version.Commit, version.Date)
		},
	})
}
