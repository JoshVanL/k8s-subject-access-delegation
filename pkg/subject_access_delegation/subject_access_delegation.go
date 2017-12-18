package subject_access_delegation

import (
	"errors"
	"fmt"
	"time"

	"github.com/hashicorp/go-multierror"
	"github.com/sirupsen/logrus"
	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	kubeinformers "k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"

	authzv1alpha1 "github.com/joshvanl/k8s-subject-access-delegation/pkg/apis/authz/v1alpha1"
	"github.com/joshvanl/k8s-subject-access-delegation/pkg/subject_access_delegation/destination_subject"
	"github.com/joshvanl/k8s-subject-access-delegation/pkg/subject_access_delegation/interfaces"
	"github.com/joshvanl/k8s-subject-access-delegation/pkg/subject_access_delegation/origin_subject"
	"github.com/joshvanl/k8s-subject-access-delegation/pkg/subject_access_delegation/trigger"
	"github.com/joshvanl/k8s-subject-access-delegation/pkg/subject_access_delegation/utils"
)

type SubjectAccessDelegation struct {
	log *logrus.Entry

	sad                 *authzv1alpha1.SubjectAccessDelegation
	kubeInformerFactory kubeinformers.SharedInformerFactory
	client              kubernetes.Interface

	originSubject       interfaces.OriginSubject
	destinationSubjects []interfaces.DestinationSubject
	triggers            []interfaces.Trigger
	roleBindings        []*rbacv1.RoleBinding
	deletionTimeStamp   time.Time
	stopCh              chan struct{}
}

func New(sad *authzv1alpha1.SubjectAccessDelegation, log *logrus.Entry, kubeInformerFactory kubeinformers.SharedInformerFactory, client kubernetes.Interface) *SubjectAccessDelegation {
	return &SubjectAccessDelegation{
		log:                 log,
		client:              client,
		kubeInformerFactory: kubeInformerFactory,
		sad:                 sad,
		stopCh:              make(chan struct{}),
	}
}

func (s *SubjectAccessDelegation) Delegate() (closed bool, err error) {
	for i := 0; i < s.Repeat(); i++ {
		s.log.Infof("Subject Access Delegation \"%s\" (%d/%d)", s.Name(), i+1, s.Repeat())

		if err := s.initDelegation(); err != nil {
			return false, fmt.Errorf("error initiating Subject Access Delegation: %v", err)
		}

		closed, err := s.ActivateTriggers()
		if err != nil {
			return false, err
		}
		if closed {
			s.log.Infof("A Trigger was found closed, exiting")
			return true, nil
		}

		s.log.Infof("All triggers fired!")

		if err := s.ApplyDelegation(); err != nil {
			return false, fmt.Errorf("failed to apply delegation: %v", err)
		}

		if err := s.ParseDeletionTime(); err != nil {
			return false, err
		}

		s.waitOnDeletion()
		if err := s.DeleteRoleBindings(); err != nil {
			return false, err
		}
	}

	s.log.Infof("Subject Access Delegation '%s' has completed", s.Name())

	return false, nil
}

func (s *SubjectAccessDelegation) initDelegation() error {
	var result *multierror.Error

	if err := s.ParseDeletionTime(); err != nil {
		result = multierror.Append(result, err)
	}

	if err := s.GetSubjects(); err != nil {
		result = multierror.Append(result, err)
	}

	if err := s.BuildTriggers(); err != nil {
		result = multierror.Append(result, err)
	}

	return result.ErrorOrNil()
}

func (s *SubjectAccessDelegation) ParseDeletionTime() error {
	t, err := utils.ParseTime(s.sad.Spec.DeletionTime)
	if err != nil {
		return fmt.Errorf("failed to parse deletion time stamp: %v", err)
	}

	s.deletionTimeStamp = t

	return nil
}

func (s *SubjectAccessDelegation) waitOnDeletion() {
	if time.Now().After(s.deletionTimeStamp) {
		return
	}

	ticker := time.After(time.Until(s.deletionTimeStamp))

	select {
	case <-ticker:
		return
	case <-s.stopCh:
		return
	}
}

func (s *SubjectAccessDelegation) ApplyDelegation() error {
	s.log.Infof("Applying Subject Access Delegation '%s'", s.Name())

	if err := s.buildRoleBindings(); err != nil {
		return fmt.Errorf("failed to build rolebindings: %v", err)
	}

	return s.applyRoleBindings()
}

func (s *SubjectAccessDelegation) applyRoleBindings() error {
	var result *multierror.Error

	for _, roleBinding := range s.roleBindings {
		_, err := s.client.Rbac().RoleBindings(s.Namespace()).Create(roleBinding)
		if err != nil {
			result = multierror.Append(result, fmt.Errorf("failed to create role binding: %v", err))
		} else {
			s.log.Infof("Role Binding '%s' Created", roleBinding.Name)
		}
	}

	return result.ErrorOrNil()
}

func (s *SubjectAccessDelegation) DeleteRoleBindings() error {
	var result *multierror.Error
	options := &metav1.DeleteOptions{}

	for _, roleBinding := range s.roleBindings {
		err := s.client.Rbac().RoleBindings(s.Namespace()).Delete(roleBinding.Name, options)
		if err != nil {
			result = multierror.Append(result, fmt.Errorf("failed to delete role binding: %v", err))
		} else {
			s.log.Infof("Role Binding '%s' Deleted", roleBinding.Name)
		}
	}

	s.roleBindings = make([]*rbacv1.RoleBinding, 0)

	return result.ErrorOrNil()
}

func (s *SubjectAccessDelegation) buildRoleBindings() error {
	var roleBindings []*rbacv1.RoleBinding
	var subjects []rbacv1.Subject
	var roleRefs []*rbacv1.RoleRef
	var err error

	roleRefs, err = s.originSubject.RoleRefs()
	if err != nil {
		return fmt.Errorf("failed to resolve Role References: %v", err)
	}

	for _, destinationSubject := range s.DestinationSubjects() {
		subjects = append(subjects, rbacv1.Subject{Name: destinationSubject.Name(), Kind: destinationSubject.Kind()})
	}

	for _, roleRef := range roleRefs {
		name := fmt.Sprintf("%s-%s-%s", s.Name(), s.Namespace(), roleRef.Name)
		roleBinding := &rbacv1.RoleBinding{
			ObjectMeta: metav1.ObjectMeta{Name: name, Namespace: s.Namespace()},
			RoleRef:    *roleRef,
			Subjects:   subjects,
		}
		//timestamp := metav1.Time{
		//	Time: time.Now(),
		//}
		//roleBinding.CreationTimestamp = timestamp

		roleBindings = append(roleBindings, roleBinding)
	}

	s.roleBindings = roleBindings

	return nil
}

func (s *SubjectAccessDelegation) ActivateTriggers() (closed bool, err error) {
	s.log.Debugf("Activating Triggers")
	for _, trigger := range s.triggers {
		trigger.Activate()
	}
	s.log.Debugf("Triggers Activated")

	ready := false

	for !ready {
		closed, err := s.waitOnTriggers()
		if err != nil {
			return false, fmt.Errorf("error waiting on triggers to fire: %v", err)
		}
		if closed {
			return true, nil
		}

		s.log.Debugf("All triggers have been satisfied, checking still true")

		ready = s.checkTriggers()
		if !ready {
			s.log.Debug("Not all triggers ready at the same time, re-waiting")
		}
	}

	s.log.Debug("All triggers ready")

	return false, nil
}

func (s *SubjectAccessDelegation) waitOnTriggers() (closed bool, err error) {
	for _, trigger := range s.triggers {
		//closed, err := trigger.WaitOn()
		//if err != nil {
		//	return false, fmt.Errorf("error waiting on trigger to fire: %v", err)
		//}
		if trigger.WaitOn() {
			return true, nil
		}
	}

	return false, nil
}

func (s *SubjectAccessDelegation) checkTriggers() (ready bool) {
	for _, trigger := range s.triggers {
		ready := trigger.Completed()
		if !ready {
			return false
		}
	}

	return true
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
	var result *multierror.Error

	originSubject, err := s.getOriginSubject()
	if err != nil {
		result = multierror.Append(result, fmt.Errorf("failed to resolve origin subject: %v", err))
	}

	destinationSubjects, err := s.getDestinationSubjects()
	if err != nil {
		result = multierror.Append(result, fmt.Errorf("failed to resolve destination subjects: %v", err))
	}

	if result == nil {
		s.originSubject = originSubject
		s.destinationSubjects = destinationSubjects
	}

	if err := s.ResolveDestinations(); err != nil {
		result = multierror.Append(result, err)
	}

	return result.ErrorOrNil()
}

func (s *SubjectAccessDelegation) getOriginSubject() (interfaces.OriginSubject, error) {
	originSubject, err := origin_subject.New(s, s.originName(), s.originKind())
	if err != nil {
		return nil, err
	}

	if err := originSubject.ResolveOrigin(); err != nil {
		return nil, err
	}

	return originSubject, nil
}

func (s *SubjectAccessDelegation) getDestinationSubjects() ([]interfaces.DestinationSubject, error) {
	var result *multierror.Error
	var destinationSubjects []interfaces.DestinationSubject

	if len(s.sad.Spec.DestinationSubjects) == 0 {
		return nil, errors.New("no destination subjects specified")
	}

	for _, destinationSubject := range s.sad.Spec.DestinationSubjects {
		subject, err := destination_subject.New(s, destinationSubject.Name, destinationSubject.Kind)

		if err != nil {
			result = multierror.Append(result, err)
		} else {
			destinationSubjects = append(destinationSubjects, subject)
		}
	}

	if result != nil {
		return nil, result.ErrorOrNil()
	}

	return destinationSubjects, result.ErrorOrNil()
}

func (s *SubjectAccessDelegation) ResolveDestinations() error {
	var result *multierror.Error

	for _, destinationSubject := range s.destinationSubjects {
		if err := destinationSubject.ResolveDestination(); err != nil {
			result = multierror.Append(result, err)
		}
	}

	return result.ErrorOrNil()
}

func (s *SubjectAccessDelegation) Delete() error {
	s.log.Debugf("Attempting to delete delegation '%s' triggers", s.Name())

	var result *multierror.Error
	for _, trigger := range s.triggers {
		if err := trigger.Delete(); err != nil {
			result = multierror.Append(result, err)
		}
	}

	close(s.stopCh)

	return result.ErrorOrNil()
}

func (s *SubjectAccessDelegation) Log() *logrus.Entry {
	return s.log
}

func (s *SubjectAccessDelegation) Namespace() string {
	return s.sad.Namespace
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

func (s *SubjectAccessDelegation) originName() string {
	return s.sad.Spec.OriginSubject.Name
}

func (s *SubjectAccessDelegation) originKind() string {
	return s.sad.Spec.OriginSubject.Kind
}

func (s *SubjectAccessDelegation) OriginSubject() interfaces.OriginSubject {
	return s.originSubject
}

func (s *SubjectAccessDelegation) DestinationSubjects() []interfaces.DestinationSubject {
	return s.destinationSubjects
}

func (s *SubjectAccessDelegation) Triggers() []authzv1alpha1.EventTrigger {
	return s.sad.Spec.EventTriggers
}

func (s *SubjectAccessDelegation) Repeat() int {
	return s.sad.Spec.Repeat
}

func (s *SubjectAccessDelegation) KubeInformerFactory() kubeinformers.SharedInformerFactory {
	return s.kubeInformerFactory
}
