package time_trigger

import (
	"fmt"
	"time"

	"github.com/sirupsen/logrus"

	authzv1alpha1 "github.com/joshvanl/k8s-subject-access-delegation/pkg/apis/authz/v1alpha1"
	"github.com/joshvanl/k8s-subject-access-delegation/pkg/subject_access_delegation/interfaces"
	"github.com/joshvanl/k8s-subject-access-delegation/pkg/subject_access_delegation/utils"
)

type TimeTrigger struct {
	log *logrus.Entry

	sad       interfaces.SubjectAccessDelegation
	timestamp time.Time

	StopCh   chan struct{}
	tickerCh <-chan time.Time
	ready    bool
}

var _ interfaces.Trigger = &TimeTrigger{}

func New(sad interfaces.SubjectAccessDelegation, trigger *authzv1alpha1.EventTrigger) (timeTrigger *TimeTrigger, err error) {

	timestamp, err := utils.ParseTime(trigger.Value)
	if err != nil {
		return nil, fmt.Errorf("failed to create new time trigger: %v", err)
	}

	sad.Log().Debugf("%+v", timestamp)

	return &TimeTrigger{
		log: sad.Log(),

		sad:       sad,
		StopCh:    make(chan struct{}),
		timestamp: timestamp,
		ready:     false,
	}, nil
}

func (t *TimeTrigger) Activate() {
	t.log.Debug("Time Trigger activated")
	t.TickTock()
}

func (t *TimeTrigger) WaitOn() (forceClosed bool, err error) {
	t.log.Debug("Trigger waiting")

	if t.watchChannels() {
		t.log.Debug("Time Trigger was force closed")
		return true, nil
	}

	t.log.Debug("Time Trigger time expired")
	return false, nil
}

func (t *TimeTrigger) watchChannels() (forceClose bool) {
	select {
	case <-t.tickerCh:
		t.ready = true
		return false
	case <-t.StopCh:
		return true
	}
}

func (t *TimeTrigger) Ready() (ready bool, err error) {
	return t.ready, nil
}

func (t *TimeTrigger) Delete() error {
	close(t.StopCh)
	return nil
}

func (t *TimeTrigger) TickTock() {
	t.tickerCh = time.After(time.Until(t.timestamp))
}
