package secret

import (
	"fmt"

	"github.com/sirupsen/logrus"
	informer "k8s.io/client-go/informers/core/v1"
	"k8s.io/client-go/tools/cache"

	authzv1alpha1 "github.com/joshvanl/k8s-subject-access-delegation/pkg/apis/authz/v1alpha1"
	"github.com/joshvanl/k8s-subject-access-delegation/pkg/interfaces"
	"github.com/joshvanl/k8s-subject-access-delegation/pkg/subject_access_delegation/utils"
)

const UpdateSecretKind = "UpdateSecretKind"

type UpdateSecret struct {
	log *logrus.Entry

	sad        interfaces.SubjectAccessDelegation
	secretName string
	replicas   int

	stopCh      chan struct{}
	completedCh chan struct{}

	count     int
	completed bool
	informer  informer.SecretInformer
}

var _ interfaces.Trigger = &UpdateSecret{}

func NewUpdateSecret(sad interfaces.SubjectAccessDelegation, trigger *authzv1alpha1.EventTrigger) (*UpdateSecret, error) {

	if !utils.ValidName(trigger.Value) {
		return nil, fmt.Errorf("not a valid name '%s', must contain only alphanumerics, '-', '.' and '*'", trigger.Value)
	}

	secretTrigger := &UpdateSecret{
		log:         sad.Log(),
		sad:         sad,
		secretName:  trigger.Value,
		replicas:    trigger.Replicas,
		stopCh:      make(chan struct{}),
		completedCh: make(chan struct{}),
		count:       0,
		completed:   false,
		informer:    sad.KubeInformerFactory().Core().V1().Secrets(),
	}

	secretTrigger.informer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		UpdateFunc: secretTrigger.updateFunc,
	})

	return secretTrigger, nil
}

func (s *UpdateSecret) updateFunc(oldObj, newObj interface{}) {

	old, err := utils.GetSecretObject(oldObj)
	if err != nil {
		s.log.Errorf("failed to get old updated secret object: %v", err)
		return
	}

	new, err := utils.GetSecretObject(newObj)
	if err != nil {
		s.log.Errorf("failed to get new updated secret object: %v", err)
		return
	}
	if new == nil || old == nil {
		s.log.Error("failed to get secret, received nil object")
	}

	match, err := utils.MatchName(old.Name, s.secretName)
	if err != nil {
		s.log.Error("failed to match secret name: %v", err)
		return
	}

	if !match || s.sad.DeletedUid(old.UID) {
		return
	}

	s.sad.AddUid(new.UID)

	s.log.Infof("A secret '%s' has been updated", new.Name)
	s.count++
	if s.count >= s.replicas {
		s.log.Infof("Required replicas was met")
		s.completed = true
		close(s.completedCh)
	}
}

func (s *UpdateSecret) WaitOn() (forceClosed bool) {
	s.log.Debug("Trigger waiting")

	if s.watchChannels() {
		s.log.Debug("Update Secret Trigger was force closed")
		return true
	}

	s.log.Debug("Update Secret Trigger completed")
	return false
}

func (s *UpdateSecret) watchChannels() (forceClose bool) {
	select {
	case <-s.stopCh:
		return true
	case <-s.completedCh:
		return false
	}
}

func (s *UpdateSecret) Activate() {
	s.log.Debug("Update Secret Trigger Activated")

	go s.informer.Informer().Run(make(chan struct{}))

	return
}

func (s *UpdateSecret) Completed() bool {
	return s.completed
}

func (s *UpdateSecret) Delete() error {
	close(s.stopCh)
	return nil
}

func (s *UpdateSecret) Replicas() int {
	return s.replicas
}

func (s *UpdateSecret) Kind() string {
	return UpdateSecretKind
}
