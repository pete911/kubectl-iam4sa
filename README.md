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

TODO
