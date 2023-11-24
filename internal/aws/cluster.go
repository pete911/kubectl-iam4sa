package aws

import (
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/eks/types"
	"strings"
	"time"
)

type Cluster struct {
	Arn         string
	Name        string
	Certificate string
	CreatedAt   time.Time
	Endpoint    string
	OidcIssuer  string
	RoleArn     string
	Status      string
}

func (c Cluster) OidcIssuerFingerprint() (string, error) {
	return FingerprintSHA1(c.OidcIssuer, false)
}

func (c Cluster) OidcIssuerId() string {
	parts := strings.Split(c.OidcIssuer, "/")
	return parts[len(parts)-1]
}

func (c Client) toCluster(cluster *types.Cluster) Cluster {
	return Cluster{
		Arn:         aws.ToString(cluster.Arn),
		Name:        aws.ToString(cluster.Name),
		Certificate: aws.ToString(cluster.CertificateAuthority.Data),
		CreatedAt:   aws.ToTime(cluster.CreatedAt),
		Endpoint:    aws.ToString(cluster.Endpoint),
		OidcIssuer:  aws.ToString(cluster.Identity.Oidc.Issuer),
		RoleArn:     aws.ToString(cluster.RoleArn),
		Status:      string(cluster.Status),
	}
}
