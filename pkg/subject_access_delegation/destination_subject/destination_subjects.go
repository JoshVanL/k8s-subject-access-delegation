package destination_subject

import (
	"fmt"

	"github.com/joshvanl/k8s-subject-access-delegation/pkg/interfaces"
	"github.com/joshvanl/k8s-subject-access-delegation/pkg/subject_access_delegation/destination_subject/group"
	"github.com/joshvanl/k8s-subject-access-delegation/pkg/subject_access_delegation/destination_subject/service_account"
	"github.com/joshvanl/k8s-subject-access-delegation/pkg/subject_access_delegation/destination_subject/user"
)

const ServiceAccountKind = "ServiceAccount"
const UserKind = "User"
const GroupKind = "Group"

func New(sad interfaces.SubjectAccessDelegation, name, kind string) ([]interfaces.DestinationSubject, error) {

	switch kind {
	case ServiceAccountKind:
		destinationSubjects, err := service_account.New(sad, name)
		if err != nil {
			return nil, err
		}
		return destinationSubjects, nil
	case UserKind:
		destinationSubject := user.New(sad, name)
		return []interfaces.DestinationSubject{destinationSubject}, nil
	case GroupKind:
		destinationSubject := group.New(sad, name)
		return []interfaces.DestinationSubject{destinationSubject}, nil
	default:
		return nil, fmt.Errorf("Subject Accesss Deletgation does not support Destination Subject Kind '%s'", kind)
	}

}
