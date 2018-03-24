package service_account

import (
	"fmt"

	"github.com/sirupsen/logrus"
	informer "k8s.io/client-go/informers/core/v1"
	"k8s.io/client-go/tools/cache"

	authzv1alpha1 "github.com/joshvanl/k8s-subject-access-delegation/pkg/apis/authz/v1alpha1"
	"github.com/joshvanl/k8s-subject-access-delegation/pkg/interfaces"
	"github.com/joshvanl/k8s-subject-access-delegation/pkg/subject_access_delegation/utils"
)

const UpdateServiceAccountKind = "UpdateServiceAccountKind"

type UpdateServiceAccount struct {
	log *logrus.Entry

	sad                interfaces.SubjectAccessDelegation
	serviceAccountName string
	replicas           int

	stopCh      chan struct{}
	completedCh chan struct{}

	count     int
	completed bool
	informer  informer.ServiceAccountInformer
}

var _ interfaces.Trigger = &UpdateServiceAccount{}

func NewUpdateServiceAccount(sad interfaces.SubjectAccessDelegation, trigger *authzv1alpha1.EventTrigger) (*UpdateServiceAccount, error) {

	if !utils.ValidName(trigger.Value) {
		return nil, fmt.Errorf("not a valid name '%s', must contain only alphanumerics, '-', '.' and '*'", trigger.Value)
	}

	serviceAccountTrigger := &UpdateServiceAccount{
		log:                sad.Log(),
		sad:                sad,
		serviceAccountName: trigger.Value,
		replicas:           trigger.Replicas,
		stopCh:             make(chan struct{}),
		completedCh:        make(chan struct{}),
		count:              0,
		completed:          false,
		informer:           sad.KubeInformerFactory().Core().V1().ServiceAccounts(),
	}

	serviceAccountTrigger.informer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		UpdateFunc: serviceAccountTrigger.updateFunc,
	})

	return serviceAccountTrigger, nil
}

func (s *UpdateServiceAccount) updateFunc(oldObj, newObj interface{}) {

	old, err := utils.GetServiceAccountObject(oldObj)
	if err != nil {
		s.log.Errorf("failed to get old updated serviceAccount object: %v", err)
		return
	}

	new, err := utils.GetServiceAccountObject(newObj)
	if err != nil {
		s.log.Errorf("failed to get new updated serviceAccount object: %v", err)
		return
	}
	if new == nil || old == nil {
		s.log.Error("failed to get serviceAccount, received nil object")
	}

	match, err := utils.MatchName(old.Name, s.serviceAccountName)
	if err != nil {
		s.log.Error("failed to match service account name: %v", err)
		return
	}

	if !match || s.sad.DeletedUid(old.UID) || s.completed {
		return
	}

	s.sad.AddUid(new.UID)

	s.log.Infof("A serviceAccount '%s' has been updated", new.Name)
	s.count++
	if s.count >= s.replicas {
		s.log.Infof("Required replicas was met")
		s.completed = true
		close(s.completedCh)
	}
}

func (s *UpdateServiceAccount) WaitOn() (forceClosed bool) {
	s.log.Debug("Trigger waiting")

	if s.watchChannels() {
		s.log.Debug("Update ServiceAccount Trigger was force closed")
		return true
	}

	s.log.Debug("Update ServiceAccount Trigger completed")
	return false
}

func (s *UpdateServiceAccount) watchChannels() (forceClose bool) {
	select {
	case <-s.stopCh:
		return true
	case <-s.completedCh:
		return false
	}
}

func (s *UpdateServiceAccount) Activate() {
	s.log.Debug("Update ServiceAccount Trigger Activated")

	go s.informer.Informer().Run(s.completedCh)

	return
}

func (s *UpdateServiceAccount) Completed() bool {
	return s.completed
}

func (s *UpdateServiceAccount) Delete() error {
	close(s.stopCh)
	return nil
}

func (s *UpdateServiceAccount) Replicas() int {
	return s.replicas
}

func (s *UpdateServiceAccount) Kind() string {
	return UpdateServiceAccountKind
}
