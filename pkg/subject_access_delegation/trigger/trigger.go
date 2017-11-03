package trigger

import (
	"github.com/joshvanl/k8s-subject-access-delegation/pkg/subject_access_delegation/interfaces"
	"github.com/joshvanl/k8s-subject-access-delegation/pkg/subject_access_delegation/trigger/timetrigger"
)

var stopChs []chan struct{}

func New(sad interfaces.SubjectAccessDelegation) ([]interfaces.Trigger, error) {
	var triggers []interfaces.Trigger
	if sad.Duration() != 0 {
		triggers = append(triggers, timetrigger.New(sad))
	}

	return triggers, nil
}

//func (t *Trigger) Delegate() error {
//	for i := 0; i < t.Repeat(); i++ {
//		t.log.Infof("Starting Subject Access Delegation \"%s\" (%d/%d)", t.sad.Name, i+1, t.Repeat())
//
//		if err := t.ValidateRoles(); err != nil {
//			return fmt."Errorf("failed to validated Role: %v", err)
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
