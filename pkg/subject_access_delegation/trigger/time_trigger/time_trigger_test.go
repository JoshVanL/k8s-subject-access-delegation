package time_trigger

import (
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/sirupsen/logrus"
)

type fakeTimeTrigger struct {
	*TimeTrigger
	ctrl *gomock.Controller
}

func newFakeTimeTrigger(t *testing.T) *fakeTimeTrigger {
	g := &fakeTimeTrigger{
		ctrl: gomock.NewController(t),
		TimeTrigger: &TimeTrigger{
			StopCh:    make(chan struct{}),
			timestamp: time.Now(),
			completed: false,
			log:       logrus.NewEntry(logrus.New()),
		},
	}

	g.TimeTrigger.log.Level = logrus.ErrorLevel

	return g
}

func TestTimeTrigger_Completed(t *testing.T) {
	g := newFakeTimeTrigger(t)
	defer g.ctrl.Finish()

	if g.Completed() {
		t.Error("expected time trigger to not be completed, it is")
	}
}

func TestTimeTrigger_Successful(t *testing.T) {
	g := newFakeTimeTrigger(t)
	defer g.ctrl.Finish()

	g.TimeTrigger.timestamp = time.Now().Add(time.Second * 2)
	g.Activate()

	if g.WaitOn() {
		t.Error("expected time trigger to not be force closed, it was")
	}

	if !g.Completed() {
		t.Error("expected time trigger to be completed, it isn't")
	}
}

func TestTimeTrigger_ForceClosed(t *testing.T) {
	g := newFakeTimeTrigger(t)
	defer g.ctrl.Finish()

	g.timestamp = time.Now().Add(time.Second * 5)
	g.Activate()

	go func(g *fakeTimeTrigger, t *testing.T) {
		time.Sleep(time.Second)
		if err := g.Delete(); err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	}(g, t)

	if !g.WaitOn() {
		t.Error("expected time trigger to be force closed, it wasn't")
	}

	if g.Completed() {
		t.Error("expected time trigger to not be completed, it is")
	}
}

func TestTimeTrigger_DoubleActivate(t *testing.T) {
	g := newFakeTimeTrigger(t)
	defer g.ctrl.Finish()

	g.timestamp = time.Now().Add(time.Second)
	g.Activate()

	if g.WaitOn() {
		t.Error("expected time trigger to not be force closed, it was")
	}

	if !g.Completed() {
		t.Error("expected time trigger to be completed, it isn't")
	}

	g.Activate()
	if g.WaitOn() {
		t.Error("expected time trigger to not be force closed, it was")
	}

	if !g.Completed() {
		t.Error("expected time trigger to be completed, it isn't")
	}

}
