package interfaces

import (
	"github.com/sirupsen/logrus"
	//rbacv1 "k8s.io/api/rbac/v1"
	"k8s.io/client-go/kubernetes"
)

type SubjectAccessDelegation interface {
	Namespace() string
	Name() string
	Kind() string
	Log() *logrus.Entry
	Client() kubernetes.Interface
	OriginName() string
	OriginKind() string
	DestinationName() string
	DestinationKind() string
	Duration() int64
}

type OriginSubject interface {
	Origin() error
	//ApplyDelegation() error
	//BuildDelegations() ([]*rbacv1.RoleBinding, error)
}

type DestinationSubject interface {
	Destination() error
}

type Trigger interface {
	Activate()
	Ready() (bool, error)
	WaitOn() error
}
