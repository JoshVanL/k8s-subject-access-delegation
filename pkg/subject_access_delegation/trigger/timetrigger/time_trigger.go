package timetrigger

import (
	"time"

	"github.com/sirupsen/logrus"

	"github.com/joshvanl/k8s-subject-access-delegation/pkg/subject_access_delegation/interfaces"
)

type TimeTrigger struct {
	log *logrus.Entry

	sad interfaces.SubjectAccessDelegation

	StopCh   chan struct{}
	tickerCh <-chan time.Time
	duration int64
	ready    bool
}

var _ interfaces.Trigger = &TimeTrigger{}

func New(sad interfaces.SubjectAccessDelegation) *TimeTrigger {

	return &TimeTrigger{
		log: sad.Log(),

		sad:      sad,
		StopCh:   make(chan struct{}),
		duration: sad.Duration(),
		ready:    false,
	}
}

func (t *TimeTrigger) Activate() {
	t.log.Debug("Time Trigger activated")
	t.TickTock()
}

func (t *TimeTrigger) WaitOn() error {
	t.log.Debug("Trigger waiting")

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
		t.ready = true
		return false
	case <-t.StopCh:
		return true
	}

	return false
}

func (t *TimeTrigger) Ready() (ready bool, err error) {
	return t.ready, nil
}

//func (t *Trigger) DeleteTrigger() error {
//	close(t.StopCh)
//	return t.removeRoleBindings()
//}
//

func (t *TimeTrigger) TickTock() {
	delta := time.Second * time.Duration(t.duration)
	t.tickerCh = time.After(delta)
}

//func (t *Trigger) Duration() int64 {
//	return t.sad.Spec.Duration
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
