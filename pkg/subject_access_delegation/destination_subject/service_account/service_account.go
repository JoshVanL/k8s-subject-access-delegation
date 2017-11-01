package service_account

import (
	"github.com/sirupsen/logrus"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes"

	"github.com/joshvanl/k8s-subject-access-delegation/pkg/subject_access_delegation/interfaces"
)

type DestinationSA struct {
	log *logrus.Entry

	namespace string

	sad interfaces.SubjectAccessDelegation

	client         kubernetes.Interface
	serviceAccount *corev1.ServiceAccount
}

var _ interfaces.DestinationSubject = &DestinationSA{}

func New(sad interfaces.SubjectAccessDelegation) *DestinationSA {
	return &DestinationSA{
		log:       sad.Log(),
		sad:       sad,
		client:    sad.Client(),
		namespace: sad.Namespace(),
	}
}

//func (d *DestinationSA) getServiceAccount(serviceAccountName, namespace string) (serviceAccount *corev1.ServiceAccount, err error) {
//	options := metav1.GetOptions{}
//	sa, err := t.client.Core().ServiceAccounts(namespace).Get(serviceAccountName, options)
//	if err != nil {
//		return nil, fmt.Errorf("failed to get service account %s: %v", serviceAccountName, err)
//	}
//
//	return sa, nil
//}
//
//func (d *DestinationSA) Name() string {
//	return d.serviceAccount.Name
//}
