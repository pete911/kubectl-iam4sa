package cmd

import (
	"fmt"
	"github.com/pete911/kubectl-iam4sa/internal/k8s"
	"github.com/spf13/cobra"
	"k8s.io/client-go/util/homedir"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
)

var logLevels = map[string]slog.Level{"debug": slog.LevelDebug, "info": slog.LevelInfo, "warn": slog.LevelWarn, "error": slog.LevelError}

type Flags struct {
	kubeconfigPath string
	logLevel       string
	namespace      string
	allNamespaces  bool
	label          string
	fieldSelector  string
}

func (f Flags) Kubeconfig() k8s.Kubeconfig {
	kubeconfig, err := k8s.NewKubeconfig(f.kubeconfigPath)
	if err != nil {
		fmt.Printf("load kubeconfig %s: %v", f.kubeconfigPath, err)
		os.Exit(1)
	}
	return kubeconfig
}

func (f Flags) Logger() *slog.Logger {
	if level, ok := logLevels[strings.ToLower(f.logLevel)]; ok {
		opts := &slog.HandlerOptions{Level: level}
		return slog.New(slog.NewJSONHandler(os.Stderr, opts))
	}

	fmt.Printf("invalid log level %s", f.logLevel)
	os.Exit(1)
	return nil
}

func (f Flags) Namespace() string {
	if f.allNamespaces {
		return ""
	}
	return f.namespace
}

func (f Flags) Label() string {
	return f.label
}

func (f Flags) FieldSelector(args []string) string {
	for _, v := range args {
		fieldSelectors := strings.Split(f.fieldSelector, ",")
		fieldSelectors = append(fieldSelectors, fmt.Sprintf("metadata.name=%s", v))
		f.fieldSelector = strings.Join(fieldSelectors, ",")
	}
	return f.fieldSelector
}

func InitPersistentFlags(cmd *cobra.Command, flags *Flags) {
	defaultKubeconfig := filepath.Join(homedir.HomeDir(), ".kube", "config")
	cmd.PersistentFlags().StringVar(
		&flags.kubeconfigPath,
		"kubeconfig",
		getStringEnv("KUBECONFIG", defaultKubeconfig),
		"path to kubeconfig file",
	)
	cmd.PersistentFlags().StringVar(
		&flags.logLevel,
		"log-level",
		"warn",
		"log level - debug, info, warn, error",
	)
	cmd.PersistentFlags().StringVarP(
		&flags.namespace,
		"namespace",
		"n",
		"default",
		"kubernetes namespace",
	)
	cmd.PersistentFlags().BoolVarP(
		&flags.allNamespaces,
		"all-namespaces",
		"A",
		false,
		"all kubernetes namespaces",
	)
	cmd.PersistentFlags().StringVarP(
		&flags.label,
		"label",
		"l",
		"",
		"kubernetes label",
	)
	cmd.PersistentFlags().StringVarP(
		&flags.fieldSelector,
		"field-selector",
		"",
		"",
		"kubernetes field selector",
	)
}

func getStringEnv(envName string, defaultValue string) string {
	if env, ok := os.LookupEnv(envName); ok {
		return env
	}
	return defaultValue
}
