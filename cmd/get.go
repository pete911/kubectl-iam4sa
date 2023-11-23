package cmd

import (
	"encoding/json"
	"fmt"
	"github.com/pete911/kubectl-iam4sa/internal/aws"
	"github.com/pete911/kubectl-iam4sa/internal/k8s"
	"github.com/spf13/cobra"
	"log/slog"
	"os"
)

var (
	cmdGet = &cobra.Command{
		Use:   "get",
		Short: "get IAM service account",
		Long:  "",
		Run:   runGetCmd,
	}
)

func init() {
	RootCmd.AddCommand(cmdGet)
}

func runGetCmd(_ *cobra.Command, args []string) {
	logger := GlobalFlags.Logger()
	kubeconfig := GlobalFlags.Kubeconfig()

	k8sClient, err := k8s.NewClient(logger, kubeconfig)
	if err != nil {
		fmt.Printf("k8s client: %v\n", err)
		os.Exit(1)
	}

	logger.Debug(fmt.Sprintf("kubeconfig: %s", kubeconfig))
	awsClient, err := aws.NewClient(logger, kubeconfig.Region, kubeconfig.ClusterName)
	if err != nil {
		fmt.Printf("aws client: %v\n", err)
		os.Exit(1)
	}

	fieldSelector := GlobalFlags.FieldSelector(args)
	sas, err := k8sClient.ListIAMServiceAccounts(GlobalFlags.Namespace(), GlobalFlags.Label(), fieldSelector)
	if err != nil {
		fmt.Printf("get IAM service accounts: %v\n", err)
		os.Exit(1)
	}
	printGet(logger, awsClient, sas)
}

func printGet(logger *slog.Logger, awsClient aws.Client, sas []k8s.ServiceAccount) {
	for _, sa := range sas {
		printGetSa(logger, awsClient, sa)
	}
}

func printGetSa(logger *slog.Logger, awsClient aws.Client, sa k8s.ServiceAccount) {
	role, err := awsClient.GetIAMRole(sa.RoleName())
	if err != nil {
		logger.Error(fmt.Sprintf("get role for %s/%s service account: %v", sa.Namespace, sa.Name, err))
	}
	roleExists := sa.RoleName() == role.Name

	oidcProviderUrl, err := awsClient.GetClusterOIDCProviderUrl()
	if err != nil {
		logger.Error(fmt.Sprintf("get cluster oidc provider url: %v", err))
	}
	fmt.Printf("oidc provider url: %s\n", oidcProviderUrl)

	//events, err := awsClient.LookupEvents(sa.Namespace, sa.Name)
	//if err != nil {
	//	logger.Error(fmt.Sprintf("lookup %s/%s event: %v", sa.Namespace, sa.Name, err))
	//}

	fmt.Printf("Name:      %s\n", sa.Name)
	fmt.Printf("Namespace: %s\n", sa.Namespace)
	fmt.Println("pods:")
	for _, pod := range sa.Pods {
		fmt.Printf("  %s\n", pod)
	}
	fmt.Println("IAM Role:")
	fmt.Printf("  ARN:    %s\n", sa.IamRoleArn)
	fmt.Printf("  Exists: %t\n", roleExists)
	if roleExists {
		fmt.Println("  Assume Policy Document:")
		jsonPrettyPrint(logger, role.AssumeRolePolicyDocument)
	}
}

func jsonPrettyPrint(logger *slog.Logger, in string) {
	var out any
	if err := json.Unmarshal([]byte(in), &out); err != nil {
		logger.Error(err.Error())
		return
	}
	b, err := json.MarshalIndent(out, "", "  ")
	if err != nil {
		logger.Error(err.Error())
		return
	}
	fmt.Println(string(b))
}
