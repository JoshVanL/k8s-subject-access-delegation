package group

import (
	"errors"
	"fmt"

	"github.com/sirupsen/logrus"
	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"

	"github.com/joshvanl/k8s-subject-access-delegation/pkg/interfaces"
)

type Group struct {
	log    *logrus.Entry
	client kubernetes.Interface
	sad    interfaces.SubjectAccessDelegation

	namespace string
	name      string
	uid       string
}

var _ interfaces.OriginSubject = &Group{}

func New(sad interfaces.SubjectAccessDelegation, name string) *Group {
	return &Group{
		log:       sad.Log(),
		client:    sad.Client(),
		sad:       sad,
		namespace: sad.Namespace(),
		name:      name,
	}
}

func (o *Group) RoleRefs() (roleRefs []*rbacv1.RoleRef, err error) {
	roleBindings, err := o.groupRoleBindings()
	if err != nil {
		return nil, err
	}

	for _, binding := range roleBindings {
		roleRef := binding.RoleRef
		roleRefs = append(roleRefs, &roleRef)
	}

	return roleRefs, nil
}

func (o *Group) groupRoleBindings() (roleBindings []rbacv1.RoleBinding, err error) {
	// make this more efficient
	options := metav1.ListOptions{}

	bindingsList, err := o.client.Rbac().RoleBindings(o.Namespace()).List(options)
	if err != nil {
		return roleBindings, fmt.Errorf("failed to retrieve Rolebindings of Group Accounts '%s': %v", o.Name(), err)
	}

	if bindingsList == nil {
		return roleBindings, errors.New("binding list is nil")
	}

	for _, binding := range bindingsList.Items {
		for _, subject := range binding.Subjects {
			if subject.Kind == rbacv1.GroupKind && subject.Name == o.Name() {
				roleBindings = append(roleBindings, binding)
				break
			}
		}
	}

	return roleBindings, nil
}

func (o *Group) ResolveOrigin() error {
	return nil
}

func (o *Group) Namespace() string {
	return o.namespace
}

func (o *Group) Name() string {
	return o.name
}

func (o *Group) Kind() string {
	return rbacv1.GroupKind
}
