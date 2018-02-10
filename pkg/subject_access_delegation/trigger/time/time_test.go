package time

//TODO: Get rid of the long wait times

import (
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/sirupsen/logrus"
)

type fakeTime struct {
	*Time
	ctrl *gomock.Controller
}

func newFakeTime(t *testing.T) *fakeTime {
	g := &fakeTime{
		ctrl: gomock.NewController(t),
		Time: &Time{
			StopCh:    make(chan struct{}),
			timestamp: time.Now(),
			completed: false,
			log:       logrus.NewEntry(logrus.New()),
		},
	}

	g.Time.log.Level = logrus.ErrorLevel

	return g
}

func TestTime_Completed(t *testing.T) {
	g := newFakeTime(t)
	defer g.ctrl.Finish()

	if g.Completed() {
		t.Error("expected time trigger to not be completed, it is")
	}
}

func TestTime_Successful(t *testing.T) {
	g := newFakeTime(t)
	defer g.ctrl.Finish()

	g.Time.timestamp = time.Now().Add(time.Millisecond)
	g.Activate()

	if g.WaitOn() {
		t.Error("expected time trigger to not be force closed, it was")
	}

	if !g.Completed() {
		t.Error("expected time trigger to be completed, it isn't")
	}
}

func TestTime_ForceClosed(t *testing.T) {
	g := newFakeTime(t)
	defer g.ctrl.Finish()

	g.timestamp = time.Now().Add(time.Second)
	g.Activate()

	go func(g *fakeTime, t *testing.T) {
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

func TestTime_DoubleActivate(t *testing.T) {
	g := newFakeTime(t)
	defer g.ctrl.Finish()

	g.timestamp = time.Now().Add(time.Nanosecond)
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
