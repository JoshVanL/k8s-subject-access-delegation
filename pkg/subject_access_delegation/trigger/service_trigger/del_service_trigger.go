package service_trigger

import (
	"github.com/sirupsen/logrus"
	informer "k8s.io/client-go/informers/core/v1"
	"k8s.io/client-go/tools/cache"

	authzv1alpha1 "github.com/joshvanl/k8s-subject-access-delegation/pkg/apis/authz/v1alpha1"
	"github.com/joshvanl/k8s-subject-access-delegation/pkg/subject_access_delegation/interfaces"
	"github.com/joshvanl/k8s-subject-access-delegation/pkg/subject_access_delegation/utils"
)

type DelServiceTrigger struct {
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

var _ interfaces.Trigger = &DelServiceTrigger{}

func NewDelServiceTrigger(sad interfaces.SubjectAccessDelegation, trigger *authzv1alpha1.EventTrigger) (serviceTrigger *DelServiceTrigger, err error) {
	serviceTrigger = &DelServiceTrigger{
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

func (p *DelServiceTrigger) delFunc(obj interface{}) {

	service, err := utils.GetServiceObject(obj)
	if err != nil {
		p.log.Errorf("failed to get deleted service object: %v", err)
		return
	}
	if service == nil {
		p.log.Error("failed to get service, received nil object")
	}

	if service.Name != p.serviceName {
		return
	}

	p.log.Infof("A service '%s' has been deleted", service.Name)
	p.count++
	if p.count >= p.replicas {
		p.log.Infof("Required replicas was met")
		p.completed = true
		close(p.completedCh)
	}
}

func (p *DelServiceTrigger) WaitOn() (forceClosed bool) {
	p.log.Debug("Trigger waiting")

	if p.watchChannels() {
		p.log.Debug("Del Service Trigger was force closed")
		return true
	}

	p.log.Debug("Del Service Trigger completed")
	return false
}

func (p *DelServiceTrigger) watchChannels() (forceClose bool) {
	select {
	case <-p.stopCh:
		return true
	case <-p.completedCh:
		return false
	}
}

func (p *DelServiceTrigger) Activate() {
	p.log.Debug("Del Service Trigger Activated")

	go p.informer.Informer().Run(make(chan struct{}))

	return
}

func (p *DelServiceTrigger) Completed() bool {
	return p.completed
}

func (p *DelServiceTrigger) Delete() error {
	close(p.stopCh)
	return nil
}

func (p *DelServiceTrigger) Replicas() int {
	return p.replicas
}
