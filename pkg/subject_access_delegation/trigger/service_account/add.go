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

const AddServiceAccountKind = "AddServiceAccount"

type AddServiceAccount struct {
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

var _ interfaces.Trigger = &AddServiceAccount{}

func NewAddServiceAccount(sad interfaces.SubjectAccessDelegation, trigger *authzv1alpha1.EventTrigger) (*AddServiceAccount, error) {

	if !utils.ValidName(trigger.Value) {
		return nil, fmt.Errorf("not a valid name '%s', must contain only alphanumerics, '-', '.' and '*'", trigger.Value)
	}

	serviceAccountTrigger := &AddServiceAccount{
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
		AddFunc: serviceAccountTrigger.addFunc,
	})

	return serviceAccountTrigger, nil
}

func (s *AddServiceAccount) addFunc(obj interface{}) {

	serviceAccount, err := utils.GetServiceAccountObject(obj)
	if err != nil {
		s.log.Errorf("failed to get added serviceAccount object: %v", err)
		return
	}
	if serviceAccount == nil {
		s.log.Error("failed to get serviceAccount, received nil object")
	}

	match, err := utils.MatchName(serviceAccount.Name, s.serviceAccountName)
	if err != nil {
		s.log.Error("failed to match service account name: %v", err)
		return
	}

	if !match || s.sad.SeenUid(serviceAccount.UID) || s.completed {
		return
	}

	s.sad.AddUid(serviceAccount.UID)

	s.log.Infof("A new serviceAccount '%s' has been added", serviceAccount.Name)
	s.count++
	if s.count >= s.replicas {
		s.log.Infof("Required replicas was met")
		s.completed = true
		close(s.completedCh)
	}
}

func (s *AddServiceAccount) WaitOn() (forceClosed bool) {
	s.log.Debug("Trigger waiting")

	if s.watchChannels() {
		s.log.Debug("Add ServiceAccount Trigger was force closed")
		return true
	}

	s.log.Debug("Add ServiceAccount Trigger completed")
	return false
}

func (s *AddServiceAccount) watchChannels() (forceClose bool) {
	select {
	case <-s.stopCh:
		return true
	case <-s.completedCh:
		return false
	}
}

func (s *AddServiceAccount) Activate() {
	s.log.Debug("Add ServiceAccount Trigger Activated")

	go s.informer.Informer().Run(s.completedCh)

	return
}

func (s *AddServiceAccount) Completed() bool {
	return s.completed
}

func (s *AddServiceAccount) Delete() error {
	close(s.stopCh)
	return nil
}

func (s *AddServiceAccount) Replicas() int {
	return s.replicas
}

func (s *AddServiceAccount) Kind() string {
	return AddServiceAccountKind
}
