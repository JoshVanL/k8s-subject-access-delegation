package trigger

import (
	"github.com/joshvanl/k8s-subject-access-delegation/pkg/subject_access_delegation/interfaces"
	"github.com/joshvanl/k8s-subject-access-delegation/pkg/subject_access_delegation/trigger/timetrigger"
)

func New(sad interfaces.SubjectAccessDelegation) ([]interfaces.Trigger, error) {
	var triggers []interfaces.Trigger
	if sad.Duration() != 0 {
		triggers = append(triggers, timetrigger.New(sad))
	}

	return triggers, nil
}
