package interfaces

import (
	"github.com/sirupsen/logrus"
	rbacv1 "k8s.io/api/rbac/v1"
	kubeinformers "k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"

	authzv1alpha1 "github.com/joshvanl/k8s-subject-access-delegation/pkg/apis/authz/v1alpha1"
)

type SubjectAccessDelegation interface {
	Namespace() string
	Name() string
	Kind() string
	Log() *logrus.Entry
	Client() kubernetes.Interface
	KubeInformerFactory() kubeinformers.SharedInformerFactory
	Triggers() []authzv1alpha1.EventTrigger
	Delegate() (closed bool, err error)
	DeleteRoleBindings() error
	Delete() error
	OriginSubject() OriginSubject
	DestinationSubjects() []DestinationSubject
	ResolveDestinations() error
}

type OriginSubject interface {
	ResolveOrigin() error
	RoleRefs() ([]*rbacv1.RoleRef, error)
	Name() string
	Kind() string
}

type DestinationSubject interface {
	ResolveDestination() error
	Name() string
	Kind() string
}

type Trigger interface {
	Activate()
	Completed() bool
	WaitOn() (forcedClosed bool)
	Delete() error
	Replicas() int
}
