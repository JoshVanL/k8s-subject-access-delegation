package subject_access_delegation

import (
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/hashicorp/go-multierror"
	"github.com/sirupsen/logrus"
	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	kubeinformers "k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/util/retry"

	authzv1alpha1 "github.com/joshvanl/k8s-subject-access-delegation/pkg/apis/authz/v1alpha1"
	clientset "github.com/joshvanl/k8s-subject-access-delegation/pkg/client/clientset/versioned"
	"github.com/joshvanl/k8s-subject-access-delegation/pkg/interfaces"
	"github.com/joshvanl/k8s-subject-access-delegation/pkg/subject_access_delegation/destination_subject"
	"github.com/joshvanl/k8s-subject-access-delegation/pkg/subject_access_delegation/origin_subject"
	"github.com/joshvanl/k8s-subject-access-delegation/pkg/subject_access_delegation/role_binding"
)

type SubjectAccessDelegation struct {
	controller interfaces.Controller
	log        *logrus.Entry

	kubeInformerFactory kubeinformers.SharedInformerFactory
	client              kubernetes.Interface
	sad                 *authzv1alpha1.SubjectAccessDelegation
	sadclientset        clientset.Interface

	originSubject       interfaces.OriginSubject
	destinationSubjects []interfaces.DestinationSubject
	triggers            []interfaces.Trigger
	deletionTriggers    []interfaces.Trigger
	deletionTimeStamp   time.Time

	roleBindings    map[string]interfaces.Binding
	bindingSubjects []rbacv1.Subject

	triggered   bool
	stopCh      chan struct{}
	isActive    bool
	revert      chan bool
	clockOffset time.Duration
	mx          *sync.Mutex
}

var (
	_ interfaces.SubjectAccessDelegation = &SubjectAccessDelegation{}
)

func New(controller interfaces.Controller,
	sad *authzv1alpha1.SubjectAccessDelegation,
	log *logrus.Entry,
	kubeInformerFactory kubeinformers.SharedInformerFactory,
	client kubernetes.Interface,
	sadclientset clientset.Interface,
	clockOffset time.Duration) *SubjectAccessDelegation {

	return &SubjectAccessDelegation{
		controller:          controller,
		log:                 log,
		client:              client,
		kubeInformerFactory: kubeInformerFactory,
		sadclientset:        sadclientset,
		sad:                 sad,
		roleBindings:        make(map[string]interfaces.Binding),
		bindingSubjects:     make([]rbacv1.Subject, 0),
		triggered:           false,
		stopCh:              make(chan struct{}),
		isActive:            false,
		clockOffset:         clockOffset,
		mx:                  new(sync.Mutex),
	}
}

func (s *SubjectAccessDelegation) Delegate() (closed bool, initError bool, err error) {
	initError = true

	s.sad.Status.Processed = true
	if err := s.updateRemoteSAD(); err != nil {
		return false, initError, fmt.Errorf("failed to set SAD Processed to true: %v", err)
	}

	for i := s.sad.Status.Iteration; i < s.Repeat(); i++ {
		if i > 0 {
			initError = false
		}

		s.isActive = false

		if err := s.updateLocalSAD(); err != nil {
			return false, initError, err
		}

		s.log.Infof("Subject Access Delegation '%s' (%d/%d)", s.Name(), i+1, s.Repeat())

		s.triggered = false

		if err := s.InitDelegation(); err != nil {
			return false, initError, fmt.Errorf("error initiating Subject Access Delegation: %v", err)
		}

		closed, err := s.ActivateTriggers()
		if err != nil {
			return false, false, err
		}

		revert, ok := <-s.revert
		if ok && revert {
			i -= 1
			continue
		}

		if closed {
			s.log.Debugf("A Trigger was found closed, exiting.")
			return true, false, nil
		}

		s.isActive = true

		if err := s.ApplyDelegation(); err != nil {
			return false, false, fmt.Errorf("failed to apply delegation: %v", err)
		}

		revert = true
		for revert {
			if err := s.BuildDeletionTriggers(); err != nil {
				return false, false, err
			}

			closed, err = s.ActivateDeletionTriggers()
			if err != nil {
				return false, false, err
			}

			revert = false
			r, ok := <-s.revert
			if ok {
				revert = r
			}

		}

		if closed {
			s.log.Debugf("A Deletion Trigger was found closed, exiting.")
			return true, false, nil
		}

		if err := s.DeleteRoleBindings(); err != nil {
			return false, false, err
		}

		if err := s.updateInteration(i + 1); err != nil {
			return false, false, err

		}

	}

	s.log.Infof("Subject Access Delegation '%s' has completed", s.Name())

	return false, false, nil
}

func (s *SubjectAccessDelegation) InitDelegation() error {
	var result *multierror.Error

	if err := s.BuildTriggers(); err != nil {
		result = multierror.Append(result, err)
	}

	if err := s.BuildDeletionTriggers(); err != nil {
		result = multierror.Append(result, err)
	}

	if err := s.GetSubjects(); err != nil {
		result = multierror.Append(result, err)
	}

	s.bindingSubjects = s.buildDestinationSubjects()

	return result.ErrorOrNil()
}

func (s *SubjectAccessDelegation) UpdateSadObject(sad *authzv1alpha1.SubjectAccessDelegation) (bool, error) {
	var result *multierror.Error

	s.mx.Lock()
	s.revert = make(chan bool)

	triggerChanged := false
	subjectChanged := true

	for _, newTrigger := range sad.Spec.EventTriggers {
		triggerChanged = true

		for _, oldTrigger := range s.sad.Spec.EventTriggers {
			if newTrigger.Value == oldTrigger.Value && newTrigger.Kind == oldTrigger.Kind && newTrigger.Replicas == oldTrigger.Replicas {
				triggerChanged = false
				break
			}
		}

		if triggerChanged {
			break
		}
	}

	if !triggerChanged {
		for _, newTrigger := range sad.Spec.DeletionTriggers {
			triggerChanged = true

			for _, oldTrigger := range s.sad.Spec.DeletionTriggers {
				if newTrigger.Value == oldTrigger.Value && newTrigger.Kind == oldTrigger.Kind && newTrigger.Replicas == oldTrigger.Replicas {
					triggerChanged = false
					break
				}
			}

			if triggerChanged {
				break
			}
		}
	}

	if triggerChanged {
		for _, trigger := range s.triggers {
			if err := trigger.Delete(); err != nil {
				result = multierror.Append(result, err)
			}
		}

		for _, delTrigger := range s.deletionTriggers {
			if err := delTrigger.Delete(); err != nil {
				result = multierror.Append(result, err)
			}
		}
		s.triggers = []interfaces.Trigger{}
		s.deletionTriggers = []interfaces.Trigger{}
	}

	if sad.Spec.OriginSubject.Name == s.sad.Spec.OriginSubject.Name && sad.Spec.OriginSubject.Kind == s.sad.Spec.OriginSubject.Kind {
		subjectChanged = false
	}

	if !subjectChanged {
		subjectChanged = true

		for _, newSubject := range sad.Spec.DestinationSubjects {
			for _, oldSubject := range s.sad.Spec.DestinationSubjects {
				if newSubject.Name == oldSubject.Name && newSubject.Kind == oldSubject.Kind {
					subjectChanged = false
					break
				}
			}

			if subjectChanged {
				break
			}
		}
	}

	if !subjectChanged && !triggerChanged {
		close(s.revert)
		s.mx.Unlock()
		return false, nil
	}

	if s.isActive {
		if err := s.DeleteRoleBindings(); err != nil {
			result = multierror.Append(result, err)
		}
	}

	s.sad = sad.DeepCopy()

	if err := s.GetSubjects(); err != nil {
		result = multierror.Append(result, err)
	}

	if s.isActive {
		if err := s.ApplyDelegation(); err != nil {
			result = multierror.Append(result, err)
		}
		if triggerChanged {
			s.revert <- triggerChanged
		}
	}

	close(s.revert)
	s.mx.Unlock()

	if result.ErrorOrNil() != nil {
		return false, result.ErrorOrNil()
	}

	return true, nil
}

func (s *SubjectAccessDelegation) ApplyDelegation() error {
	s.log.Infof("Applying Subject Access Delegation '%s'", s.Name())

	bindings, err := s.buildRoleBindings()
	if err != nil {
		return fmt.Errorf("failed to build role bindings: %v", err)
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

	roleRefs, clusterRoleRefs := s.originSubject.RoleRefs()

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

	s.roleBindings = make(map[string]interfaces.Binding)

	return result.ErrorOrNil()
}

func (s *SubjectAccessDelegation) createRoleBinding(binding interfaces.Binding) error {
	binding, exists, err := binding.CreateRoleBinding()
	if err != nil {
		return fmt.Errorf("failed to create role binding: %v", err)
	}

	if exists {
		s.log.Infof("Role Binding '%s' already exists, skipping", binding.Name())
		return nil
	}

	s.log.Infof("Role Binding '%s' Created", binding.Name())

	s.roleBindings[binding.Name()] = binding

	if err := s.updateBindingList(binding); err != nil {
		return err
	}

	return nil
}

func (s *SubjectAccessDelegation) deleteRoleBinding(binding interfaces.Binding) error {
	if binding == nil {
		return nil
	}

	if b, ok := s.roleBindings[binding.Name()]; !ok || b == nil {
		s.roleBindings[binding.Name()] = nil

		return fmt.Errorf("failed to find binding '%s' stored locally", binding.Name())
	}

	if err := s.deleteBindingList(binding); err != nil {
		return err
	}

	if err := binding.DeleteRoleBinding(); err != nil {
		return err
	}

	s.log.Infof("Role Binding '%s' Deleted", binding.Name())

	return nil
}

func (s *SubjectAccessDelegation) cleanUpBindings() error {
	var result *multierror.Error

	if err := s.deleteAllBindings(); err != nil {
		result = multierror.Append(result, err)
	}

	if err := s.updateRemoteSAD(); err != nil {
		result = multierror.Append(result, fmt.Errorf("failed to update trigger status against API server: %v", err))
	}

	return result.ErrorOrNil()
}

func (s *SubjectAccessDelegation) deleteAllBindings() error {
	var result *multierror.Error

	options := &metav1.DeleteOptions{}

	for _, binding := range s.sad.Status.RoleBindings {
		if err := s.client.Rbac().RoleBindings(s.Namespace()).Delete(binding, options); err != nil {
			result = multierror.Append(result, err)
		}
	}

	for _, binding := range s.sad.Status.ClusterRoleBindings {
		if err := s.client.Rbac().ClusterRoleBindings().Delete(binding, options); err != nil {
			result = multierror.Append(result, err)
		}
	}

	s.sad.Status.RoleBindings = make([]string, 0)
	s.sad.Status.ClusterRoleBindings = make([]string, 0)

	return result.ErrorOrNil()
}

func (s *SubjectAccessDelegation) updateInteration(iteration int) error {
	if err := s.updateLocalSAD(); err != nil {
		return err
	}

	s.sad.Status.Iteration = iteration

	if err := s.updateRemoteSAD(); err != nil {
		return fmt.Errorf("failed to update iteration status against API server: %v", err)
	}

	return nil
}

func (s *SubjectAccessDelegation) updateBindingList(binding interfaces.Binding) error {
	if err := s.updateLocalSAD(); err != nil {
		return err
	}

	if binding.Kind() == role_binding.RoleBindingKind {

		deepCopy := s.sad.DeepCopy().Status.RoleBindings
		deepCopy = append(deepCopy, binding.Name())
		s.sad.Status.RoleBindings = deepCopy

	} else {

		deepCopy := s.sad.DeepCopy().Status.ClusterRoleBindings
		deepCopy = append(deepCopy, binding.Name())
		s.sad.Status.ClusterRoleBindings = deepCopy

	}

	if err := s.updateRemoteSAD(); err != nil {
		return fmt.Errorf("failed to update binding list against API server: %v", err)
	}

	return nil
}

func (s *SubjectAccessDelegation) deleteBindingList(binding interfaces.Binding) error {
	if err := s.updateLocalSAD(); err != nil {
		return err
	}

	found := false

	if binding.Kind() == role_binding.RoleBindingKind {

		deepCopy := s.sad.DeepCopy().Status.RoleBindings

		for i, name := range deepCopy {
			if binding.Name() == name {

				deepCopy[i] = deepCopy[len(deepCopy)-1]
				deepCopy = deepCopy[:len(deepCopy)-1]
				s.sad.Status.RoleBindings = deepCopy
				found = true

				break
			}
		}

	} else {

		deepCopy := s.sad.DeepCopy().Status.ClusterRoleBindings

		for i, name := range deepCopy {
			if binding.Name() == name {

				deepCopy[i] = deepCopy[len(deepCopy)-1]
				deepCopy = deepCopy[:len(deepCopy)-1]
				s.sad.Status.ClusterRoleBindings = deepCopy
				found = true

				break
			}
		}
	}

	if !found {
		return fmt.Errorf("failed to find binding '%s' in SAD API object", binding.Name())
	}

	if err := s.updateRemoteSAD(); err != nil {
		return fmt.Errorf("failed to update binding list against API server: %v", err)
	}

	return nil
}

func (s *SubjectAccessDelegation) updateLocalSAD() error {
	options := metav1.GetOptions{}
	sad, err := s.sadclientset.Authz().SubjectAccessDelegations(s.Namespace()).Get(s.sad.Name, options)
	if err != nil {
		return fmt.Errorf("failed to get latest SAD from API server: %v", err)
	}

	s.sad = sad

	return nil
}

func (s *SubjectAccessDelegation) updateRemoteSAD() error {
	options := metav1.GetOptions{}

	retryErr := retry.RetryOnConflict(retry.DefaultRetry, func() error {
		sad, err := s.sadclientset.Authz().SubjectAccessDelegations(s.Namespace()).Get(s.sad.Name, options)
		if err != nil {
			return fmt.Errorf("failed to get latest sad object for updating: %v", err)
		}

		sad.Spec = s.sad.Spec
		sad.Status = s.sad.Status

		_, err = s.sadclientset.Authz().SubjectAccessDelegations(s.Namespace()).Update(sad)
		if err != nil {
			return fmt.Errorf("failed to update SAD in API server: %v", err)
		}

		s.sad = sad

		return nil
	})

	if retryErr != nil {
		return fmt.Errorf("failed to update SAD: %v", retryErr)
	}

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
		subjects, err := destination_subject.New(s, destinationSubject.Name, destinationSubject.Kind)

		if err != nil {
			result = multierror.Append(result, err)
		} else {
			destinationSubjects = append(destinationSubjects, subjects...)
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
	var result *multierror.Error

	if !s.sad.Status.Processed {
		return nil
	}

	s.log.Debugf("Attempting to delete delegation '%s' bindings", s.Name())
	if err := s.deleteAllBindings(); err != nil {
		result = multierror.Append(result, err)
	}

	s.log.Debugf("Attempting to delete delegation '%s' Origin Subject", s.Name())
	if s.OriginSubject() != nil {
		s.OriginSubject().Delete()
	}

	s.log.Debugf("Attempting to delete delegation '%s' triggers", s.Name())
	for _, trigger := range s.triggers {
		if err := trigger.Delete(); err != nil {
			result = multierror.Append(result, err)
		}
	}

	for _, trigger := range s.deletionTriggers {
		if err := trigger.Delete(); err != nil {
			result = multierror.Append(result, err)
		}
	}

	close(s.stopCh)

	return result.ErrorOrNil()
}

func (s *SubjectAccessDelegation) AddRoleBinding(addBinding interfaces.Binding) error {
	if s.triggered {
		if err := s.createRoleBinding(addBinding); err != nil {
			return err
		}
	}

	return nil
}

func (s *SubjectAccessDelegation) DeleteRoleBinding(delBinding interfaces.Binding) error {
	if s.triggered {

		if err := s.deleteRoleBinding(delBinding); err != nil {
			return fmt.Errorf("error deleting binding: %v", err)
		}

	}

	return nil
}

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

func (s *SubjectAccessDelegation) TimeActivated() int64 {
	return s.sad.Status.TimeActivated
}

func (s *SubjectAccessDelegation) TimeFired() int64 {
	return s.sad.Status.TimeFired
}

func (s *SubjectAccessDelegation) originKind() string {
	return s.sad.Spec.OriginSubject.Kind
}

func (s *SubjectAccessDelegation) OriginSubject() interfaces.OriginSubject {
	return s.originSubject
}

func (s *SubjectAccessDelegation) Triggers() []interfaces.Trigger {
	return s.triggers
}

func (s *SubjectAccessDelegation) DestinationSubjects() []interfaces.DestinationSubject {
	return s.destinationSubjects
}

func (s *SubjectAccessDelegation) Repeat() int {
	return s.sad.Spec.Repeat
}

func (s *SubjectAccessDelegation) KubeInformerFactory() kubeinformers.SharedInformerFactory {
	return kubeinformers.NewSharedInformerFactory(s.client, time.Second*30)
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

func (s *SubjectAccessDelegation) SAD() *authzv1alpha1.SubjectAccessDelegation {
	return s.sad
}
