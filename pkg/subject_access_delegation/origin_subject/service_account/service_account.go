package service_account

import (
	"fmt"

	"github.com/sirupsen/logrus"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"

	"github.com/joshvanl/k8s-subject-access-delegation/pkg/subject_access_delegation/interfaces"
)

type OriginSA struct {
	log    *logrus.Entry
	client kubernetes.Interface
	sad    interfaces.SubjectAccessDelegation

	namespace      string
	name           string
	serviceAccount *corev1.ServiceAccount
}

var _ interfaces.OriginSubject = &OriginSA{}

func New(sad interfaces.SubjectAccessDelegation) *OriginSA {
	return &OriginSA{
		log:       sad.Log(),
		client:    sad.Client(),
		sad:       sad,
		namespace: sad.Namespace(),
		name:      sad.OriginName(),
	}
}

func (o *OriginSA) RoleRefs() (roleRefs []*rbacv1.RoleRef, err error) {
	roleBindings, err := o.getSARoleBindings()
	if err != nil {
		return nil, err
	}

	for _, binding := range roleBindings {
		roleRefs = append(roleRefs, &binding.RoleRef)
	}

	return roleRefs, nil
}

func (o *OriginSA) getSARoleBindings() (roleBindings []*rbacv1.RoleBinding, err error) {
	// make this more efficient
	options := metav1.ListOptions{}

	bindingsList, err := o.client.Rbac().RoleBindings(o.Namespace()).List(options)
	if err != nil {
		return roleBindings, fmt.Errorf("failed to retrieve Rolebindings of Service Account '%s': %v", o.Name(), err)
	}

	for _, binding := range bindingsList.Items {
		for _, subject := range binding.Subjects {
			if subject.Kind == "ServiceAccount" && subject.Name == o.Name() {
				roleBindings = append(roleBindings, &binding)
			}
		}
	}

	return roleBindings, nil
}

func (o *OriginSA) getServiceAccount() error {
	options := metav1.GetOptions{}

	serviceAccount, err := o.client.Core().ServiceAccounts(o.Namespace()).Get(o.Name(), options)
	if err != nil {
		return fmt.Errorf("failed to get Service Account '%s': %v", o.Name(), err)
	}

	o.serviceAccount = serviceAccount

	return nil
}

func (o *OriginSA) ResolveOrigin() error {
	return o.getServiceAccount()
}

func (o *OriginSA) Namespace() string {
	return o.namespace
}

func (o *OriginSA) Name() string {
	return o.name
}
