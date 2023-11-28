# kubectl-iam4sa
Debug IAM roles for service accounts. User needs to have access to cluster, AWS IAM and CloudTrail API.

```shell
Available Commands:
  cluster  EKS cluster oidc information
  get      get IAM service account
  help     help about any command
  list     list IAM service accounts
  version  print version

Flags:
  -A, --all-namespaces          all kubernetes namespaces
      --field-selector string   kubernetes field selector
  -h, --help                    help for this command
      --kubeconfig string       path to kubeconfig file (default "~/.kube/config")
  -l, --label string            kubernetes label
      --log-level string        log level - debug, info, warn, error (default "warn")
  -n, --namespace string        kubernetes namespace (default "default")
```

## cluster information

`kubectl-iam4sa cluster`

```
Name:        main
Status:      ACTIVE
Endpoint:    https://123456789123.gr7.eu-west-2.eks.amazonaws.com
Created:     2023-10-13T08:41:17Z
OIDC Issuer:
  Url:         https://oidc.eks.eu-west-2.amazonaws.com/id/abcxyz123
  Thumbprint:  9e9e9e9e999999999eeeee9992e9999998888877
OIDC Provider:
  Arn:         arn:aws:iam::123456789123:oidc-provider/oidc.eks.eu-west-2.amazonaws.com/id/abcxyz123
  Url:         oidc.eks.eu-west-2.amazonaws.com/id/abcxyz123
  Created:     2023-10-13T08:48:18Z
  Client Ids:
    sts.amazonaws.com
  Thumbprints:
    9e9e9e9e999999999eeeee9992e9999998888877
```

Verify that the OIDC Provider is found and has matching url and thumbprint.

## list service accounts with IAM role

`kubectl-iam4sa list -A` - list service accounts in all namespaces
```
NAMESPACE   SERVICE ACCOUNT                      PODS  IAM ROLE ACCOUNT  IAM ROLE              EVENTS  FAILED
default     ebs-csi-controller-sa                2     123456789123      ebs-csi-controller    0       0
karpenter   karpenter                            2     123456789123      karpenter-controller  15      0
prometheus  amp-iamproxy-ingest-service-account  1     123456789123      prometheus            40      25
```
List displays service accounts with `eks.amazonaws.com/role-arn` annotations, number of pods that use this service
account. IAM Role account and name is from the service account annotation. Events is a number of events
(from CloudTrail) in the past 12 hours for this service account.

## get service account

`kubectl-iam4sa get -n <namespace> <service-account>`
```
Name:      amp-iamproxy-ingest-service-account
Namespace: prometheus
Pods:
  prometheus-server-abc-xyz

Service Account Role: arn:aws:iam::123456789123:role/prometheus
{
  "Statement": [
    {
      "Action": "sts:AssumeRoleWithWebIdentity",
      "Condition": {
        "StringEquals": {
          "oidc.eks.eu-west-2.amazonaws.com/id/abcxyz123:aud": "sts.amazonaws.com",
          "oidc.eks.eu-west-2.amazonaws.com/id/abcxyz123:sub": "system:serviceaccount:prometheus:amp-iamproxy-ingest-service-account"
        }
      },
      "Effect": "Allow",
      "Principal": {
        "Federated": "arn:aws:iam::123456789123:oidc-provider/oidc.eks.eu-west-2.amazonaws.com/id/abcxyz123"
      },
      "Sid": ""
    }
  ],
  "Version": "2012-10-17"
}

Failed Events:
TIME                  CODE          MESSAGE                    REQUEST ROLE                                     ACTUAL ROLE
2023-11-23T15:35:48Z  AccessDenied  An unknown error occurred  arn:aws:iam::123456789123:role/promethus-ingest  arn:aws:iam::123456789123:role/prometheus
2023-11-23T15:19:08Z  AccessDenied  An unknown error occurred  arn:aws:iam::123456789123:role/promethus-ingest  arn:aws:iam::123456789123:role/prometheus
```

List more detailed information about service account(s) and IAM role(s). Verify principal and condition in the trust
policy with output from `kubectl-iam4sa cluster` output.

In the example above, we can see in the failed events, that the pod is requesting `prometheus-ingest` role, but the role
that is set in annotation is `prometheus`. In this case most likely the pod needs to be restarted.

## download

- [binary](https://github.com/pete911/kubectl-iam4sa/releases)

## build/install

### brew

- add tap `brew tap pete911/tap`
- install `brew install kubectl-iam4sa`

### go

[go](https://golang.org/dl/) has to be installed.
- build `go build`
- install `go install`
