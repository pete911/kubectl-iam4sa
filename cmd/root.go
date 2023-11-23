package cmd

import (
	"github.com/spf13/cobra"
)

var (
	RootCmd = &cobra.Command{}

	Version     string
	GlobalFlags Flags
)

func init() {
	InitPersistentFlags(RootCmd, &GlobalFlags)
}
