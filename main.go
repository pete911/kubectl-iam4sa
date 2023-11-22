package main

import (
	"github.com/pete911/kubectl-iam4sa/cmd"
	"os"
)

var Version = "dev"

func main() {
	cmd.Version = Version
	if err := cmd.RootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
