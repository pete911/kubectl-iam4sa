package aws

import (
	"fmt"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/iam/types"
	"net/url"
	"time"
)

type Role struct {
	ARN                      string
	Name                     string
	Description              string
	AssumeRolePolicyDocument string
	CreateDate               time.Time
	RoleLastUsed             time.Time
}

func (c Client) toRole(role *types.Role) Role {
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
