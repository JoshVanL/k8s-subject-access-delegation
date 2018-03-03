package service

import (
	"github.com/sirupsen/logrus"
	informer "k8s.io/client-go/informers/core/v1"
	"k8s.io/client-go/tools/cache"

	authzv1alpha1 "github.com/joshvanl/k8s-subject-access-delegation/pkg/apis/authz/v1alpha1"
	"github.com/joshvanl/k8s-subject-access-delegation/pkg/interfaces"
	"github.com/joshvanl/k8s-subject-access-delegation/pkg/subject_access_delegation/utils"
)

type DelService struct {
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

var _ interfaces.Trigger = &DelService{}

func NewDelService(sad interfaces.SubjectAccessDelegation, trigger *authzv1alpha1.EventTrigger) (*DelService, error) {
	serviceTrigger := &DelService{
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
		DeleteFunc: serviceTrigger.delFunc,
	})

	return serviceTrigger, nil
}

func (p *DelService) delFunc(obj interface{}) {

	service, err := utils.GetServiceObject(obj)
	if err != nil {
		p.log.Errorf("failed to get deleted service object: %v", err)
		return
	}
	if service == nil {
		p.log.Error("failed to get service, received nil object")
	}

	if service.Name != p.serviceName || p.sad.DeletedUid(service.UID) {
		return
	}

	p.sad.AddUid(service.UID)

	p.log.Infof("A service '%s' has been deleted", service.Name)
	p.count++
	if p.count >= p.replicas {
		p.log.Infof("Required replicas was met")
		p.completed = true
		close(p.completedCh)
	}
}

func (p *DelService) WaitOn() (forceClosed bool) {
	p.log.Debug("Trigger waiting")

	if p.watchChannels() {
		p.log.Debug("Del Service Trigger was force closed")
		return true
	}

	p.log.Debug("Del Service Trigger completed")
	return false
}

func (p *DelService) watchChannels() (forceClose bool) {
	select {
	case <-p.stopCh:
		return true
	case <-p.completedCh:
		return false
	}
}

func (p *DelService) Activate() {
	p.log.Debug("Del Service Trigger Activated")

	go p.informer.Informer().Run(make(chan struct{}))

	return
}

func (p *DelService) Completed() bool {
	return p.completed
}

func (p *DelService) Delete() error {
	close(p.stopCh)
	return nil
}

func (p *DelService) Replicas() int {
	return p.replicas
}
