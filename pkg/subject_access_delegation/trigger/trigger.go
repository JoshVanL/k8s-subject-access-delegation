package trigger

import (
	"fmt"

	"github.com/hashicorp/go-multierror"

	"github.com/joshvanl/k8s-subject-access-delegation/pkg/subject_access_delegation/interfaces"
	"github.com/joshvanl/k8s-subject-access-delegation/pkg/subject_access_delegation/trigger/pod_trigger"
	"github.com/joshvanl/k8s-subject-access-delegation/pkg/subject_access_delegation/trigger/time_trigger"
)

const TimeKind = "Time"
const AddPodKind = "AddPod"
const DelPodKind = "DelPod"

func New(sad interfaces.SubjectAccessDelegation) ([]interfaces.Trigger, error) {
	var triggers []interfaces.Trigger
	var result *multierror.Error

	for _, trigger := range sad.Triggers() {
		switch trigger.Kind {
		case TimeKind:
			timeTrigger, err := time_trigger.New(sad, &trigger)
			if err != nil {
				result = multierror.Append(result, fmt.Errorf("failed to add new Time Tigger: %v", err))
				break
			}
			triggers = append(triggers, timeTrigger)

		case AddPodKind:
			addPodTrigger, err := pod_trigger.NewAddPodTrigger(sad, &trigger)
			if err != nil {
				result = multierror.Append(result, fmt.Errorf("failed to add new Add Pod Tigger: %v", err))
				break
			}
			triggers = append(triggers, addPodTrigger)

		case DelPodKind:
			delPodTrigger, err := pod_trigger.NewDelPodTrigger(sad, &trigger)
			if err != nil {
				result = multierror.Append(result, fmt.Errorf("failed to add new Del Pod Tigger: %v", err))
				break
			}
			triggers = append(triggers, delPodTrigger)

		default:
			result = multierror.Append(result, fmt.Errorf("Subject Access Delegation does not support Trigger Kind '%s'", trigger.Kind))
		}
	}

	return triggers, result.ErrorOrNil()
}
