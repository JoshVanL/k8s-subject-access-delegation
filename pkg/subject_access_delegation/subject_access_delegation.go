package subject_access_delegation

import (
	"fmt"

	"github.com/hashicorp/go-multierror"
	"github.com/sirupsen/logrus"
	"k8s.io/client-go/kubernetes"

	authzv1alpha1 "github.com/joshvanl/k8s-subject-access-delegation/pkg/apis/authz/v1alpha1"
	"github.com/joshvanl/k8s-subject-access-delegation/pkg/subject_access_delegation/destination_subject"
	"github.com/joshvanl/k8s-subject-access-delegation/pkg/subject_access_delegation/interfaces"
	"github.com/joshvanl/k8s-subject-access-delegation/pkg/subject_access_delegation/origin_subject"
	"github.com/joshvanl/k8s-subject-access-delegation/pkg/subject_access_delegation/trigger"
)

type SubjectAccessDelegation struct {
	log *logrus.Entry

	sad       *authzv1alpha1.SubjectAccessDelegation
	namespace string
	stopChs   []chan struct{}
	client    kubernetes.Interface

	originSubject      interfaces.OriginSubject
	destinationSubject interfaces.DestinationSubject
	triggers           []interfaces.Trigger
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
	for i := 0; i < s.Repeat(); i++ {
		s.log.Infof("Subject Access Delegation \"%s\" (%d/%d)", s.Name(), i+1, s.Repeat())

		if err := s.GetSubjects(); err != nil {
			return err
		}

		if err := s.BuildTriggers(); err != nil {
			return err
		}

		if err := s.ActivateTriggers(); err != nil {
			return err
		}

		s.log.Infof("All triggers fired!")
	}

	return nil
}

func (s *SubjectAccessDelegation) ActivateTriggers() error {
	s.log.Debugf("Activating Triggers")
	for _, trigger := range s.triggers {
		go trigger.Activate()
	}

	var err error
	ready := false

	for !ready {
		if err = s.waitOnTriggers(); err != nil {
			return fmt.Errorf("error waiting on triggers to fire: %v", err)
		}

		s.log.Debugf("All triggers have been satisfied, checking still true")

		ready, err = s.checkTriggers()
		if err != nil {
			return fmt.Errorf("error waiting on triggers to fire: %v", err)
		}

		if !ready {
			s.log.Debug("Not all triggers ready at the same time, re-waiting")
		}
	}

	return nil
}

func (s *SubjectAccessDelegation) waitOnTriggers() error {
	for _, trigger := range s.triggers {
		if err := trigger.WaitOn(); err != nil {
			return fmt.Errorf("error waiting on trigger to fire: %v", err)
		}
	}

	return nil
}

func (s *SubjectAccessDelegation) checkTriggers() (ready bool, err error) {
	for _, trigger := range s.triggers {
		ready, err := trigger.Ready()
		if err != nil {
			return false, fmt.Errorf("error checking trigger status: %v", err)
		}

		if !ready {
			return false, nil
		}
	}

	return true, nil
}

func (s *SubjectAccessDelegation) BuildTriggers() error {
	triggers, err := trigger.New(s)
	if err != nil {
		return fmt.Errorf("failed to build triggers: %v", err)
	}

	s.triggers = triggers
	return nil
}

func (s *SubjectAccessDelegation) GetSubjects() error {
	var result error

	originSubject, err := s.getOriginSubject()
	if err != nil {
		result = multierror.Append(result, fmt.Errorf("failed to resolve origin subject: %v", err))
	}

	destinationSubject, err := s.getDestinationSubject()
	if err != nil {
		result = multierror.Append(result, fmt.Errorf("failed to resolve destination subject: %v", err))
	}

	if result == nil {
		s.originSubject = originSubject
		s.destinationSubject = destinationSubject
	}

	return result
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

func (s *SubjectAccessDelegation) OriginKind() string {
	return s.sad.Spec.OriginSubject.Kind
}

func (s *SubjectAccessDelegation) DestinationKind() string {
	return s.sad.Spec.DestinationSubject.Kind
}

func (s *SubjectAccessDelegation) DestinationName() string {
	return s.sad.Spec.DestinationSubject.Name
}

func (s *SubjectAccessDelegation) Duration() int64 {
	return s.sad.Spec.Duration
}

func (s *SubjectAccessDelegation) Repeat() int {
	return s.sad.Spec.Repeat
}
