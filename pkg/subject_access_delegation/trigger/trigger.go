package trigger

import (
	"github.com/joshvanl/k8s-subject-access-delegation/pkg/subject_access_delegation/interfaces"
	"github.com/joshvanl/k8s-subject-access-delegation/pkg/subject_access_delegation/trigger/timetrigger"
)

func New(sad interfaces.SubjectAccessDelegation) interfaces.Trigger {
	var trigger interfaces.Trigger
	trigger = timetrigger.New(sad)
	return trigger
}

//func (t *Trigger) Delegate() error {
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
