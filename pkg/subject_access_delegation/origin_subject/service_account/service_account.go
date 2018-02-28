package service_account

import (
	"errors"
	"fmt"

	"github.com/sirupsen/logrus"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"

	"github.com/joshvanl/k8s-subject-access-delegation/pkg/interfaces"
)

const serviceAccountKind = "ServiceAccount"

type ServiceAccount struct {
	log    *logrus.Entry
	client kubernetes.Interface
	sad    interfaces.SubjectAccessDelegation

	namespace      string
	name           string
	serviceAccount *corev1.ServiceAccount
}

var _ interfaces.OriginSubject = &ServiceAccount{}

func New(sad interfaces.SubjectAccessDelegation, name string) *ServiceAccount {
	return &ServiceAccount{
		log:       sad.Log(),
		client:    sad.Client(),
		sad:       sad,
		namespace: sad.Namespace(),
		name:      name,
	}
}

func (o *ServiceAccount) RoleRefs() (roleRefs []*rbacv1.RoleRef, err error) {
	roleBindings, err := o.getRoleBindings()
	if err != nil {
		return nil, err
	}

	for _, binding := range roleBindings {
		roleRef := binding.RoleRef
		roleRefs = append(roleRefs, &roleRef)
	}

	return roleRefs, nil
}

func (o *ServiceAccount) getRoleBindings() (roleBindings []rbacv1.RoleBinding, err error) {
	// make this more efficient
	options := metav1.ListOptions{}

	bindingsList, err := o.client.Rbac().RoleBindings(o.Namespace()).List(options)
	if err != nil {
		return roleBindings, fmt.Errorf("failed to retrieve Rolebindings of Service Account '%s': %v", o.Name(), err)
	}

	if bindingsList == nil {
		return roleBindings, errors.New("bindings list is nil")
	}

	for _, binding := range bindingsList.Items {
		for _, subject := range binding.Subjects {
			if subject.Kind == serviceAccountKind && subject.Name == o.Name() {
				roleBindings = append(roleBindings, binding)
				continue
			}
		}
	}

	return roleBindings, nil
}

func (o *ServiceAccount) getServiceAccount() error {
	options := metav1.GetOptions{}

	serviceAccount, err := o.client.Core().ServiceAccounts(o.Namespace()).Get(o.Name(), options)
	if err != nil {
		return fmt.Errorf("failed to get Service Account '%s': %v", o.Name(), err)
	}

	if serviceAccount == nil {
		return errors.New("service account is nil")
	}

	o.serviceAccount = serviceAccount

	return nil
}

func (o *ServiceAccount) ResolveOrigin() error {
	return o.getServiceAccount()
}

func (o *ServiceAccount) Namespace() string {
	return o.namespace
}

func (o *ServiceAccount) Name() string {
	return o.name
}

func (o *ServiceAccount) Kind() string {
	return serviceAccountKind
}
