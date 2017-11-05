package interfaces

import (
	"github.com/sirupsen/logrus"
	//rbacv1 "k8s.io/api/rbac/v1"
	"k8s.io/client-go/kubernetes"

	authzv1alpha1 "github.com/joshvanl/k8s-subject-access-delegation/pkg/apis/authz/v1alpha1"
)

type SubjectAccessDelegation interface {
	Namespace() string
	Name() string
	Kind() string
	Log() *logrus.Entry
	Client() kubernetes.Interface
	OriginName() string
	OriginKind() string
	DestinationSubjects() []authzv1alpha1.DestinationSubject
	Duration() int64
}

type OriginSubject interface {
	ResolveOrigin() error
	//ApplyDelegation() error
	//BuildDelegations() ([]*rbacv1.RoleBinding, error)
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
	Ready() (bool, error)
	WaitOn() (forcedClosed bool, err error)
	Delete() error
}
