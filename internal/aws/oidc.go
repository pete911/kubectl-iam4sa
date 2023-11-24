package aws

import (
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/iam"
	"time"
)

type OidcProvider struct {
	ClientIDs   []string
	CreateDate  time.Time
	Thumbprints []string
	Url         string
}

func toOidcProvider(oidc *iam.GetOpenIDConnectProviderOutput) OidcProvider {
	return OidcProvider{
		ClientIDs:   oidc.ClientIDList,
		CreateDate:  aws.ToTime(oidc.CreateDate),
		Thumbprints: oidc.ThumbprintList,
		Url:         aws.ToString(oidc.Url),
	}
}
