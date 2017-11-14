package destination_subjects

import (
	"errors"
	"fmt"

	"github.com/hashicorp/go-multierror"
	"github.com/sirupsen/logrus"

	"github.com/joshvanl/k8s-subject-access-delegation/pkg/subject_access_delegation/destination_subjects/pod"
	"github.com/joshvanl/k8s-subject-access-delegation/pkg/subject_access_delegation/destination_subjects/service_account"
	"github.com/joshvanl/k8s-subject-access-delegation/pkg/subject_access_delegation/destination_subjects/user"
	"github.com/joshvanl/k8s-subject-access-delegation/pkg/subject_access_delegation/interfaces"
)

const PodKind = "Pod"
const ServiceAccountKind = "ServiceAccount"
const UserKind = "User"

type DestinationSubjects struct {
	log *logrus.Entry

	sad interfaces.SubjectAccessDelegation

	destinationSubjects []interfaces.DestinationSubject
}

var _ interfaces.DestinationSubjects = &DestinationSubjects{}

func New(sad interfaces.SubjectAccessDelegation) (interfaces.DestinationSubjects, error) {
	var result *multierror.Error
	var subjects []interfaces.DestinationSubject

	if len(sad.DestinationSubjects()) == 0 {
		return nil, errors.New("no destinaton subjects given")
	}

	for _, subject := range sad.DestinationSubjects() {

		switch subject.Kind {
		case ServiceAccountKind:
			subjects = append(subjects, service_account.New(sad, subject.Name))
		case PodKind:
			subjects = append(subjects, pod.New(sad, subject.Name))
		case UserKind:
			subjects = append(subjects, user.New(sad, subject.Name))
		default:
			result = multierror.Append(result, fmt.Errorf("Subject Accesss Deletgation does not support Destination Subject Kind '%s'", subject.Kind))
		}
	}

	destinationSubjects := &DestinationSubjects{
		log:                 sad.Log(),
		sad:                 sad,
		destinationSubjects: subjects,
	}

	return destinationSubjects, result.ErrorOrNil()
}

func (d *DestinationSubjects) ResolveDestinations() error {
	var result *multierror.Error

	for _, destinationSubject := range d.destinationSubjects {
		if err := destinationSubject.ResolveDestination(); err != nil {
			result = multierror.Append(result, err)
		}
	}

	return result.ErrorOrNil()
}

func (d *DestinationSubjects) Subjects() []interfaces.DestinationSubject {
	return d.destinationSubjects
}
