package service_account

import (
	"fmt"

	"github.com/sirupsen/logrus"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"

	"github.com/joshvanl/k8s-subject-access-delegation/pkg/subject_access_delegation/interfaces"
)

type DestinationSA struct {
	log    *logrus.Entry
	client kubernetes.Interface
	sad    interfaces.SubjectAccessDelegation

	namespace      string
	name           string
	serviceAccount *corev1.ServiceAccount
}

var _ interfaces.DestinationSubject = &DestinationSA{}

func New(sad interfaces.SubjectAccessDelegation) *DestinationSA {
	return &DestinationSA{
		log:       sad.Log(),
		client:    sad.Client(),
		sad:       sad,
		namespace: sad.Namespace(),
		name:      sad.DestinationName(),
	}
}

func (d *DestinationSA) getServiceAccount() error {
	options := metav1.GetOptions{}

	sa, err := d.client.Core().ServiceAccounts(d.Namespace()).Get(d.Name(), options)
	if err != nil {
		return fmt.Errorf("failed to get service account %s: %v", d.Name(), err)
	}
	d.serviceAccount = sa

	return nil
}

func (d *DestinationSA) Destination() error {
	return d.getServiceAccount()
}

func (d *DestinationSA) Namespace() string {
	return d.namespace
}

func (d *DestinationSA) Name() string {
	return d.name
}
