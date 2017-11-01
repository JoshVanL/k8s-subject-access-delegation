package timetrigger

import (
	//"fmt"
	"time"

	"github.com/sirupsen/logrus"
	//corev1 "k8s.io/api/core/v1"
	//rbacv1 "k8s.io/api/rbac/v1"
	//metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	//"k8s.io/client-go/kubernetes"

	//authzv1alpha1 "github.com/joshvanl/k8s-subject-access-delegation/pkg/apis/authz/v1alpha1"
	"github.com/joshvanl/k8s-subject-access-delegation/pkg/subject_access_delegation/interfaces"
)

type TimeTrigger struct {
	log          *logrus.Entry
	creationTime *time.Time

	sad    interfaces.SubjectAccessDelegation
	StopCh chan struct{}
}

var _ interfaces.Trigger = &TimeTrigger{}

func New(sad interfaces.SubjectAccessDelegation) *TimeTrigger {
	now := time.Now()

	return &TimeTrigger{
		log:          sad.Log(),
		creationTime: &now,

		sad: sad,
		//client: sad.Client(),
		StopCh: make(chan struct{}),
	}
}

//func (t *TimeTrigger) Delegate() error {
//	for i := 0; i < t.Repeat(); i++ {
//		t.log.Infof("Starting Subject Access Delegation \"%s\" (%d/%d)", t.sad.Name, i+1, t.Repeat())
//
//		if err := t.ValidateRoles(); err != nil {
//			return fmt.Errorf("failed to validated Role: %v", err)
//		}
//
//		close := t.TickTock()
//		if close {
//			return nil
//		}
//
//		if err := t.applyRoleBindings(); err != nil {
//			return err
//		}
//	}
//
//	return nil
//}

//func (t *Trigger) DeleteTrigger() error {
//	close(t.StopCh)
//	return t.removeRoleBindings()
//}
//
//func (t *Trigger) TickTock() (close bool) {
//	delta := time.Second * time.Duration(t.sad.Spec.Duration)
//	ticker := time.NewTicker(delta)
//
//	select {
//	case <-t.StopCh:
//		return true
//	case <-ticker.C:
//		return false
//	}
//
//	//Get roles of origin subject
//	// Update to origin of subject
//
//	return false
//}
//
//func (t *Trigger) Duration() int64 {
//	return t.sad.Spec.Duration
//}
//
//func (t *Trigger) CreationTime() *time.Time {
//	return t.creationTime
//}
//
//func (t *Trigger) Repeat() int {
//	return t.sad.Spec.Repeat
//}

//func (t *Trigger) Namespace() string {
//	return t.namespace
//}
//
//func (t *Trigger) RoleBindings() []*rbacv1.RoleBinding {
//	return t.roleBindings
//}
