package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/tools/clientcmd/api"
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

type Kubeconfig struct {
	RestConfig *rest.Config
	Exec       Exec
}

func (k Kubeconfig) Region() string {
	for i := range k.Exec.Args {
		if k.Exec.Args[i] == "--region" {
			return k.Exec.Args[i+1]
		}
	}
	return k.Exec.Env["AWS_REGION"]
}

func (k Kubeconfig) Profile() string {
	for i := range k.Exec.Args {
		if k.Exec.Args[i] == "--profile" {
			return k.Exec.Args[i+1]
		}
	}
	return k.Exec.Env["AWS_PROFILE"]
}

func (k Kubeconfig) ClusterName() string {
	for i := range k.Exec.Args {
		if k.Exec.Args[i] == "--cluster-name" {
			return k.Exec.Args[i+1]
		}
	}
	return ""
}

type Exec struct {
	Command string
	Args    []string
	Env     map[string]string
}

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

func GetKubeconfig() Kubeconfig {
	clientConfig := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(
		&clientcmd.ClientConfigLoadingRules{ExplicitPath: KubeconfigPath},
		nil)

	apiConfig, err := clientConfig.RawConfig()
	if err != nil {
		fmt.Printf("raw config %s: %v", KubeconfigPath, err)
		os.Exit(1)
	}

	restConfig, err := clientConfig.ClientConfig()
	if err != nil {
		fmt.Printf("client config %s: %v", KubeconfigPath, err)
		os.Exit(1)
	}

	user := apiConfig.Contexts[apiConfig.CurrentContext].AuthInfo
	exec := apiConfig.AuthInfos[user].Exec
	return Kubeconfig{
		RestConfig: restConfig,
		Exec: Exec{
			Command: exec.Command,
			Args:    exec.Args,
			Env:     toEnv(exec.Env),
		},
	}
}

func toEnv(env []api.ExecEnvVar) map[string]string {
	out := make(map[string]string)
	for _, v := range env {
		out[v.Name] = v.Value
	}
	return out
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
