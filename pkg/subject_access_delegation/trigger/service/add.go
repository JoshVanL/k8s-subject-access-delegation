package service

import (
	"github.com/sirupsen/logrus"
	informer "k8s.io/client-go/informers/core/v1"
	"k8s.io/client-go/tools/cache"

	authzv1alpha1 "github.com/joshvanl/k8s-subject-access-delegation/pkg/apis/authz/v1alpha1"
	"github.com/joshvanl/k8s-subject-access-delegation/pkg/interfaces"
	"github.com/joshvanl/k8s-subject-access-delegation/pkg/subject_access_delegation/utils"
)

type AddService struct {
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

var _ interfaces.Trigger = &AddService{}

func NewAddService(sad interfaces.SubjectAccessDelegation, trigger *authzv1alpha1.EventTrigger) (serviceTrigger *AddService, err error) {
	serviceTrigger = &AddService{
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
		//TODO: Fix this
		AddFunc: serviceTrigger.addFunc,
	})

	return serviceTrigger, nil
}

func (p *AddService) addFunc(obj interface{}) {

	service, err := utils.GetServiceObject(obj)
	if err != nil {
		p.log.Errorf("failed to get added service object: %v", err)
		return
	}
	if service == nil {
		p.log.Error("failed to get service, received nil object")
	}

	if service.Name != p.serviceName {
		return
	}

	p.log.Infof("A new service '%s' has been added", service.Name)
	p.count++
	if p.count >= p.replicas {
		p.log.Infof("Required replicas was met")
		p.completed = true
		close(p.completedCh)
	}
}

func (p *AddService) WaitOn() (forceClosed bool) {
	p.log.Debug("Trigger waiting")

	if p.watchChannels() {
		p.log.Debug("Add Service Trigger was force closed")
		return true
	}

	p.log.Debug("Add Service Trigger completed")
	return false
}

func (p *AddService) watchChannels() (forceClose bool) {
	select {
	case <-p.stopCh:
		return true
	case <-p.completedCh:
		return false
	}
}

func (p *AddService) Activate() {
	p.log.Debug("Add Service Trigger Activated")

	go p.informer.Informer().Run(make(chan struct{}))

	return
}

func (p *AddService) Completed() bool {
	return p.completed
}

func (p *AddService) Delete() error {
	close(p.stopCh)
	return nil
}

func (p *AddService) Replicas() int {
	return p.replicas
}
