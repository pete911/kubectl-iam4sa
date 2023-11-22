package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
	"log/slog"
	"os"
	"path/filepath"
)

var (
	Version        string
	KubeconfigPath string

	RootCmd = &cobra.Command{}
)

func init() {
	defaultKubeconfig := filepath.Join(homedir.HomeDir(), ".kube", "config")
	RootCmd.PersistentFlags().StringVar(
		&KubeconfigPath,
		"kubeconfig",
		getStringEnv("KUBECONFIG", defaultKubeconfig),
		"path to kubeconfig file",
	)
	// TODO add log level flag
}

func RestConfig() *rest.Config {
	config, err := clientcmd.BuildConfigFromFlags("", KubeconfigPath)
	if err != nil {
		fmt.Printf("build kube config from flags: %v", err)
		os.Exit(1)
	}
	return config
}

func Logger() *slog.Logger {
	// TODO get log level from flag
	opts := &slog.HandlerOptions{Level: slog.LevelInfo}
	return slog.New(slog.NewJSONHandler(os.Stderr, opts))
}

func getStringEnv(envName string, defaultValue string) string {
	if env, ok := os.LookupEnv(envName); ok {
		return env
	}
	return defaultValue
}
