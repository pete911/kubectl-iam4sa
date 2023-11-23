# kubectl-iam4sa
Debug IAM roles for service accounts. User needs to have access to cluster, AWS IAM and CloudTrail API.

## list service accounts with IAM role

`kubectl-iam4sa list -A` - list service accounts in all namespaces
```shell
NAMESPACE   SERVICE ACCOUNT                      PODS  IAM ROLE ACCOUNT  IAM ROLE              EVENTS  FAILED
default     ebs-csi-controller-sa                2     123456789123      ebs-csi-controller    0       0
karpenter   karpenter                            2     123456789123      karpenter-controller  15      0
prometheus  amp-iamproxy-ingest-service-account  1     123456789123      prometheus            40      25
```
List displays service accounts with `eks.amazonaws.com/role-arn` annotations, number of pods that use this service
account. IAM Role account and name is from the service account annotation. Events is a number of events
(from CloudTrail) in the past 12 hours for this service account.

## get service account

`kubeclt-iam4sa get -n <namespace> <service-account>`
```shell
Namespace: prometheus Name: amp-iamproxy-ingest-service-account
pods:
  prometheus-server-abc-xyz
IAM Role ARN: arn:aws:iam::123456789123:role/prometheus
  Expected Federated Principal: arn:aws:iam::123456789123:oidc-provider/oidc.eks.eu-west-2.amazonaws.com/id/abcxyz123
  Expected aud: oidc.eks.eu-west-2.amazonaws.com/id/abcxyz123:aud": "sts.amazon.com"
  Expected sub: oidc.eks.eu-west-2.amazonaws.com/id/abcxyz123:sub": "system:serviceaccount:prometheus:amp-iamproxy-ingest-service-account"
  Assume Policy Document:
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

List more detailed information about service account(s). IAM Trust policy and also expected principal, aud and sub. This
makes it easier to verify if the IAM policy is configured correctly. In the example above, the pod is requesting
`prometheus-ingest` role, but the role that is set in annotation is `prometheus`. In this case
most likely the pod needs to be restarted.
