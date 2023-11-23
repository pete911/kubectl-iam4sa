package aws

import (
	"encoding/json"
	"fmt"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/cloudtrail/types"
	"time"
)

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

func (c Client) toEvents(events []types.Event) Events {
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
