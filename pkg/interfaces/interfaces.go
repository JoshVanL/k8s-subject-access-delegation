package interfaces

import (
	"github.com/sirupsen/logrus"
	rbacv1 "k8s.io/api/rbac/v1"
	"k8s.io/apimachinery/pkg/types"
	kubeinformers "k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"

	authzv1alpha1 "github.com/joshvanl/k8s-subject-access-delegation/pkg/apis/authz/v1alpha1"
)

type Controller interface {
	SeenUid(uid types.UID) bool
	DeletedUid(uid types.UID) bool
	AddUid(uid types.UID)
	DeleteUid(uid types.UID)
}

type SubjectAccessDelegation interface {
	Name() string
	Namespace() string
	Kind() string
	Log() *logrus.Entry
	Client() kubernetes.Interface
	KubeInformerFactory() kubeinformers.SharedInformerFactory

	ResolveDestinations() error
	OriginSubject() OriginSubject
	DestinationSubjects() []DestinationSubject
	Triggers() []Trigger

	Delegate() (closed bool, err error)
	DeleteRoleBindings() error
	Delete() error

	AddRoleBinding(addBinding Binding) error
	UpdateRoleBinding(old, new Binding) error
	DeleteRoleBinding(delBining Binding) error
	BindingSubjects() []rbacv1.Subject
	UpdateSadObject(sad *authzv1alpha1.SubjectAccessDelegation) (bool, error)

	SeenUid(uid types.UID) bool
	DeletedUid(uid types.UID) bool
	AddUid(uid types.UID)
	DeleteUid(uid types.UID)

	UpdateTriggerFired(uid int, fired bool) error
	TimeActivated() int64
	TimeFired() int64
}

type OriginSubject interface {
	ResolveOrigin() error
	RoleRefs() (roleRefs []*rbacv1.RoleRef, clusterRoleRefs []*rbacv1.RoleRef)
	Name() string
	Kind() string
	Delete()
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
	Kind() string
}

type Binding interface {
	Name() string
	Kind() string
	RoleRef() *rbacv1.RoleRef
	CreateRoleBinding() (Binding, bool, error)
	DeleteRoleBinding() error
}
