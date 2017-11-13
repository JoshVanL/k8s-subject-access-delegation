package user

import (
	"fmt"

	"github.com/sirupsen/logrus"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"

	"github.com/joshvanl/k8s-subject-access-delegation/pkg/subject_access_delegation/interfaces"
)

type User struct {
	log    *logrus.Entry
	client kubernetes.Interface
	sad    interfaces.SubjectAccessDelegation

	namespace string
	name      string
	user      *corev1.ServiceAccount
}

var _ interfaces.DestinationSubject = &User{}

func New(sad interfaces.SubjectAccessDelegation, name string) *User {
	return &User{
		log:       sad.Log(),
		client:    sad.Client(),
		sad:       sad,
		namespace: sad.Namespace(),
		name:      name,
	}
}

func (d *User) getUserAccount() error {
	options := metav1.GetOptions{}

	sa, err := d.client.Core().ServiceAccounts(d.Namespace()).Get(d.Name(), options)
	if err != nil {
		return fmt.Errorf("failed to get service account '%s': %v", d.Name(), err)
	}
	d.user = sa

	return nil
}

func (d *User) ResolveDestination() error {
	return d.getUserAccount()
}

func (d *User) Namespace() string {
	return d.namespace
}

func (d *User) Name() string {
	return d.name
}

func (d *User) Kind() string {
	return "User"
}
