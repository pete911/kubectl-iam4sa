package aws

import (
	"context"
	"errors"
	"fmt"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/aws/transport/http"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/cloudtrail"
	cloudtrailtypes "github.com/aws/aws-sdk-go-v2/service/cloudtrail/types"
	"github.com/aws/aws-sdk-go-v2/service/eks"
	"github.com/aws/aws-sdk-go-v2/service/iam"
	"github.com/aws/aws-sdk-go-v2/service/sts"
	"github.com/pete911/kubectl-iam4sa/internal/errs"
	"log/slog"
	"time"
)

const eventsHours = 12

type Client struct {
	logger           *slog.Logger
	clusterName      string
	account          string
	region           string
	iamClient        *iam.Client
	cloudTrailClient *cloudtrail.Client
	eksClient        *eks.Client
}

func NewClient(logger *slog.Logger, region, clusterName string) (Client, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		return Client{}, err
	}

	// we should use the same region as is in the kubeconfig, if this is not the case, log warning
	if region == "" {
		logger.Warn(fmt.Sprintf("no region supplied, defaulting to %s, this can be different from cluster region", cfg.Region))
	} else {
		cfg.Region = region
	}

	out, err := sts.NewFromConfig(cfg).GetCallerIdentity(ctx, &sts.GetCallerIdentityInput{})
	if err != nil {
		return Client{}, err
	}
	account := aws.ToString(out.Account)

	return Client{
		logger:           logger,
		clusterName:      clusterName,
		account:          account,
		region:           cfg.Region,
		iamClient:        iam.NewFromConfig(cfg),
		cloudTrailClient: cloudtrail.NewFromConfig(cfg),
		eksClient:        eks.NewFromConfig(cfg),
	}, nil
}

func (c Client) GetIAMRole(roleName string) (Role, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	out, err := c.iamClient.GetRole(ctx, &iam.GetRoleInput{RoleName: aws.String(roleName)})
	if err != nil {
		err = handleResponseError(err, fmt.Sprintf("role %s", roleName))
		return Role{}, err
	}
	return c.toRole(out.Role), nil
}

func (c Client) LookupEvents(namespace, serviceAccount string) (Events, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	username := fmt.Sprintf("system:serviceaccount:%s:%s", namespace, serviceAccount)
	in := &cloudtrail.LookupEventsInput{
		LookupAttributes: []cloudtrailtypes.LookupAttribute{{
			AttributeKey:   cloudtrailtypes.LookupAttributeKeyUsername,
			AttributeValue: aws.String(username),
		}},
		StartTime: aws.Time(time.Now().Add(-(eventsHours * time.Hour))),
	}

	var events []cloudtrailtypes.Event
	for {
		out, err := c.cloudTrailClient.LookupEvents(ctx, in)
		if err != nil {
			err = handleResponseError(err, fmt.Sprintf("events for %s user", username))
			return nil, err
		}
		events = append(events, out.Events...)
		if aws.ToString(out.NextToken) == "" {
			break
		}
		in.NextToken = out.NextToken
	}
	return c.toEvents(events), nil
}

func (c Client) DescribeCluster() (Cluster, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	out, err := c.eksClient.DescribeCluster(ctx, &eks.DescribeClusterInput{Name: aws.String(c.clusterName)})
	if err != nil {
		err = handleResponseError(err, fmt.Sprintf("cluster %s", c.clusterName))
		return Cluster{}, err
	}
	return c.toCluster(out.Cluster), nil
}

func (c Client) GetClusterOidcProvider(clusterOidcIssuerId string) (OidcProvider, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	arn := fmt.Sprintf("arn:aws:iam::%s:oidc-provider/oidc.eks.%s.amazonaws.com/id/%s", c.account, c.region, clusterOidcIssuerId)
	out, err := c.iamClient.GetOpenIDConnectProvider(ctx, &iam.GetOpenIDConnectProviderInput{OpenIDConnectProviderArn: aws.String(arn)})
	if err != nil {
		err = handleResponseError(err, fmt.Sprintf("oidc provider %s", arn))
		return OidcProvider{}, err
	}
	return toOidcProvider(out), nil
}

// handleResponseError converts error to custom error (if possible) to make handling of errors easier
func handleResponseError(err error, requestName string) error {
	var responseError *http.ResponseError
	if errors.As(err, &responseError) && responseError.HTTPStatusCode() == 404 {
		return errs.NewErrNotFound(fmt.Sprintf("%s: not found", requestName))
	}
	return fmt.Errorf("%s: %w", requestName, err)
}
