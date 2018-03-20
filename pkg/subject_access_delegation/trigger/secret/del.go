package secret

import (
	"github.com/sirupsen/logrus"
	informer "k8s.io/client-go/informers/core/v1"
	"k8s.io/client-go/tools/cache"

	authzv1alpha1 "github.com/joshvanl/k8s-subject-access-delegation/pkg/apis/authz/v1alpha1"
	"github.com/joshvanl/k8s-subject-access-delegation/pkg/interfaces"
	"github.com/joshvanl/k8s-subject-access-delegation/pkg/subject_access_delegation/utils"
)

const DelSecretKind = "DelSecret"

type DelSecret struct {
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

var _ interfaces.Trigger = &DelSecret{}

func NewDelSecret(sad interfaces.SubjectAccessDelegation, trigger *authzv1alpha1.EventTrigger) (*DelSecret, error) {
	secretTrigger := &DelSecret{
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
		DeleteFunc: secretTrigger.delFunc,
	})

	return secretTrigger, nil
}

func (s *DelSecret) delFunc(obj interface{}) {

	secret, err := utils.GetSecretObject(obj)
	if err != nil {
		s.log.Errorf("failed to get deleted secret object: %v", err)
		return
	}
	if secret == nil {
		s.log.Error("failed to get secret, received nil object")
	}

	if secret.Name != s.secretName || s.sad.DeletedUid(secret.UID) {
		return
	}

	s.sad.DeleteUid(secret.UID)

	s.log.Infof("A secret '%s' has been deleted", secret.Name)
	s.count++
	if s.count >= s.replicas {
		s.log.Infof("Required replicas was met")
		s.completed = true
		close(s.completedCh)
	}
}

func (s *DelSecret) WaitOn() (forceClosed bool) {
	s.log.Debug("Trigger waiting")

	if s.watchChannels() {
		s.log.Debug("Del Secret Trigger was force closed")
		return true
	}

	s.log.Debug("Del Secret Trigger completed")
	return false
}

func (s *DelSecret) watchChannels() (forceClose bool) {
	select {
	case <-s.stopCh:
		return true
	case <-s.completedCh:
		return false
	}
}

func (s *DelSecret) Activate() {
	s.log.Debug("Del Secret Trigger Activated")

	go s.informer.Informer().Run(make(chan struct{}))

	return
}

func (s *DelSecret) Completed() bool {
	return s.completed
}

func (s *DelSecret) Delete() error {
	close(s.stopCh)
	return nil
}

func (s *DelSecret) Replicas() int {
	return s.replicas
}

func (s *DelSecret) Kind() string {
	return DelSecretKind
}