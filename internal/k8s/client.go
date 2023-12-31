package k8s

import (
	"context"
	"fmt"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	corev1 "k8s.io/client-go/kubernetes/typed/core/v1"
	"log/slog"
	"strings"
	"time"
)

const iamRoleARNAnnotation = "eks.amazonaws.com/role-arn"

type ServiceAccount struct {
	Name       string
	Namespace  string
	IamRoleArn string
	Pods       []string
}

func (s ServiceAccount) RoleAccount() string {
	parts := strings.Split(s.IamRoleArn, ":")
	return parts[len(parts)-2]
}

func (s ServiceAccount) RoleName() string {
	parts := strings.Split(s.IamRoleArn, "/")
	return parts[len(parts)-1]
}

type Client struct {
	logger *slog.Logger
	config Kubeconfig
	coreV1 corev1.CoreV1Interface
}

func NewClient(logger *slog.Logger, config Kubeconfig) (Client, error) {
	cs, err := kubernetes.NewForConfig(config.RestConfig)
	if err != nil {
		return Client{}, err
	}
	return Client{
		logger: logger,
		config: config,
		coreV1: cs.CoreV1(),
	}, nil
}

func (c Client) ListIAMServiceAccounts(namespace, labelSelector, fieldSelector string) ([]ServiceAccount, error) {
	if namespace == "" {
		return c.listAllIAMServiceAccounts(labelSelector, fieldSelector)
	}
	return c.listIAMServiceAccounts(namespace, labelSelector, fieldSelector)
}

func (c Client) listAllIAMServiceAccounts(labelSelector, fieldSelector string) ([]ServiceAccount, error) {
	namespaces, err := c.getNamespaces()
	if err != nil {
		return nil, fmt.Errorf("get namespaces: %w", err)
	}

	var allServiceAccounts []ServiceAccount
	for _, namespace := range namespaces {
		serviceAccounts, err := c.listIAMServiceAccounts(namespace, labelSelector, fieldSelector)
		if err != nil {
			return nil, err
		}
		allServiceAccounts = append(allServiceAccounts, serviceAccounts...)
	}
	return allServiceAccounts, nil
}

func (c Client) listIAMServiceAccounts(namespace, labelSelector, fieldSelector string) ([]ServiceAccount, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	serviceAccountList, err := c.coreV1.ServiceAccounts(namespace).List(ctx, metav1.ListOptions{LabelSelector: labelSelector, FieldSelector: fieldSelector})
	if err != nil {
		return nil, err
	}

	var serviceAccounts []ServiceAccount
	for _, serviceAccount := range serviceAccountList.Items {
		if roleARN, ok := serviceAccount.Annotations[iamRoleARNAnnotation]; ok {
			pods, err := c.listPods(namespace, serviceAccount.Name)
			if err != nil {
				return nil, fmt.Errorf("list pods for %s/%s service account: %v", namespace, serviceAccount.Name, err)
			}
			serviceAccounts = append(serviceAccounts, ServiceAccount{
				Name:       serviceAccount.Name,
				Namespace:  serviceAccount.Namespace,
				IamRoleArn: roleARN,
				Pods:       pods,
			})
		}
	}
	return serviceAccounts, nil
}

func (c Client) listPods(namespace, serviceAccountName string) ([]string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	podList, err := c.coreV1.Pods(namespace).List(ctx, metav1.ListOptions{FieldSelector: fmt.Sprintf("spec.serviceAccountName=%s", serviceAccountName)})
	if err != nil {
		return nil, err
	}
	var out []string
	for _, pod := range podList.Items {
		out = append(out, pod.Name)
	}
	return out, nil
}

func (c Client) getNamespaces() ([]string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	namespaceList, err := c.coreV1.Namespaces().List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}
	var out []string
	for _, ns := range namespaceList.Items {
		out = append(out, ns.Name)
	}
	return out, nil
}
