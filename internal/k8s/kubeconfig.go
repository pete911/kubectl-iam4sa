package k8s

import (
	"fmt"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/tools/clientcmd/api"
)

type Kubeconfig struct {
	RestConfig  *rest.Config
	ClusterName string
	Region      string
	Profile     string
}

func (k Kubeconfig) String() string {
	return fmt.Sprintf("cluster name: %s region %s", k.ClusterName, k.Region)
}

func NewKubeconfig(kubeconfigPath string) (Kubeconfig, error) {
	clientConfig := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(
		&clientcmd.ClientConfigLoadingRules{ExplicitPath: kubeconfigPath},
		nil)

	apiConfig, err := clientConfig.RawConfig()
	if err != nil {
		return Kubeconfig{}, fmt.Errorf("raw config: %v", err)
	}

	restConfig, err := clientConfig.ClientConfig()
	if err != nil {
		return Kubeconfig{}, fmt.Errorf("client configs: %v", err)
	}

	user := apiConfig.Contexts[apiConfig.CurrentContext].AuthInfo
	exec := apiConfig.AuthInfos[user].Exec
	if exec.Command != "aws" {
		if exec.Command == "" {
			return Kubeconfig{}, fmt.Errorf("exec command is not set in current %s contex, cannot determine cluster name and region", apiConfig.CurrentContext)
		}
		return Kubeconfig{}, fmt.Errorf("unexpected exec command %s for current %s contex, expected 'aws'", exec.Command, apiConfig.CurrentContext)
	}

	env := execEnvToMap(exec.Env)
	return Kubeconfig{
		RestConfig:  restConfig,
		ClusterName: getClusterName(exec.Args),
		Region:      getRegion(exec.Args, env),
		Profile:     getProfile(exec.Args, env),
	}, nil
}

func getRegion(args []string, env map[string]string) string {
	if v := getFlagValue(args, "--region"); v != "" {
		return v
	}
	return env["AWS_REGION"]
}

func getProfile(args []string, env map[string]string) string {
	if v := getFlagValue(args, "--profile"); v != "" {
		return v
	}
	return env["AWS_PROFILE"]
}

func getClusterName(args []string) string {
	return getFlagValue(args, "--cluster-name")
}

func getFlagValue(args []string, flag string) string {
	for i := range args {
		if args[i] == flag && len(args) > i+1 {
			return args[i+1]
		}
	}
	return ""
}

func execEnvToMap(env []api.ExecEnvVar) map[string]string {
	out := make(map[string]string)
	for _, v := range env {
		out[v.Name] = v.Value
	}
	return out
}
