package timetrigger

import (
	"errors"
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

	sad      interfaces.SubjectAccessDelegation
	StopCh   chan struct{}
	tickerCh <-chan time.Time
	duration int64
}

var _ interfaces.Trigger = &TimeTrigger{}

func New(sad interfaces.SubjectAccessDelegation) *TimeTrigger {
	now := time.Now()

	return &TimeTrigger{
		log:          sad.Log(),
		creationTime: &now,

		sad:      sad,
		StopCh:   make(chan struct{}),
		duration: sad.Duration(),
	}
}

func (t *TimeTrigger) Activate() {
	t.TickTock()
}

func (t *TimeTrigger) WaitOn() error {
	forceClose := t.watchChannels()
	if forceClose {
		t.log.Debug("time trigger force closed")
	}
	t.log.Debug("time trigger time expired")

	return nil
}

func (t *TimeTrigger) watchChannels() (forceClose bool) {
	select {
	case <-t.tickerCh:
		return false
	case <-t.StopCh:
		return true
	}

	return false
}

func (t *TimeTrigger) Ready() (ready bool, err error) {
	select {
	case _, ok := <-t.tickerCh:
		if ok {
			return true, nil

		} else {
			return false, errors.New("channel was unexpectedly closed")
		}

	default:
		return false, nil
	}

	return false, nil
}

//func (t *Trigger) DeleteTrigger() error {
//	close(t.StopCh)
//	return t.removeRoleBindings()
//}
//

func (t *TimeTrigger) TickTock() {
	delta := time.Second * time.Duration(t.duration)
	t.tickerCh = time.NewTicker(delta).C
}

//func (t *Trigger) Duration() int64 {
//	return t.sad.Spec.Duration
//}
//
//func (t *Trigger) CreationTime() *time.Time {
//	return t.creationTime
//}
//

func (t *TimeTrigger) Repeat() int64 {
	return t.sad.Duration()
}

//func (t *Trigger) Namespace() string {
//	return t.namespace
//}
//
//func (t *Trigger) RoleBindings() []*rbacv1.RoleBinding {
//	return t.roleBindings
//}
