package service

import (
	"fmt"

	"github.com/sirupsen/logrus"
	informer "k8s.io/client-go/informers/core/v1"
	"k8s.io/client-go/tools/cache"

	authzv1alpha1 "github.com/joshvanl/k8s-subject-access-delegation/pkg/apis/authz/v1alpha1"
	"github.com/joshvanl/k8s-subject-access-delegation/pkg/interfaces"
	"github.com/joshvanl/k8s-subject-access-delegation/pkg/subject_access_delegation/utils"
)

const UpdateServiceKind = "UpdateService"

type UpdateService struct {
	log *logrus.Entry

	sad         interfaces.SubjectAccessDelegation
	serviceName string
	replicas    int

	stopCh      chan struct{}
	completedCh chan struct{}

	count     int
	completed bool
	informer  informer.ServiceInformer
}

var _ interfaces.Trigger = &UpdateService{}

func NewUpdateService(sad interfaces.SubjectAccessDelegation, trigger *authzv1alpha1.EventTrigger) (*UpdateService, error) {

	if !utils.ValidName(trigger.Value) {
		return nil, fmt.Errorf("not a valid name '%s', must contain only alphanumerics, '-', '.' and '*'", trigger.Value)
	}

	serviceTrigger := &UpdateService{
		log:         sad.Log(),
		sad:         sad,
		serviceName: trigger.Value,
		replicas:    trigger.Replicas,
		stopCh:      make(chan struct{}),
		completedCh: make(chan struct{}),
		count:       0,
		completed:   false,
		informer:    sad.KubeInformerFactory().Core().V1().Services(),
	}

	serviceTrigger.informer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		UpdateFunc: serviceTrigger.updateFunc,
	})

	return serviceTrigger, nil
}

func (s *UpdateService) updateFunc(oldObj, newObj interface{}) {
	old, err := utils.GetServiceObject(oldObj)
	if err != nil {
		s.log.Errorf("failed to get old service object: %v", err)
		return
	}

	new, err := utils.GetServiceObject(newObj)
	if err != nil {
		s.log.Errorf("failed to get updated service object: %v", err)
		return
	}

	if new == nil || old == nil {
		s.log.Error("failed to get service, received nil object")
	}

	match, err := utils.MatchName(old.Name, s.serviceName)
	if err != nil {
		s.log.Error("failed to match service name: %v", err)
		return
	}

	if !match || s.sad.DeletedUid(old.UID) {
		return
	}

	s.sad.AddUid(new.UID)

	s.log.Infof("A service '%s' has been updated", new.Name)
	s.count++
	if s.count >= s.replicas {
		s.log.Infof("Required replicas was met")
		s.completed = true
		close(s.completedCh)
	}
}

func (s *UpdateService) WaitOn() (forceClosed bool) {
	s.log.Debug("Trigger waiting")

	if s.watchChannels() {
		s.log.Debug("Update Service Trigger was force closed")
		return true
	}

	s.log.Debug("Update Service Trigger completed")
	return false
}

func (s *UpdateService) watchChannels() (forceClose bool) {
	select {
	case <-s.stopCh:
		return true
	case <-s.completedCh:
		return false
	}
}

func (s *UpdateService) Activate() {
	s.log.Debug("Update Service Trigger Activated")

	go s.informer.Informer().Run(make(chan struct{}))

	return
}

func (s *UpdateService) Completed() bool {
	return s.completed
}

func (s *UpdateService) Delete() error {
	close(s.stopCh)
	return nil
}

func (s *UpdateService) Replicas() int {
	return s.replicas
}

func (s *UpdateService) Kind() string {
	return UpdateServiceKind
}
