package cmd

import (
	"fmt"
	"github.com/pete911/kubectl-iam4sa/internal/aws"
	"github.com/pete911/kubectl-iam4sa/internal/k8s"
	"github.com/pete911/kubectl-iam4sa/internal/out"
	"github.com/spf13/cobra"
	"log/slog"
	"os"
)

var (
	cmdList = &cobra.Command{
		Use:   "list",
		Short: "list IAM service accounts",
		Long:  "",
		Run:   runListCmd,
	}
)

func init() {
	RootCmd.AddCommand(cmdList)
}

func runListCmd(_ *cobra.Command, args []string) {
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
		fmt.Printf("list IAM service accounts: %v\n", err)
		os.Exit(1)
	}

	if err := printListTable(logger, awsClient, sas); err != nil {
		fmt.Printf("print table: %v\n", err)
		os.Exit(1)
	}
}

func printListTable(logger *slog.Logger, awsClient aws.Client, sas []k8s.ServiceAccount) error {
	table := out.NewTable()
	if err := table.AddRow("NAMESPACE", "SERVICE ACCOUNT", "PODS", "IAM ROLE ACCOUNT", "IAM ROLE", "EVENTS", "FAILED"); err != nil {
		return err
	}
	for _, sa := range sas {
		events, err := awsClient.LookupEvents(sa.Namespace, sa.Name)
		if err != nil {
			logger.Error(fmt.Sprintf("lookup %s/%s event: %v", sa.Namespace, sa.Name, err))
		}

		numPods := fmt.Sprintf("%d", len(sa.Pods))
		numEvents := fmt.Sprintf("%d", len(events))
		numFailedEvents := fmt.Sprintf("%d", len(events.FailedEvents()))
		if err := table.AddRow(sa.Namespace, sa.Name, numPods, sa.RoleAccount(), sa.RoleName(), numEvents, numFailedEvents); err != nil {
			return err
		}
	}
	return table.Print()
}
