package cmd

import (
	"errors"
	"fmt"
	"github.com/pete911/kubectl-iam4sa/internal/aws"
	"github.com/pete911/kubectl-iam4sa/internal/errs"
	"github.com/spf13/cobra"
	"log/slog"
	"os"
	"time"
)

var (
	cmdCluster = &cobra.Command{
		Use:   "cluster",
		Short: "EKS cluster oidc information",
		Long:  "",
		Run:   runClusterCmd,
	}
)

func init() {
	RootCmd.AddCommand(cmdCluster)
}

func runClusterCmd(_ *cobra.Command, args []string) {
	logger := GlobalFlags.Logger()
	kubeconfig := GlobalFlags.Kubeconfig()

	logger.Debug(fmt.Sprintf("kubeconfig: %s", kubeconfig))
	awsClient, err := aws.NewClient(logger, kubeconfig.Region, kubeconfig.ClusterName)
	if err != nil {
		fmt.Printf("aws client: %v\n", err)
		os.Exit(1)
	}

	cluster, err := awsClient.DescribeCluster()
	if err != nil {
		fmt.Printf("describe cluster: %v\n", err)
		os.Exit(1)
	}

	oidcProvider, err := awsClient.GetClusterOidcProvider(cluster.OidcIssuerId())
	if err != nil {
		// continue if the error is not found, we want to display to the user that there's no oidc provider
		var errNotFound *errs.ErrNotFound
		if !errors.As(err, &errNotFound) {
			fmt.Printf("get cluster oidc provider: %v\n", err)
			os.Exit(1)
		}
	}
	printCluster(logger, cluster, oidcProvider)
}

func printCluster(logger *slog.Logger, cluster aws.Cluster, oidcProvider aws.OidcProvider) {
	fingerprint, err := cluster.OidcIssuerFingerprint()
	if err != nil {
		logger.Error(fmt.Sprintf("oidc cluster issuer fingerprint: %v", err))
	}

	fmt.Printf("Name:        %s\n", cluster.Name)
	fmt.Printf("Status:      %s\n", cluster.Status)
	fmt.Printf("Endpoint:    %s\n", cluster.Endpoint)
	fmt.Printf("Created:     %s\n", cluster.CreatedAt.Format(time.RFC3339))
	fmt.Println("OIDC Issuer:")
	fmt.Printf("  Url:         %s\n", cluster.OidcIssuer)
	fmt.Printf("  Thumbprint:  %s\n", fingerprint)
	if oidcProvider.Url == "" {
		fmt.Println("OIDC Provider: not found")
		return
	}
	fmt.Println("OIDC Provider:")
	fmt.Printf("  Url:         %s\n", oidcProvider.Url)
	fmt.Printf("  Created:     %s\n", oidcProvider.CreateDate.Format(time.RFC3339))
	fmt.Println("  Client Ids:")
	for _, id := range oidcProvider.ClientIDs {
		fmt.Printf("    %s\n", id)
	}
	fmt.Println("  Thumbprints:")
	for _, thumbprint := range oidcProvider.Thumbprints {
		fmt.Printf("    %s\n", thumbprint)
	}
}
