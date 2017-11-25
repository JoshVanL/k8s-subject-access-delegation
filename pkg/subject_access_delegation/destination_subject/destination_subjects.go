package destination_subject

import (
	"fmt"

	"github.com/joshvanl/k8s-subject-access-delegation/pkg/subject_access_delegation/destination_subject/pod"
	"github.com/joshvanl/k8s-subject-access-delegation/pkg/subject_access_delegation/destination_subject/service_account"
	"github.com/joshvanl/k8s-subject-access-delegation/pkg/subject_access_delegation/destination_subject/user"
	"github.com/joshvanl/k8s-subject-access-delegation/pkg/subject_access_delegation/interfaces"
)

const PodKind = "Pod"
const ServiceAccountKind = "ServiceAccount"
const UserKind = "User"

func New(sad interfaces.SubjectAccessDelegation, name, kind string) (interfaces.DestinationSubject, error) {
	var destinationSubject interfaces.DestinationSubject

	switch kind {
	case ServiceAccountKind:
		destinationSubject = service_account.New(sad, name)
		return destinationSubject, nil
	case PodKind:
		destinationSubject = pod.New(sad, name)
		return destinationSubject, nil
	case UserKind:
		destinationSubject = user.New(sad, name)
		return destinationSubject, nil
	}

	return nil, fmt.Errorf("Subject Accesss Deletgation does not support Destination Subject Kind '%s'", kind)
}

//func (d *DestinationSubjects) ResolveDestinations() error {
//	var result *multierror.Error
//
//	for _, destinationSubject := range d.destinationSubjects {
//		if err := destinationSubject.ResolveDestination(); err != nil {
//			result = multierror.Append(result, err)
//		}
//	}
//
//	return result.ErrorOrNil()
//}
//
//func (d *DestinationSubjects) Subjects() []interfaces.DestinationSubject {
//	return d.destinationSubjects
//}
