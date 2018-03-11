package cluster_role

import (
	"errors"
	"fmt"

	"github.com/sirupsen/logrus"
	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"

	"github.com/joshvanl/k8s-subject-access-delegation/pkg/interfaces"
)

type ClusterRole struct {
	log    *logrus.Entry
	client kubernetes.Interface
	sad    interfaces.SubjectAccessDelegation

	name string
	role *rbacv1.ClusterRole
}

var _ interfaces.OriginSubject = &ClusterRole{}

const ClusterRoleKind = "ClusterRole"

func New(sad interfaces.SubjectAccessDelegation, name string) *ClusterRole {
	return &ClusterRole{
		log:    sad.Log(),
		client: sad.Client(),
		sad:    sad,
		name:   name,
	}
}

func (c *ClusterRole) RoleRefs() (roleRefs []*rbacv1.RoleRef, clusterRoleRefs []*rbacv1.RoleRef, err error) {
	return nil,
		[]*rbacv1.RoleRef{
			&rbacv1.RoleRef{
				Kind: ClusterRoleKind,
				Name: c.Name(),
			},
		}, nil
}

func (c *ClusterRole) getRole() error {
	options := metav1.GetOptions{}

	role, err := c.client.Rbac().ClusterRoles().Get(c.Name(), options)
	if err != nil {
		return fmt.Errorf("failed to get cluster role '%s': %v", c.Name(), err)
	}

	if role == nil {
		return errors.New("role is nil")
	}

	c.role = role
	return nil
}

func (c *ClusterRole) ResolveOrigin() error {
	return c.getRole()
}

func (c *ClusterRole) Name() string {
	return c.name
}

func (c *ClusterRole) Kind() string {
	return ClusterRoleKind
}
