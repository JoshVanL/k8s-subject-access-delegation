package interfaces

import (
	"github.com/sirupsen/logrus"
	rbacv1 "k8s.io/api/rbac/v1"
	"k8s.io/client-go/kubernetes"

	authzv1alpha1 "github.com/joshvanl/k8s-subject-access-delegation/pkg/apis/authz/v1alpha1"
)

type SubjectAccessDelegation interface {
	Namespace() string
	Name() string
	Kind() string
	Log() *logrus.Entry
	Client() kubernetes.Interface
	DestinationSubjects() []authzv1alpha1.DestinationSubject
	Triggers() []authzv1alpha1.EventTrigger
	OriginName() string
	OriginKind() string
}

type OriginSubject interface {
	ResolveOrigin() error
	RoleRefs() ([]*rbacv1.RoleRef, error)
}

type DestinationSubject interface {
	ResolveDestination() error
	Name() string
	Kind() string
}

type DestinationSubjects interface {
	ResolveDestinations() error
	Subjects() []DestinationSubject
}

type Trigger interface {
	Activate()
	Completed() (bool, error)
	WaitOn() (forcedClosed bool, err error)
	Delete() error
}
