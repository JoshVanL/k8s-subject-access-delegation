package subject_access_delegation

import (
	"errors"
	"fmt"
	"time"

	"github.com/hashicorp/go-multierror"
	"github.com/sirupsen/logrus"
	rbacv1 "k8s.io/api/rbac/v1"
	"k8s.io/apimachinery/pkg/types"
	kubeinformers "k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"

	authzv1alpha1 "github.com/joshvanl/k8s-subject-access-delegation/pkg/apis/authz/v1alpha1"
	"github.com/joshvanl/k8s-subject-access-delegation/pkg/interfaces"
	"github.com/joshvanl/k8s-subject-access-delegation/pkg/subject_access_delegation/destination_subject"
	"github.com/joshvanl/k8s-subject-access-delegation/pkg/subject_access_delegation/origin_subject"
	"github.com/joshvanl/k8s-subject-access-delegation/pkg/subject_access_delegation/role_binding"
	"github.com/joshvanl/k8s-subject-access-delegation/pkg/subject_access_delegation/trigger"
	"github.com/joshvanl/k8s-subject-access-delegation/pkg/subject_access_delegation/utils"
)

type SubjectAccessDelegation struct {
	controller interfaces.Controller
	log        *logrus.Entry

	sad                 *authzv1alpha1.SubjectAccessDelegation
	kubeInformerFactory kubeinformers.SharedInformerFactory
	client              kubernetes.Interface

	originSubject       interfaces.OriginSubject
	destinationSubjects []interfaces.DestinationSubject
	triggers            []interfaces.Trigger
	deletionTimeStamp   time.Time

	roleBindings    []interfaces.Binding
	clusterBindings []*rbacv1.RoleBinding
	bindingSubjects []rbacv1.Subject

	triggered   bool
	stopCh      chan struct{}
	clockOffset time.Duration
}

var _ interfaces.SubjectAccessDelegation = &SubjectAccessDelegation{}

func New(controller interfaces.Controller, sad *authzv1alpha1.SubjectAccessDelegation, log *logrus.Entry, kubeInformerFactory kubeinformers.SharedInformerFactory, client kubernetes.Interface, clockOffset time.Duration) *SubjectAccessDelegation {
	return &SubjectAccessDelegation{
		controller:          controller,
		log:                 log,
		client:              client,
		kubeInformerFactory: kubeInformerFactory,
		sad:                 sad,
		roleBindings:        make([]interfaces.Binding, 0),
		bindingSubjects:     make([]rbacv1.Subject, 0),
		triggered:           false,
		stopCh:              make(chan struct{}),
		clockOffset:         clockOffset,
	}
}

func (s *SubjectAccessDelegation) Delegate() (closed bool, err error) {
	for i := 0; i < s.Repeat(); i++ {
		s.log.Infof("Subject Access Delegation \"%s\" (%d/%d)", s.Name(), i+1, s.Repeat())

		s.triggered = false

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
		s.triggered = true

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

	s.bindingSubjects = s.buildDestinationSubjects()

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
	if s.RealNow().After(s.deletionTimeStamp) {
		return
	}

	ticker := time.After(time.Until(s.RealTime(s.deletionTimeStamp)))

	select {
	case <-ticker:
		return
	case <-s.stopCh:
		return
	}
}

func (s *SubjectAccessDelegation) ApplyDelegation() error {
	s.log.Infof("Applying Subject Access Delegation '%s'", s.Name())

	bindings, err := s.buildRoleBindings()
	if err != nil {
		return fmt.Errorf("failed to build rolebindings: %v", err)
	}

	var result *multierror.Error

	for _, roleBinding := range bindings {
		if err := s.createRoleBinding(roleBinding); err != nil {
			result = multierror.Append(result, err)
		}
	}

	return result.ErrorOrNil()
}

func (s *SubjectAccessDelegation) buildRoleBindings() ([]interfaces.Binding, error) {
	var roleBindings []interfaces.Binding

	roleRefs, clusterRoleRefs, err := s.originSubject.RoleRefs()
	if err != nil {
		return nil, fmt.Errorf("failed to resolve Role References: %v", err)
	}

	for _, roleRef := range roleRefs {
		roleBindings = append(roleBindings, role_binding.NewRoleBinding(s, roleRef))
	}

	for _, roleRef := range clusterRoleRefs {
		roleBindings = append(roleBindings, role_binding.NewRoleBinding(s, roleRef))
	}

	return roleBindings, nil
}

func (s *SubjectAccessDelegation) DeleteRoleBindings() error {
	var result *multierror.Error

	for _, binding := range s.roleBindings {
		if err := s.deleteRoleBinding(binding); err != nil {
			result = multierror.Append(result, err)
		}
	}

	s.roleBindings = make([]interfaces.Binding, 0)

	return result.ErrorOrNil()
}

func (s *SubjectAccessDelegation) createRoleBinding(binding interfaces.Binding) error {
	binding, err := binding.CreateRoleBinding()
	if err != nil {
		return fmt.Errorf("failed to create role binding: %v", err)
	}

	s.log.Infof("Role Binding '%s' Created", binding.Name())

	s.roleBindings = append(s.roleBindings, binding)

	return nil
}

func (s *SubjectAccessDelegation) deleteRoleBinding(binding interfaces.Binding) error {
	if err := binding.DeleteRoleBinding(); err != nil {
		return err
	}

	s.log.Infof("Role Binding '%s' Deleted", binding.Name)

	return nil
}

func (s *SubjectAccessDelegation) ActivateTriggers() (closed bool, err error) {
	s.log.Debug("Activating Triggers")
	for _, trigger := range s.triggers {
		trigger.Activate()
	}

	s.log.Info("Triggers Activated")

	ready := false

	for !ready {
		if s.waitOnTriggers() {
			return true, nil
		}

		s.log.Info("All triggers have been satisfied, checking still true")

		ready = s.checkTriggers()
		if !ready {
			s.log.Info("Not all triggers ready at the same time, re-waiting")
		}
	}

	s.log.Info("All triggers ready")

	return false, nil
}

func (s *SubjectAccessDelegation) waitOnTriggers() (closed bool) {
	for _, trigger := range s.triggers {
		if trigger.WaitOn() {
			return true
		}
	}

	return false
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

func (s *SubjectAccessDelegation) AddRoleBinding(addBinding interfaces.Binding) error {
	if s.triggered {
		// We need to create a rolebinding for sad, not the origin subject one
		if err := s.createRoleBinding(addBinding); err != nil {
			return err
		}
	}

	return nil
}

// This is a bit wrong, although a role ref may be duped and this is fine, the
// names will be wrong
func (s *SubjectAccessDelegation) DeleteRoleBinding(delBinding interfaces.Binding) bool {
	if s.triggered {
		for i, binding := range s.roleBindings {
			if binding.RoleRef().Name == delBinding.RoleRef().Name && binding.RoleRef().Kind == delBinding.RoleRef().Kind {

				copy(s.roleBindings[i:], s.roleBindings[i+1:])
				s.roleBindings[len(s.roleBindings)-1] = nil
				s.roleBindings = s.roleBindings[:len(s.roleBindings)-1]

				if err := s.deleteRoleBinding(binding); err != nil {
					s.log.Errorf("Tryed to delete role binding but something went very wrong: %v", err)
				}

				return true
			}
		}

	}

	return false
}

// Here we will want to delete all the new bindings THEN delete all the old or
// via versa
func (s *SubjectAccessDelegation) UpdateRoleBinding(old, new interfaces.Binding) error {
	var result *multierror.Error

	if s.triggered {
		if err := s.deleteRoleBinding(old); err != nil {
			result = multierror.Append(result, err)
		}

		if err := s.createRoleBinding(new); err != nil {
			result = multierror.Append(result, err)
		}
	}

	return result.ErrorOrNil()
}

func (s *SubjectAccessDelegation) buildDestinationSubjects() []rbacv1.Subject {
	var subjects []rbacv1.Subject

	for _, destinationSubject := range s.DestinationSubjects() {
		subjects = append(subjects, rbacv1.Subject{Name: destinationSubject.Name(), Kind: destinationSubject.Kind(), Namespace: s.Namespace()})
	}

	return subjects
}

func (s *SubjectAccessDelegation) RealTime(time time.Time) time.Time {
	return time.Add(s.clockOffset)
}

func (s *SubjectAccessDelegation) RealNow() time.Time {
	return time.Now().Add(s.clockOffset)
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

func (s *SubjectAccessDelegation) SeenUid(uid types.UID) bool {
	return s.controller.SeenUid(uid)
}

func (s *SubjectAccessDelegation) DeletedUid(uid types.UID) bool {
	return s.controller.SeenUid(uid)
}

func (s *SubjectAccessDelegation) AddUid(uid types.UID) {
	s.controller.AddUid(uid)
}

func (s *SubjectAccessDelegation) DeleteUid(uid types.UID) {
	s.controller.DeleteUid(uid)
}

func (s *SubjectAccessDelegation) BindingSubjects() []rbacv1.Subject {
	return s.bindingSubjects
}
