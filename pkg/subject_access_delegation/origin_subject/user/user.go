package user

import (
	"fmt"

	"github.com/sirupsen/logrus"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"

	"github.com/joshvanl/k8s-subject-access-delegation/pkg/subject_access_delegation/interfaces"
)

type OriginUser struct {
	log    *logrus.Entry
	client kubernetes.Interface
	sad    interfaces.SubjectAccessDelegation

	namespace string
	name      string
	user      *corev1.ServiceAccount
}

var _ interfaces.OriginSubject = &OriginUser{}

func New(sad interfaces.SubjectAccessDelegation) *OriginUser {
	return &OriginUser{
		log:       sad.Log(),
		client:    sad.Client(),
		sad:       sad,
		namespace: sad.Namespace(),
		name:      sad.OriginName(),
	}
}

func (o *OriginUser) RoleRefs() (roleRefs []*rbacv1.RoleRef, err error) {
	roleBindings, err := o.getSARoleBindings()
	if err != nil {
		return nil, err
	}

	for _, binding := range roleBindings {
		roleRefs = append(roleRefs, &binding.RoleRef)
	}

	return roleRefs, nil
}

func (o *OriginUser) getSARoleBindings() (roleBindings []*rbacv1.RoleBinding, err error) {
	// make this more efficient
	options := metav1.ListOptions{}

	bindingsList, err := o.client.Rbac().RoleBindings(o.Namespace()).List(options)
	if err != nil {
		return roleBindings, fmt.Errorf("failed to retrieve Rolebindings of User Account '%s': %v", o.Name(), err)
	}

	for _, binding := range bindingsList.Items {
		for _, subject := range binding.Subjects {
			if subject.Kind == rbacv1.UserKind && subject.Name == o.Name() {
				roleBindings = append(roleBindings, &binding)
			}
		}
	}

	return roleBindings, nil
}

func (o *OriginUser) getUserAccount() error {
	options := metav1.GetOptions{}

	serviceAccount, err := o.client.Core().ServiceAccounts(o.Namespace()).Get(o.Name(), options)
	if err != nil {
		return fmt.Errorf("failed to get User Account '%s': %v", o.Name(), err)
	}

	o.user = serviceAccount

	return nil
}

func (o *OriginUser) ResolveOrigin() error {
	return o.getUserAccount()
}

func (o *OriginUser) Namespace() string {
	return o.namespace
}

func (o *OriginUser) Name() string {
	return o.name
}
