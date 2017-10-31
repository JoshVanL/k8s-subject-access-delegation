package trigger

import (
	"errors"
	"fmt"
	"time"

	"github.com/sirupsen/logrus"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	//rbacapi "k8s.io/kubernetes/pkg/apis/rbac"

	authzv1alpha1 "github.com/joshvanl/k8s-subject-access-delegation/pkg/apis/authz/v1alpha1"
)

type Trigger struct {
	log          *logrus.Entry
	creationTime *time.Time

	sad    *authzv1alpha1.SubjectAccessDelegation
	client kubernetes.Interface

	role           *rbacv1.Role
	serviceAccount *corev1.ServiceAccount
	roleBinding    *rbacv1.RoleBinding
}

func New(log *logrus.Entry, sad *authzv1alpha1.SubjectAccessDelegation, client kubernetes.Interface) *Trigger {
	now := time.Now()

	return &Trigger{
		log:          log,
		creationTime: &now,

		sad:    sad,
		client: client,
	}
}

func (t *Trigger) ValidateRole() error {
	options := metav1.GetOptions{}

	role, err := t.client.Rbac().Roles(t.sad.Spec.Namespace).Get(t.sad.Spec.OriginSubject.Name, options)
	if err != nil {
		return fmt.Errorf("failed to get role %s: %v", t.sad.Spec.OriginSubject.Name, err)
	}

	t.role = role

	return nil
}

func (t *Trigger) ValidateServiceAccount() error {
	options := metav1.GetOptions{}

	sa, err := t.client.Core().ServiceAccounts(t.sad.Spec.Namespace).Get(t.sad.Spec.DestinationSubject.Name, options)
	if err != nil {
		return fmt.Errorf("failed to get service account %s: %v", t.sad.Spec.DestinationSubject.Name, err)
	}

	t.serviceAccount = sa

	return nil
}

func (t *Trigger) CreateRoleBinding() {
	t.roleBinding = &rbacv1.RoleBinding{
		ObjectMeta: metav1.ObjectMeta{Name: "a-role-binding", Namespace: t.serviceAccount.Namespace},
		Subjects:   []rbacv1.Subject{{Kind: "ServiceAccount", Name: t.serviceAccount.Name}},
		RoleRef:    rbacv1.RoleRef{Kind: "Role", Name: t.role.Name},
	}

	//roleBinding.ObjectMeta.Name = "a-role-binding"
	//roleBinding.ObjectMeta.Namespace = t.serviceAccount.Namespace
}

func (t *Trigger) Delegate() error {
	for i := 0; i < t.sad.Spec.Repeat; i++ {
		t.log.Infof("Starting Subject Access Delegation \"TODO: ADD NAME\" (%d/%d)", i+1, t.sad.Spec.Repeat)

		if err := t.ValidateRole(); err != nil {
			return fmt.Errorf("failed to validated Role: %v", err)
		}

		t.TickTock()

		if err := t.ValidateServiceAccount(); err != nil {
			return fmt.Errorf("failed to validated Service Account: %v", err)
		}

		t.CreateRoleBinding()

		if err := t.ApplyRoleBinding(); err != nil {
			return err
		}
	}

	return nil
}

func (t *Trigger) ApplyRoleBinding() error {
	if t.roleBinding == nil {
		return errors.New("no role binding specified")
	}
	_, err := t.client.Rbac().RoleBindings(t.serviceAccount.Namespace).Create(t.roleBinding)
	if err != nil {
		return fmt.Errorf("failed to create role binding: %v", err)
	}

	t.log.Infof("Role Binding \"%s\" Created", t.roleBinding.Name)

	return nil
}

func (t *Trigger) TickTock() {
	delta := time.Second * time.Duration(t.sad.Spec.Duration)
	ticker := time.NewTicker(delta)
	<-ticker.C

	//Get roles of origin subject
	// Update to origin of subject
}

func (t *Trigger) Duration() int64 {
	return t.sad.Spec.Duration
}

func (t *Trigger) CreationTime() *time.Time {
	return t.creationTime
}

func (t *Trigger) Repeat() int {
	return t.sad.Spec.Repeat
}
