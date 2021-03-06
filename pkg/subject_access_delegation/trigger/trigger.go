package trigger

import (
	"fmt"

	"github.com/hashicorp/go-multierror"

	authzv1alpha1 "github.com/joshvanl/k8s-subject-access-delegation/pkg/apis/authz/v1alpha1"
	"github.com/joshvanl/k8s-subject-access-delegation/pkg/interfaces"
	"github.com/joshvanl/k8s-subject-access-delegation/pkg/subject_access_delegation/trigger/end_points"
	"github.com/joshvanl/k8s-subject-access-delegation/pkg/subject_access_delegation/trigger/node"
	"github.com/joshvanl/k8s-subject-access-delegation/pkg/subject_access_delegation/trigger/pod"
	"github.com/joshvanl/k8s-subject-access-delegation/pkg/subject_access_delegation/trigger/secret"
	"github.com/joshvanl/k8s-subject-access-delegation/pkg/subject_access_delegation/trigger/service"
	"github.com/joshvanl/k8s-subject-access-delegation/pkg/subject_access_delegation/trigger/service_account"
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

		case service_account.AddServiceAccountKind:
			addServiceAccountTrigger, err := service_account.NewAddServiceAccount(sad, &trigger)
			if err != nil {
				result = multierror.Append(result, fmt.Errorf("failed to add new Add ServiceAccount Tigger: %v", err))
				break
			}
			triggers = append(triggers, addServiceAccountTrigger)

		case service_account.DelServiceAccountKind:
			delServiceAccountTrigger, err := service_account.NewDelServiceAccount(sad, &trigger)
			if err != nil {
				result = multierror.Append(result, fmt.Errorf("failed to add new Del ServiceAccount Tigger: %v", err))
				break
			}
			triggers = append(triggers, delServiceAccountTrigger)

		case service_account.UpdateServiceAccountKind:
			updateServiceAccountTrigger, err := service_account.NewUpdateServiceAccount(sad, &trigger)
			if err != nil {
				result = multierror.Append(result, fmt.Errorf("failed to add new Update ServiceAccount Tigger: %v", err))
				break
			}
			triggers = append(triggers, updateServiceAccountTrigger)

		case end_points.AddEndPointsKind:
			addEndPointsTrigger, err := end_points.NewAddEndPoints(sad, &trigger)
			if err != nil {
				result = multierror.Append(result, fmt.Errorf("failed to add new Add EndPoints Tigger: %v", err))
				break
			}
			triggers = append(triggers, addEndPointsTrigger)

		case end_points.DelEndPointsKind:
			delEndPointsTrigger, err := end_points.NewDelEndPoints(sad, &trigger)
			if err != nil {
				result = multierror.Append(result, fmt.Errorf("failed to add new Del EndPoints Tigger: %v", err))
				break
			}
			triggers = append(triggers, delEndPointsTrigger)

		case end_points.UpdateEndPointsKind:
			updateEndPointsTrigger, err := end_points.NewUpdateEndPoints(sad, &trigger)
			if err != nil {
				result = multierror.Append(result, fmt.Errorf("failed to add new Update EndPoints Tigger: %v", err))
				break
			}
			triggers = append(triggers, updateEndPointsTrigger)

		default:
			result = multierror.Append(result, fmt.Errorf("Subject Access Delegation does not support Trigger Kind '%s'", trigger.Kind))
		}
	}

	return triggers, result.ErrorOrNil()
}
