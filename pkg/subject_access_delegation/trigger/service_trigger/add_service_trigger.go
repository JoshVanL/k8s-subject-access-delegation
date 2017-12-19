package service_trigger

import (
	"github.com/sirupsen/logrus"
	informer "k8s.io/client-go/informers/core/v1"
	"k8s.io/client-go/tools/cache"

	authzv1alpha1 "github.com/joshvanl/k8s-subject-access-delegation/pkg/apis/authz/v1alpha1"
	"github.com/joshvanl/k8s-subject-access-delegation/pkg/subject_access_delegation/interfaces"
	"github.com/joshvanl/k8s-subject-access-delegation/pkg/subject_access_delegation/utils"
)

type AddServiceTrigger struct {
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

var _ interfaces.Trigger = &AddServiceTrigger{}

func NewAddServiceTrigger(sad interfaces.SubjectAccessDelegation, trigger *authzv1alpha1.EventTrigger) (serviceTrigger *AddServiceTrigger, err error) {
	serviceTrigger = &AddServiceTrigger{
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

func (p *AddServiceTrigger) addFunc(obj interface{}) {

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

func (p *AddServiceTrigger) WaitOn() (forceClosed bool) {
	p.log.Debug("Trigger waiting")

	if p.watchChannels() {
		p.log.Debug("Add Service Trigger was force closed")
		return true
	}

	p.log.Debug("Add Service Trigger completed")
	return false
}

func (p *AddServiceTrigger) watchChannels() (forceClose bool) {
	select {
	case <-p.stopCh:
		return true
	case <-p.completedCh:
		return false
	}
}

func (p *AddServiceTrigger) Activate() {
	p.log.Debug("Add Service Trigger Activated")

	go p.informer.Informer().Run(make(chan struct{}))

	return
}

func (p *AddServiceTrigger) Completed() bool {
	return p.completed
}

func (p *AddServiceTrigger) Delete() error {
	close(p.stopCh)
	return nil
}

func (p *AddServiceTrigger) Replicas() int {
	return p.replicas
}
