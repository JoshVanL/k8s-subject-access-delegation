package trigger

import (
	"fmt"

	"github.com/hashicorp/go-multierror"

	"github.com/joshvanl/k8s-subject-access-delegation/pkg/interfaces"
	"github.com/joshvanl/k8s-subject-access-delegation/pkg/subject_access_delegation/trigger/node"
	"github.com/joshvanl/k8s-subject-access-delegation/pkg/subject_access_delegation/trigger/pod"
	"github.com/joshvanl/k8s-subject-access-delegation/pkg/subject_access_delegation/trigger/service"
	"github.com/joshvanl/k8s-subject-access-delegation/pkg/subject_access_delegation/trigger/time"
)

const TimeKind = "Time"
const AddNodeKind = "AddNode"
const DelNodeKind = "DelNode"
const AddServiceKind = "AddService"
const DelServiceKind = "DelService"
const AddPodKind = "AddPod"
const DelPodKind = "DelPod"

func New(sad interfaces.SubjectAccessDelegation) ([]interfaces.Trigger, error) {
	var triggers []interfaces.Trigger
	var result *multierror.Error

	for _, trigger := range sad.Triggers() {
		switch trigger.Kind {
		case TimeKind:
			timeTrigger, err := time.New(sad, &trigger)
			if err != nil {
				result = multierror.Append(result, fmt.Errorf("failed to add new Time Tigger: %v", err))
				break
			}
			triggers = append(triggers, timeTrigger)

		case AddNodeKind:
			addNodeTrigger, err := node.NewAddNode(sad, &trigger)
			if err != nil {
				result = multierror.Append(result, fmt.Errorf("failed to add new Add Node Tigger: %v", err))
				break
			}
			triggers = append(triggers, addNodeTrigger)

		case DelNodeKind:
			delNodeTrigger, err := node.NewDelNode(sad, &trigger)
			if err != nil {
				result = multierror.Append(result, fmt.Errorf("failed to add new Del Node Tigger: %v", err))
				break
			}
			triggers = append(triggers, delNodeTrigger)

		case AddServiceKind:
			addServiceTrigger, err := service.NewAddService(sad, &trigger)
			if err != nil {
				result = multierror.Append(result, fmt.Errorf("failed to add new Add Service Tigger: %v", err))
				break
			}
			triggers = append(triggers, addServiceTrigger)

		case DelServiceKind:
			delServiceTrigger, err := service.NewDelService(sad, &trigger)
			if err != nil {
				result = multierror.Append(result, fmt.Errorf("failed to add new Del Service Tigger: %v", err))
				break
			}
			triggers = append(triggers, delServiceTrigger)

		case AddPodKind:
			addPodTrigger, err := pod.NewAddPod(sad, &trigger)
			if err != nil {
				result = multierror.Append(result, fmt.Errorf("failed to add new Add Pod Tigger: %v", err))
				break
			}
			triggers = append(triggers, addPodTrigger)

		case DelPodKind:
			delPodTrigger, err := pod.NewDelPod(sad, &trigger)
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
