package trigger

import (
	"errors"
	"fmt"

	"github.com/hashicorp/go-multierror"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (t *Trigger) BuildRoleBinding(role string) (roleBinding *rbacv1.RoleBinding, err error) {
	sa, err := t.getServiceAccount(t.sad.Spec.DestinationSubject.Name, t.Namespace())
	if err != nil {
		return nil, fmt.Errorf("failed to validated Service Account: %v", err)
	}

	Name := fmt.Sprintf("%s-role-binding", t.sad.Name)
	roleBinding = &rbacv1.RoleBinding{
		ObjectMeta: metav1.ObjectMeta{Name: Name, Namespace: sa.Namespace},
		Subjects:   []rbacv1.Subject{{Kind: "ServiceAccount", Name: sa.Name}},
		RoleRef:    rbacv1.RoleRef{Kind: "Role", Name: role},
	}

	return roleBinding, nil
}

func (t *Trigger) ValidateRoles() error {
	options := metav1.GetOptions{}

	if t.sad.Spec.OriginSubject.Kind == "Role" {
		role, err := t.client.Rbac().Roles(t.Namespace()).Get(t.sad.Spec.OriginSubject.Name, options)
		if err != nil {
			return fmt.Errorf("failed to get role %s: %v", t.sad.Spec.OriginSubject.Name, err)
		}

		t.roles = []*rbacv1.Role{role}
	} else if t.sad.Spec.OriginSubject.Kind == "ServiceAccount" {
		sa, err := t.getServiceAccount(t.sad.Spec.OriginSubject.Name, t.Namespace())
		if err != nil {
			return err
		}

		roleBindings, err := t.serviceAccountRoleBindings(sa)
		if err != nil {
			return err
		}
		t.roleBindings = roleBindings
	}

	return nil
}

func (t *Trigger) applyRoleBindings() error {
	if t.sad.Spec.OriginSubject.Kind == "Role" {
		return t.applyRoles()

	} else {
		var result error

		sa, err := t.getServiceAccount(t.sad.Spec.OriginSubject.Name, t.Namespace())
		if err != nil {
			return err
		}

		bindings, err := t.serviceAccountRoleBindings(sa)
		if err != nil {
			return err
		}

		for _, binding := range bindings {
			_, err := t.client.Rbac().RoleBindings(t.Namespace()).Create(binding)
			if err != nil {
				result = multierror.Append(result, fmt.Errorf("failed to create role binding: %v", err))

			} else {
				t.log.Infof("Role Binding \"%s\" Created", binding.Name)
			}
		}
		return nil
	}

	return nil
}

func (t *Trigger) applyRoles() error {
	t.roleBindings = []*rbacv1.RoleBinding{}
	for _, role := range t.roles {
		roleBinding, err := t.BuildRoleBinding(role.Name)
		if err != nil {
			return err
		}
		if roleBinding == nil {
			return errors.New("no role binding specified")
		}

		_, err = t.client.Rbac().RoleBindings(t.Namespace()).Create(roleBinding)
		if err != nil {
			return fmt.Errorf("failed to create role binding: %v", err)
		}

		t.log.Infof("Role Binding \"%s\" Created", roleBinding.Name)
		t.roleBindings = append(t.roleBindings, roleBinding)
	}

	return nil
}

func (t *Trigger) removeRoleBindings() error {
	var result error

	for _, binding := range t.RoleBindings() {
		if binding == nil {
			result = multierror.Append(result, errors.New("no role binding specified"))
			break
		}

		options := &metav1.DeleteOptions{}
		err := t.client.Rbac().RoleBindings(t.Namespace()).Delete(binding.Name, options)
		if err != nil {
			result = multierror.Append(result, fmt.Errorf("failed to delete role binding: %v", err))
		}

		t.log.Infof("Role Binding \"%s\" Deleted", binding.Name)
	}

	return result
}

func (t *Trigger) getRoleBindsings(namespace string) (roleBindings *rbacv1.RoleBindingList, err error) {
	options := metav1.ListOptions{}
	return t.client.Rbac().RoleBindings(t.Namespace()).List(options)
}

func (t *Trigger) serviceAccountRoleBindings(serviceAccount *corev1.ServiceAccount) (roleBindings []*rbacv1.RoleBinding, err error) {
	var bindings []*rbacv1.RoleBinding

	allRoleBindings, err := t.getRoleBindsings(t.Namespace())
	if err != nil {
		return roleBindings, fmt.Errorf("failed to get role bindings: %v", err)
	}

	for _, binding := range allRoleBindings.Items {
		for _, subject := range binding.Subjects {
			if subject.Name == serviceAccount.Name {
				bindings = append(bindings, &binding)
				break
			}
		}
	}

	return bindings, nil
}
