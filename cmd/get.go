package cmd

import (
	"encoding/json"
	"fmt"
	"github.com/pete911/kubectl-iam4sa/internal/aws"
	"github.com/pete911/kubectl-iam4sa/internal/k8s"
	"github.com/pete911/kubectl-iam4sa/internal/out"
	"github.com/spf13/cobra"
	"log/slog"
	"os"
	"time"
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

	events, err := awsClient.LookupEvents(sa.Namespace, sa.Name)
	if err != nil {
		logger.Error(fmt.Sprintf("lookup %s/%s event: %v", sa.Namespace, sa.Name, err))
	}
	failedEvents := events.FailedEvents()

	printSA(sa)

	fmt.Println()
	printRole(logger, sa, role)

	// if there are any failed events, lets print them
	if len(failedEvents) != 0 {
		fmt.Println()
		fmt.Println("Failed Events:")
		printEvents(logger, sa, failedEvents)
	}
}

func printSA(sa k8s.ServiceAccount) {
	fmt.Printf("Name:      %s\n", sa.Name)
	fmt.Printf("Namespace: %s\n", sa.Namespace)
	fmt.Println("Pods:")
	for _, pod := range sa.Pods {
		fmt.Printf("  %s\n", pod)
	}
}

func printRole(logger *slog.Logger, sa k8s.ServiceAccount, role aws.Role) {
	fmt.Printf("Service Account Role: %s\n", sa.IamRoleArn)
	if role.ARN == "" {
		fmt.Println("AWS Role Policy Document: not found")
		return
	}
	jsonPrettyPrint(logger, role.AssumeRolePolicyDocument)
}

func printEvents(logger *slog.Logger, sa k8s.ServiceAccount, events []aws.Event) {
	table := out.NewTable(logger)
	table.AddRow("TIME", "CODE", "MESSAGE", "REQUEST ROLE", "SA ROLE")
	for i, event := range events {
		// print max last 5 failed events
		if i == 5 {
			break
		}
		table.AddRow(event.EventTime.Format(time.RFC3339), event.ErrorCode, event.ErrorMessage, event.RequestParameters.RoleArn, sa.IamRoleArn)
	}
	table.Print()
}

func jsonPrettyPrint(logger *slog.Logger, in string) {
	var inJson any
	if err := json.Unmarshal([]byte(in), &inJson); err != nil {
		logger.Error(err.Error())
		return
	}
	b, err := json.MarshalIndent(inJson, "", "  ")
	if err != nil {
		logger.Error(err.Error())
		return
	}
	fmt.Println(string(b))
}
