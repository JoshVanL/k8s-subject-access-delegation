package trigger

import (
	"fmt"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

//func (t *Trigger) getServiceAccountRoles(serviceAccount, namespace string) error {
//	sa, err := t.getServiceAccount(serviceAccount, namespace)
//	if err != nil {
//		return err
//	}
//
//	return nil
//}

func (t *Trigger) getServiceAccount(serviceAccountName, namespace string) (serviceAccount *corev1.ServiceAccount, err error) {
	options := metav1.GetOptions{}
	sa, err := t.client.Core().ServiceAccounts(namespace).Get(serviceAccountName, options)
	if err != nil {
		return nil, fmt.Errorf("failed to get service account %s: %v", serviceAccountName, err)
	}

	return sa, nil
}
