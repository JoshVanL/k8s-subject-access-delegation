package trigger

import (
	"fmt"

	"github.com/hashicorp/go-multierror"

	authzv1alpha1 "github.com/joshvanl/k8s-subject-access-delegation/pkg/apis/authz/v1alpha1"
	"github.com/joshvanl/k8s-subject-access-delegation/pkg/interfaces"
	"github.com/joshvanl/k8s-subject-access-delegation/pkg/subject_access_delegation/trigger/node"
	"github.com/joshvanl/k8s-subject-access-delegation/pkg/subject_access_delegation/trigger/pod"
	"github.com/joshvanl/k8s-subject-access-delegation/pkg/subject_access_delegation/trigger/secret"
	"github.com/joshvanl/k8s-subject-access-delegation/pkg/subject_access_delegation/trigger/service"
	"github.com/joshvanl/k8s-subject-access-delegation/pkg/subject_access_delegation/trigger/time"
)

func New(sad interfaces.SubjectAccessDelegation, sadTriggers []authzv1alpha1.EventTrigger) ([]interfaces.Trigger, error) {
	var triggers []interfaces.Trigger
	var result *multierror.Error

	for _, trigger := range sadTriggers {
		switch trigger.Kind {
		case time.TimeKind:
			timeTrigger, err := time.New(sad, &trigger)
			if err != nil {
				result = multierror.Append(result, fmt.Errorf("failed to add new Time Tigger: %v", err))
				break
			}
			triggers = append(triggers, timeTrigger)

		case node.AddNodeKind:
			addNodeTrigger, err := node.NewAddNode(sad, &trigger)
			if err != nil {
				result = multierror.Append(result, fmt.Errorf("failed to add new Add Node Tigger: %v", err))
				break
			}
			triggers = append(triggers, addNodeTrigger)

		case node.DelNodeKind:
			delNodeTrigger, err := node.NewDelNode(sad, &trigger)
			if err != nil {
				result = multierror.Append(result, fmt.Errorf("failed to add new Del Node Tigger: %v", err))
				break
			}
			triggers = append(triggers, delNodeTrigger)

		case node.UpdateNodeKind:
			updateNodeTrigger, err := node.NewUpdateNode(sad, &trigger)
			if err != nil {
				result = multierror.Append(result, fmt.Errorf("failed to add new Update Node Tigger: %v", err))
				break
			}
			triggers = append(triggers, updateNodeTrigger)

		case service.AddServiceKind:
			addServiceTrigger, err := service.NewAddService(sad, &trigger)
			if err != nil {
				result = multierror.Append(result, fmt.Errorf("failed to add new Add Service Tigger: %v", err))
				break
			}
			triggers = append(triggers, addServiceTrigger)

		case service.DelServiceKind:
			delServiceTrigger, err := service.NewDelService(sad, &trigger)
			if err != nil {
				result = multierror.Append(result, fmt.Errorf("failed to add new Del Service Tigger: %v", err))
				break
			}
			triggers = append(triggers, delServiceTrigger)

		case service.UpdateServiceKind:
			updateServiceTrigger, err := service.NewUpdateService(sad, &trigger)
			if err != nil {
				result = multierror.Append(result, fmt.Errorf("failed to add new Update Service Tigger: %v", err))
				break
			}
			triggers = append(triggers, updateServiceTrigger)

		case pod.AddPodKind:
			addPodTrigger, err := pod.NewAddPod(sad, &trigger)
			if err != nil {
				result = multierror.Append(result, fmt.Errorf("failed to add new Add Pod Tigger: %v", err))
				break
			}
			triggers = append(triggers, addPodTrigger)

		case pod.DelPodKind:
			delPodTrigger, err := pod.NewDelPod(sad, &trigger)
			if err != nil {
				result = multierror.Append(result, fmt.Errorf("failed to add new Del Pod Tigger: %v", err))
				break
			}
			triggers = append(triggers, delPodTrigger)

		case pod.UpdatePodKind:
			updatePodTrigger, err := pod.NewUpdatePod(sad, &trigger)
			if err != nil {
				result = multierror.Append(result, fmt.Errorf("failed to add new Update Pod Tigger: %v", err))
				break
			}
			triggers = append(triggers, updatePodTrigger)

		case secret.AddSecretKind:
			addSecretTrigger, err := secret.NewAddSecret(sad, &trigger)
			if err != nil {
				result = multierror.Append(result, fmt.Errorf("failed to add new Add Secret Tigger: %v", err))
				break
			}
			triggers = append(triggers, addSecretTrigger)

		case secret.DelSecretKind:
			delSecretTrigger, err := secret.NewDelSecret(sad, &trigger)
			if err != nil {
				result = multierror.Append(result, fmt.Errorf("failed to add new Del Secret Tigger: %v", err))
				break
			}
			triggers = append(triggers, delSecretTrigger)

		case secret.UpdateSecretKind:
			updateSecretTrigger, err := secret.NewUpdateSecret(sad, &trigger)
			if err != nil {
				result = multierror.Append(result, fmt.Errorf("failed to add new Update Secret Tigger: %v", err))
				break
			}
			triggers = append(triggers, updateSecretTrigger)

		default:
			result = multierror.Append(result, fmt.Errorf("Subject Access Delegation does not support Trigger Kind '%s'", trigger.Kind))
		}
	}

	return triggers, result.ErrorOrNil()
}
