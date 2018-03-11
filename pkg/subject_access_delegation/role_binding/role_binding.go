package role_binding

import (
	"fmt"

	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/joshvanl/k8s-subject-access-delegation/pkg/interfaces"
)

var _ interfaces.Binding = &RoleBinding{}
var _ interfaces.Binding = &ClusterRoleBinding{}

type RoleBinding struct {
	roleBinding *rbacv1.RoleBinding
	sad         interfaces.SubjectAccessDelegation
}

type ClusterRoleBinding struct {
	clusterRoleBinding *rbacv1.ClusterRoleBinding
	sad                interfaces.SubjectAccessDelegation
}

func NewRoleBinding(sad interfaces.SubjectAccessDelegation, roleRef *rbacv1.RoleRef) interfaces.Binding {
	name := fmt.Sprintf("%s-%s-%s-%s", sad.Name(), sad.OriginSubject().Name(), sad.Namespace(), roleRef.Name)

	return &RoleBinding{
		sad: sad,
		roleBinding: &rbacv1.RoleBinding{
			ObjectMeta: metav1.ObjectMeta{Name: name, Namespace: sad.Namespace()},
			RoleRef:    *roleRef,
			Subjects:   sad.BindingSubjects(),
		},
	}
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

func NewFromRoleBinding(sad interfaces.SubjectAccessDelegation, roleBinding *rbacv1.RoleBinding) interfaces.Binding {
	return &RoleBinding{
		sad:         sad,
		roleBinding: roleBinding,
	}
}

func NewFromClusterRoleBinding(sad interfaces.SubjectAccessDelegation, roleBinding *rbacv1.ClusterRoleBinding) interfaces.Binding {
	return &ClusterRoleBinding{
		sad:                sad,
		clusterRoleBinding: roleBinding,
	}
}

func (r *RoleBinding) CreateRoleBinding() (interfaces.Binding, error) {
	binding, err := r.sad.Client().Rbac().RoleBindings(r.sad.Namespace()).Create(r.roleBinding)
	if err != nil {
		return nil, fmt.Errorf("failed to create role binding: %v", err)
	}

	r.roleBinding = binding

	return r, nil
}

func (c *ClusterRoleBinding) CreateRoleBinding() (interfaces.Binding, error) {
	binding, err := c.sad.Client().Rbac().ClusterRoleBindings().Create(c.clusterRoleBinding)
	if err != nil {
		return nil, fmt.Errorf("failed to create cluster role binding: %v", err)
	}

	c.clusterRoleBinding = binding

	return c, nil
}

func (r *RoleBinding) DeleteRoleBinding() error {
	options := &metav1.DeleteOptions{}
	if err := r.sad.Client().Rbac().RoleBindings(r.sad.Namespace()).Delete(r.roleBinding.Name, options); err != nil {
		return fmt.Errorf("failed to delete role binding: %v", err)
	}

	return nil
}

func (c *ClusterRoleBinding) DeleteRoleBinding() error {
	options := &metav1.DeleteOptions{}
	if err := c.sad.Client().Rbac().ClusterRoleBindings().Delete(c.clusterRoleBinding.Name, options); err != nil {
		return fmt.Errorf("failed to delete cluster role binding: %v", err)
	}

	return nil
}

func (r *RoleBinding) Name() string {
	return r.roleBinding.Name
}

func (c *ClusterRoleBinding) Name() string {
	return c.clusterRoleBinding.Name
}

func (r *RoleBinding) RoleRef() *rbacv1.RoleRef {
	return &r.roleBinding.RoleRef
}

func (c *ClusterRoleBinding) RoleRef() *rbacv1.RoleRef {
	return &c.clusterRoleBinding.RoleRef
}
