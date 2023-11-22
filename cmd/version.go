package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
)

var (
	cmdVersion = &cobra.Command{
		Use:   "version",
		Short: "print version",
		Long:  "",
		Run:   runVersionCmd,
	}
)

func init() {
	RootCmd.AddCommand(cmdVersion)
}

func runVersionCmd(_ *cobra.Command, _ []string) {
	fmt.Println(Version)
}
