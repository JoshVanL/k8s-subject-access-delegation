package service_account

import (
	"errors"
	"fmt"

	"github.com/sirupsen/logrus"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"

	"github.com/joshvanl/k8s-subject-access-delegation/pkg/subject_access_delegation/interfaces"
)

type ServiceAccount struct {
	log    *logrus.Entry
	client kubernetes.Interface
	sad    interfaces.SubjectAccessDelegation

	namespace      string
	name           string
	serviceAccount *corev1.ServiceAccount
}

var _ interfaces.DestinationSubject = &ServiceAccount{}

func New(sad interfaces.SubjectAccessDelegation, name string) *ServiceAccount {
	return &ServiceAccount{
		log:       sad.Log(),
		client:    sad.Client(),
		sad:       sad,
		namespace: sad.Namespace(),
		name:      name,
	}
}

func (d *ServiceAccount) getServiceAccount() error {
	options := metav1.GetOptions{}

	sa, err := d.client.Core().ServiceAccounts(d.Namespace()).Get(d.Name(), options)
	if err != nil {
		return fmt.Errorf("failed to get service account '%s': %v", d.Name(), err)
	}
	d.serviceAccount = sa

	return nil
}

func (d *ServiceAccount) ResolveDestination() error {
	if err := d.getServiceAccount(); err != nil {
		return err
	}

	if d.serviceAccount == nil {
		return errors.New("service account is nil")
	}

	return nil
}

func (d *ServiceAccount) Namespace() string {
	return d.namespace
}

func (d *ServiceAccount) Name() string {
	return d.name
}

func (d *ServiceAccount) Kind() string {
	return "ServiceAccount"
}
