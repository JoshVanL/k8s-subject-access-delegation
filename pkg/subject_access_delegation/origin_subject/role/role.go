package role

import (
	"errors"
	"fmt"

	"github.com/sirupsen/logrus"
	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"

	"github.com/joshvanl/k8s-subject-access-delegation/pkg/interfaces"
)

type Role struct {
	log    *logrus.Entry
	client kubernetes.Interface
	sad    interfaces.SubjectAccessDelegation

	namespace string
	name      string
	role      *rbacv1.Role
}

var _ interfaces.OriginSubject = &Role{}

const RoleKind = "Role"

func New(sad interfaces.SubjectAccessDelegation, name string) *Role {
	return &Role{
		log:       sad.Log(),
		client:    sad.Client(),
		sad:       sad,
		namespace: sad.Namespace(),
		name:      name,
	}
}

func (o *Role) RoleRefs() (roleRefs []*rbacv1.RoleRef, clusterRoleRefs []*rbacv1.RoleRef, err error) {
	return []*rbacv1.RoleRef{
		&rbacv1.RoleRef{
			Kind: RoleKind,
			Name: o.Name(),
		},
	}, nil, nil
}

func (o *Role) getRole() error {
	options := metav1.GetOptions{}

	role, err := o.client.Rbac().Roles(o.Namespace()).Get(o.Name(), options)
	if err != nil {
		return fmt.Errorf("failed to get role '%s': %v", o.Name(), err)
	}

	if role == nil {
		return errors.New("role is nil")
	}

	o.role = role
	return nil
}

func (o *Role) ResolveOrigin() error {
	return o.getRole()
}

func (o *Role) Namespace() string {
	return o.namespace
}

func (o *Role) Name() string {
	return o.name
}

func (o *Role) Kind() string {
	return RoleKind
}
