package group

import (
	"errors"
	"fmt"

	"github.com/sirupsen/logrus"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"

	"github.com/joshvanl/k8s-subject-access-delegation/pkg/subject_access_delegation/interfaces"
)

type Group struct {
	log    *logrus.Entry
	client kubernetes.Interface
	sad    interfaces.SubjectAccessDelegation

	namespace string
	name      string
	group     *corev1.ServiceAccount
}

var _ interfaces.DestinationSubject = &Group{}

func New(sad interfaces.SubjectAccessDelegation, name string) *Group {
	return &Group{
		log:       sad.Log(),
		client:    sad.Client(),
		sad:       sad,
		namespace: sad.Namespace(),
		name:      name,
	}
}

func (d *Group) getGroupAccount() error {
	options := metav1.GetOptions{}

	sa, err := d.client.Core().ServiceAccounts(d.Namespace()).Get(d.Name(), options)
	if err != nil {
		return fmt.Errorf("failed to get service account '%s': %v", d.Name(), err)
	}
	d.group = sa

	return nil
}

func (d *Group) ResolveDestination() error {
	if err := d.getGroupAccount(); err != nil {
		return err
	}

	if d.group == nil {
		return errors.New("service account is nil")
	}

	return nil
}

func (d *Group) Namespace() string {
	return d.namespace
}

func (d *Group) Name() string {
	return d.name
}

func (d *Group) Kind() string {
	return "Group"
}
