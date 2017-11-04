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
	delta := time.Second * time.Duration(t.Duration())
	t.tickerCh = time.After(delta)
}

func (t *TimeTrigger) Duration() int64 {
	return t.duration
}

func (t *TimeTrigger) Repeat() int64 {
	return t.sad.Duration()
}
