package time

import (
	"fmt"
	"time"

	"github.com/sirupsen/logrus"

	authzv1alpha1 "github.com/joshvanl/k8s-subject-access-delegation/pkg/apis/authz/v1alpha1"
	"github.com/joshvanl/k8s-subject-access-delegation/pkg/subject_access_delegation/interfaces"
	"github.com/joshvanl/k8s-subject-access-delegation/pkg/subject_access_delegation/utils"
)

type Time struct {
	log *logrus.Entry

	sad       interfaces.SubjectAccessDelegation
	timestamp time.Time
	replicas  int

	StopCh    chan struct{}
	tickerCh  <-chan time.Time
	completed bool
}

var _ interfaces.Trigger = &Time{}

func New(sad interfaces.SubjectAccessDelegation, trigger *authzv1alpha1.EventTrigger) (timeTrigger *Time, err error) {

	timestamp, err := utils.ParseTime(trigger.Value)
	if err != nil {
		return nil, fmt.Errorf("failed to create new time trigger: %v", err)
	}

	sad.Log().Debugf("%+v", timestamp)

	return &Time{
		log:       sad.Log(),
		sad:       sad,
		replicas:  trigger.Replicas,
		StopCh:    make(chan struct{}),
		timestamp: timestamp,
		completed: false,
	}, nil
}

func (t *Time) Activate() {
	t.log.Debug("Time Trigger activated")
	t.TickTock()
}

func (t *Time) WaitOn() (forceClosed bool) {
	t.log.Debug("Trigger waiting")

	if t.watchChannels() {
		t.log.Debug("Time Trigger was force closed")
		return true
	}

	t.log.Debug("Time Trigger time expired")
	return false
}

func (t *Time) watchChannels() (forceClose bool) {
	select {
	case <-t.StopCh:
		return true
	case <-t.tickerCh:
		t.completed = true
		return false
	}
}

func (t *Time) Completed() bool {
	return t.completed
}

func (t *Time) Delete() error {
	close(t.StopCh)
	return nil
}

func (t *Time) TickTock() {
	t.tickerCh = time.After(time.Until(t.timestamp))
}

func (t *Time) Replicas() int {
	return t.replicas
}
