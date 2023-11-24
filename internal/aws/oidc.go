package aws

import (
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/iam"
	"time"
)

type OidcProvider struct {
	Arn         string
	ClientIDs   []string
	CreateDate  time.Time
	Thumbprints []string
	Url         string
}

func toOidcProvider(oidc *iam.GetOpenIDConnectProviderOutput, arn string) OidcProvider {
	return OidcProvider{
		Arn:         arn,
		ClientIDs:   oidc.ClientIDList,
		CreateDate:  aws.ToTime(oidc.CreateDate),
		Thumbprints: oidc.ThumbprintList,
		Url:         aws.ToString(oidc.Url),
	}
}
