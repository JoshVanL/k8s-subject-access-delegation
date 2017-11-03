package destination_subject

import (
	"fmt"

	"github.com/joshvanl/k8s-subject-access-delegation/pkg/subject_access_delegation/destination_subject/service_account"
	"github.com/joshvanl/k8s-subject-access-delegation/pkg/subject_access_delegation/interfaces"
)

const RoleKind = "Role"
const ServiceAccountKind = "ServiceAccount"

func New(sad interfaces.SubjectAccessDelegation) (interfaces.DestinationSubject, error) {
	var destinationSubject interfaces.DestinationSubject

	if sad.DestinationKind() == ServiceAccountKind {
		destinationSubject = service_account.New(sad)
		return destinationSubject, nil
	}

	return nil, fmt.Errorf("Subject Accesss Deletgation does not support Destination Subject Kind '%s'", sad.DestinationKind())
}
