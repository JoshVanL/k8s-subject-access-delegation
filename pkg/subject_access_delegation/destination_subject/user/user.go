package user

import (
	"github.com/sirupsen/logrus"
	"k8s.io/client-go/kubernetes"

	"github.com/joshvanl/k8s-subject-access-delegation/pkg/subject_access_delegation/interfaces"
)

type User struct {
	log    *logrus.Entry
	client kubernetes.Interface
	sad    interfaces.SubjectAccessDelegation

	namespace string
	name      string
	uid       string
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

func (d *User) ResolveDestination() error {
	return nil
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
