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
	"github.com/aws/aws-sdk-go-v2/service/iam"
	iamtypes "github.com/aws/aws-sdk-go-v2/service/iam/types"
	"github.com/pete911/kubectl-iam4sa/internal/errs"
	"k8s.io/apimachinery/pkg/util/json"
	"log/slog"
	"net/url"
	"time"
)

const eventsHours = 12

type Client struct {
	logger           *slog.Logger
	iamClient        *iam.Client
	cloudTrailClient *cloudtrail.Client
}

func NewClient(logger *slog.Logger, region string) (Client, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		return Client{}, err
	}
	cfg.Region = region

	return Client{
		logger:           logger.With("component", "aws"),
		iamClient:        iam.NewFromConfig(cfg),
		cloudTrailClient: cloudtrail.NewFromConfig(cfg),
	}, nil
}

type Role struct {
	ARN                      string
	Name                     string
	Description              string
	AssumeRolePolicyDocument string
	CreateDate               time.Time
	RoleLastUsed             time.Time
}

func (c Client) toRole(role *iamtypes.Role) Role {
	roleName := aws.ToString(role.RoleName)
	document, err := url.QueryUnescape(aws.ToString(role.AssumeRolePolicyDocument))
	if err != nil {
		c.logger.Warn(fmt.Sprintf("unescape %s assume role policy: %v", roleName, err))
	}

	return Role{
		ARN:                      aws.ToString(role.Arn),
		Name:                     roleName,
		Description:              aws.ToString(role.Description),
		AssumeRolePolicyDocument: document,
		CreateDate:               aws.ToTime(role.CreateDate),
		RoleLastUsed:             aws.ToTime(role.RoleLastUsed.LastUsedDate),
	}
}

func (c Client) GetIAMRole(roleName string) (Role, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	out, err := c.iamClient.GetRole(ctx, &iam.GetRoleInput{RoleName: aws.String(roleName)})
	if err != nil {
		var responseError *http.ResponseError
		if errors.As(err, &responseError) && responseError.HTTPStatusCode() == 404 {
			return Role{}, errs.NewErrNotFound(fmt.Sprintf("role %s not found", roleName))
		}
		return Role{}, err
	}
	return c.toRole(out.Role), nil
}

type Events []Event

func (e Events) FailedEvents() Events {
	var out Events
	for _, event := range e {
		if event.ErrorMessage != "" || event.ErrorCode != "" {
			out = append(out, event)
		}
	}
	return out
}

type Event struct {
	EventTime         time.Time         `json:"-"`
	EventId           string            `json:"-"`
	EventSource       string            `json:"-"`
	EventName         string            `json:"-"`
	UserName          string            `json:"-"`
	ErrorCode         string            `json:"errorCode"`    // set when there's error
	ErrorMessage      string            `json:"errorMessage"` // set when there's error
	UserIdentity      UserIdentity      `json:"userIdentity"`
	Region            string            `json:"awsRegion"`
	SourceIP          string            `json:"sourceIPAddress"`
	UserAgent         string            `json:"userAgent"`
	RequestParameters RequestParameters `json:"requestParameters"`
	RequestId         string            `json:"requestId"`
	EventType         string            `json:"eventType"`
}

type UserIdentity struct {
	Type             string `json:"type"`
	PrincipalId      string `json:"principalId"`
	UserName         string `json:"userName"`
	IdentityProvider string `json:"identityProvider"`
}

type RequestParameters struct {
	RoleArn         string `json:"roleArn"`
	RoleSessionName string `json:"roleSessionName"`
}

func (c Client) toEvents(events []cloudtrailtypes.Event) Events {
	var out []Event
	for _, e := range events {
		// set these fields from response, if json unmarshal fails, at least we have some info
		event := Event{
			EventTime:   aws.ToTime(e.EventTime),
			EventId:     aws.ToString(e.EventId),
			EventSource: aws.ToString(e.EventSource),
			EventName:   aws.ToString(e.EventName),
			UserName:    aws.ToString(e.Username),
		}
		if err := json.Unmarshal([]byte(aws.ToString(e.CloudTrailEvent)), &event); err != nil {
			c.logger.Warn(fmt.Sprintf("unmrshal %s event: %v", event.EventId, err))
		}
		out = append(out, event)
	}
	return out
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
			var responseError *http.ResponseError
			if errors.As(err, &responseError) && responseError.HTTPStatusCode() == 404 {
				return nil, errs.NewErrNotFound(fmt.Sprintf("events for %s user not found", username))
			}
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
