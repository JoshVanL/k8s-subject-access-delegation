package trigger

import (
	"fmt"
	"time"

	"github.com/sirupsen/logrus"
	//corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	//metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"

	authzv1alpha1 "github.com/joshvanl/k8s-subject-access-delegation/pkg/apis/authz/v1alpha1"
)

type Trigger struct {
	log          *logrus.Entry
	creationTime *time.Time

	sad       *authzv1alpha1.SubjectAccessDelegation
	namespace string
	StopCh    chan struct{}

	client       kubernetes.Interface
	roles        []*rbacv1.Role
	roleBindings []*rbacv1.RoleBinding
}

func New(log *logrus.Entry, sad *authzv1alpha1.SubjectAccessDelegation, client kubernetes.Interface, namespace string) *Trigger {
	now := time.Now()

	return &Trigger{
		log:          log,
		creationTime: &now,

		sad:       sad,
		client:    client,
		namespace: namespace,
		StopCh:    make(chan struct{}),
	}
}

func (t *Trigger) Delegate() error {
	for i := 0; i < t.Repeat(); i++ {
		t.log.Infof("Starting Subject Access Delegation \"%s\" (%d/%d)", t.sad.Name, i+1, t.Repeat())

		if err := t.ValidateRoles(); err != nil {
			return fmt.Errorf("failed to validated Role: %v", err)
		}

		close := t.TickTock()
		if close {
			return nil
		}

		if err := t.applyRoleBindings(); err != nil {
			return err
		}
	}

	return nil
}

func (t *Trigger) DeleteTrigger() error {
	close(t.StopCh)
	return t.removeRoleBindings()
}

func (t *Trigger) TickTock() (close bool) {
	delta := time.Second * time.Duration(t.sad.Spec.Duration)
	ticker := time.NewTicker(delta)

	select {
	case <-t.StopCh:
		return true
	case <-ticker.C:
		return false
	}

	//Get roles of origin subject
	// Update to origin of subject

	return false
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

func (t *Trigger) Namespace() string {
	return t.namespace
}

func (t *Trigger) RoleBindings() []*rbacv1.RoleBinding {
	return t.roleBindings
}
