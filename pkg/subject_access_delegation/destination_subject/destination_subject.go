package destination_subject

import (
	"github.com/joshvanl/k8s-subject-access-delegation/pkg/subject_access_delegation/destination_subject/service_account"
	"github.com/joshvanl/k8s-subject-access-delegation/pkg/subject_access_delegation/interfaces"
)

const RoleKind = "Role"
const ServiceAccountKind = "ServiceAccount"

func New(sad interfaces.SubjectAccessDelegation) (interfaces.DestinationSubject, error) {
	var destinationSubject interfaces.DestinationSubject

	if sad.Kind() == ServiceAccountKind {
		destinationSubject = service_account.New(sad)
		return destinationSubject, nil
	}

	//unsupported kind
	return nil, nil
}
