package role_binding

import (
	"fmt"

	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/joshvanl/k8s-subject-access-delegation/pkg/interfaces"
)

const ClusterRoleBindingKind = "ClusterRoleBinding"

var _ interfaces.Binding = &ClusterRoleBinding{}

type ClusterRoleBinding struct {
	clusterRoleBinding *rbacv1.ClusterRoleBinding
	sad                interfaces.SubjectAccessDelegation
}

func NewClusterRoleBinding(sad interfaces.SubjectAccessDelegation, roleRef *rbacv1.RoleRef) interfaces.Binding {
	name := fmt.Sprintf("%s-%s-%s-%s", sad.Name(), sad.OriginSubject().Name(), sad.Namespace(), roleRef.Name)

	return &ClusterRoleBinding{
		sad: sad,
		clusterRoleBinding: &rbacv1.ClusterRoleBinding{
			ObjectMeta: metav1.ObjectMeta{Name: name},
			RoleRef:    *roleRef,
			Subjects:   sad.BindingSubjects(),
		},
	}
}

func NewFromClusterRoleBinding(sad interfaces.SubjectAccessDelegation, roleBinding *rbacv1.ClusterRoleBinding) interfaces.Binding {
	name := fmt.Sprintf("%s-%s-%s-%s", sad.Name(), sad.OriginSubject().Name(), sad.Namespace(), roleBinding.RoleRef.Name)

	newBinding := new(rbacv1.ClusterRoleBinding)
	newBinding.Name = name
	newBinding.RoleRef = roleBinding.DeepCopy().RoleRef
	newBinding.Subjects = sad.BindingSubjects()

	return &ClusterRoleBinding{
		sad:                sad,
		clusterRoleBinding: newBinding,
	}
}

func (c *ClusterRoleBinding) CreateRoleBinding() (interfaces.Binding, bool, error) {
	options := metav1.GetOptions{}

	b, err := c.sad.Client().Rbac().ClusterRoleBindings().Get(c.Name(), options)
	if err != nil && b != nil && b.UID != "" {
		return nil, true, nil
	}

	binding, err := c.sad.Client().Rbac().ClusterRoleBindings().Create(c.clusterRoleBinding)
	if err != nil {
		return nil, false, fmt.Errorf("failed to create cluster role binding: %v", err)
	}

	c.clusterRoleBinding = binding

	return c, false, nil
}

func (c *ClusterRoleBinding) DeleteRoleBinding() error {
	options := &metav1.DeleteOptions{}
	if err := c.sad.Client().Rbac().ClusterRoleBindings().Delete(c.clusterRoleBinding.Name, options); err != nil {
		return fmt.Errorf("failed to delete cluster role binding: %v", err)
	}

	return nil
}

func (c *ClusterRoleBinding) Name() string {
	return c.clusterRoleBinding.Name
}

func (c *ClusterRoleBinding) RoleRef() *rbacv1.RoleRef {
	return &c.clusterRoleBinding.RoleRef
}

func (c *ClusterRoleBinding) Kind() string {
	return ClusterRoleBindingKind
}
