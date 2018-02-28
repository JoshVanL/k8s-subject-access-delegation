package group

import (
	"github.com/sirupsen/logrus"
	"k8s.io/client-go/kubernetes"

	"github.com/joshvanl/k8s-subject-access-delegation/pkg/subject_access_delegation/interfaces"
)

type Group struct {
	log    *logrus.Entry
	client kubernetes.Interface
	sad    interfaces.SubjectAccessDelegation

	namespace string
	name      string
	groudID   string
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

func (d *Group) ResolveDestination() error {
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
