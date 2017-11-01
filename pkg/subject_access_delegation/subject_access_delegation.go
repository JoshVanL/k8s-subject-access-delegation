package subject_access_delegation

import (
	"fmt"

	"github.com/sirupsen/logrus"
	"k8s.io/client-go/kubernetes"

	authzv1alpha1 "github.com/joshvanl/k8s-subject-access-delegation/pkg/apis/authz/v1alpha1"
	"github.com/joshvanl/k8s-subject-access-delegation/pkg/subject_access_delegation/destination_subject"
	"github.com/joshvanl/k8s-subject-access-delegation/pkg/subject_access_delegation/interfaces"
	"github.com/joshvanl/k8s-subject-access-delegation/pkg/subject_access_delegation/origin_subject"
)

type SubjectAccessDelegation struct {
	log *logrus.Entry

	sad       *authzv1alpha1.SubjectAccessDelegation
	namespace string
	client    kubernetes.Interface

	originSubject      interfaces.OriginSubject
	destinationSubject interfaces.DestinationSubject
}

func New(log *logrus.Entry, sad *authzv1alpha1.SubjectAccessDelegation, namespace string, client kubernetes.Interface) *SubjectAccessDelegation {
	return &SubjectAccessDelegation{
		log:    log,
		sad:    sad,
		client: client,
	}
}

func (s *SubjectAccessDelegation) GetSubjects() error {
	originSubject, err := s.getOriginSubject()
	if err != nil {
		return fmt.Errorf("failed to resolve origin subject: %v", err)
	}

	destinationSubject, err := s.getDestinationSubject()
	if err != nil {
		return fmt.Errorf("failed to resolve destination subject: %v", err)
	}

	s.originSubject = originSubject
	s.destinationSubject = destinationSubject

	return nil
}

func (s *SubjectAccessDelegation) getOriginSubject() (interfaces.OriginSubject, error) {
	return origin_subject.New(s)
}

func (s *SubjectAccessDelegation) getDestinationSubject() (interfaces.DestinationSubject, error) {

	return destination_subject.New(s)
}

func (s *SubjectAccessDelegation) Log() *logrus.Entry {
	return s.log
}

func (s *SubjectAccessDelegation) Namespace() string {
	return s.namespace
}

func (s *SubjectAccessDelegation) Kind() string {
	return s.sad.Kind
}

func (s *SubjectAccessDelegation) Client() kubernetes.Interface {
	return s.client
}

func (s *SubjectAccessDelegation) Name() string {
	return s.sad.Name
}
