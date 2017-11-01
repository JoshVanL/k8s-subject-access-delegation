package role

import (
	"fmt"

	"github.com/sirupsen/logrus"
	//corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	//authzv1alpha1 "github.com/joshvanl/k8s-subject-access-delegation/pkg/apis/authz/v1alpha1"
	//"github.com/joshvanl/k8s-subject-access-delegation/pkg/trigger"
	"github.com/joshvanl/k8s-subject-access-delegation/pkg/subject_access_delegation/interfaces"
)

type OriginRole struct {
	log    *logrus.Entry
	client kubernetes.Interface
	sad    interfaces.SubjectAccessDelegation

	namespace string
	name      string
	role      *rbacv1.Role
}

var _ interfaces.OriginSubject = &OriginRole{}

func New(sad interfaces.SubjectAccessDelegation) *OriginRole {
	return &OriginRole{
		log:       sad.Log(),
		client:    sad.Client(),
		sad:       sad,
		namespace: sad.Namespace(),
		name:      sad.OriginName(),
	}
}

func (o *OriginRole) getRole() error {
	options := metav1.GetOptions{}

	role, err := o.client.Rbac().Roles(o.Namespace()).Get(o.Name(), options)
	if err != nil {
		return fmt.Errorf("failed to get role '%s': %v", o.Name(), err)
	}
	o.role = role

	return nil
}

func (o *OriginRole) Origin() error {
	return o.getRole()
}

//func (o *OriginRole) buildDelegation() error {
//	var roleBinding *rbacv1.RoleBinding
//
//	sa, err := t.getServiceAccount(t.sad.Spec.DestinationSubject.Name, t.Namespace())
//	if err != nil {
//		return roleBindings, fmt.Errorf("failed to validated Service Account: %v", err)
//	}
//
//	Name := fmt.Sprintf("%s-role-binding", t.sad.Name)
//	roleBinding = &rbacv1.RoleBinding{
//		ObjectMeta: metav1.ObjectMeta{Name: Name, Namespace: sa.Namespace},
//		Subjects:   []rbacv1.Subject{{Kind: "ServiceAccount", Name: sa.Name}},
//		RoleRef:    rbacv1.RoleRef{Kind: "Role", Name: role},
//	}
//
//	return []*rbacv1.RoleBinding{roleBinding}, nil
//}

func (o *OriginRole) Namespace() string {
	return o.namespace
}

func (o *OriginRole) Name() string {
	return o.name
}
