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

func New(sad *authzv1alpha1.SubjectAccessDelegation, client kubernetes.Interface, log *logrus.Entry) *SubjectAccessDelegation {
	return &SubjectAccessDelegation{
		log:       log,
		sad:       sad,
		client:    client,
		namespace: sad.Namespace,
	}
}

func (s *SubjectAccessDelegation) Delegate() error {
	if err := s.GetSubjects(); err != nil {
		return err
	}

	return nil
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
	originSubject, err := origin_subject.New(s)
	if err != nil {
		return nil, err
	}

	if err := originSubject.Origin(); err != nil {
		return nil, err
	}

	return originSubject, nil
}

func (s *SubjectAccessDelegation) getDestinationSubject() (interfaces.DestinationSubject, error) {
	destinationSubject, err := destination_subject.New(s)
	if err != nil {
		return nil, err
	}

	if err := destinationSubject.Destination(); err != nil {
		return nil, err
	}

	return destinationSubject, nil
}

func (s *SubjectAccessDelegation) Delete() error {
	return nil
}

func (s *SubjectAccessDelegation) Log() *logrus.Entry {
	return s.log
}

func (s *SubjectAccessDelegation) Namespace() string {
	return s.namespace
}

func (s *SubjectAccessDelegation) Kind() string {
	return s.sad.Name
}

func (s *SubjectAccessDelegation) Client() kubernetes.Interface {
	return s.client
}

func (s *SubjectAccessDelegation) Name() string {
	return s.sad.Name
}

func (s *SubjectAccessDelegation) OriginName() string {
	return s.sad.Spec.OriginSubject.Name
}

func (s *SubjectAccessDelegation) DestinationName() string {
	return s.sad.Spec.DestinationSubject.Name
}
