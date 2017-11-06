package trigger

import (
	"fmt"

	"github.com/hashicorp/go-multierror"

	"github.com/joshvanl/k8s-subject-access-delegation/pkg/subject_access_delegation/interfaces"
	"github.com/joshvanl/k8s-subject-access-delegation/pkg/subject_access_delegation/trigger/timetrigger"
)

func New(sad interfaces.SubjectAccessDelegation) ([]interfaces.Trigger, error) {
	var triggers []interfaces.Trigger
	var result *multierror.Error

	for _, trigger := range sad.Triggers() {
		switch trigger.Kind {
		case "Time":
			timeTrigger, err := timetrigger.New(sad, &trigger)
			if err != nil {
				result = multierror.Append(result, fmt.Errorf("failed to add new time tigger: %v", err))
			} else {
				triggers = append(triggers, timeTrigger)
			}

		default:
			result = multierror.Append(result, fmt.Errorf("Subject Access Delegation does not support Trigger Kind '%s'", trigger.Kind))
		}
	}

	return triggers, result.ErrorOrNil()
}
