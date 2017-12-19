package node_trigger

import (
	"github.com/sirupsen/logrus"
	informer "k8s.io/client-go/informers/core/v1"
	"k8s.io/client-go/tools/cache"

	authzv1alpha1 "github.com/joshvanl/k8s-subject-access-delegation/pkg/apis/authz/v1alpha1"
	"github.com/joshvanl/k8s-subject-access-delegation/pkg/subject_access_delegation/interfaces"
	"github.com/joshvanl/k8s-subject-access-delegation/pkg/subject_access_delegation/utils"
)

type DelNodeTrigger struct {
	log *logrus.Entry

	sad      interfaces.SubjectAccessDelegation
	nodeName string
	replicas int

	stopCh      chan struct{}
	completedCh chan struct{}

	count     int
	completed bool
	informer  informer.NodeInformer
}

var _ interfaces.Trigger = &DelNodeTrigger{}

func NewDelNodeTrigger(sad interfaces.SubjectAccessDelegation, trigger *authzv1alpha1.EventTrigger) (nodeTrigger *DelNodeTrigger, err error) {
	nodeTrigger = &DelNodeTrigger{
		log:         sad.Log(),
		sad:         sad,
		nodeName:    trigger.Value,
		replicas:    trigger.Replicas,
		stopCh:      make(chan struct{}),
		completedCh: make(chan struct{}),
		count:       0,
		completed:   false,
		informer:    sad.KubeInformerFactory().Core().V1().Nodes(),
	}

	nodeTrigger.informer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		DeleteFunc: nodeTrigger.delFunc,
	})

	return nodeTrigger, nil
}

func (p *DelNodeTrigger) delFunc(obj interface{}) {

	node, err := utils.GetNodeObject(obj)
	if err != nil {
		p.log.Errorf("failed to get deleted node object: %v", err)
		return
	}
	if node == nil {
		p.log.Error("failed to get node, received nil object")
	}

	if node.Name != p.nodeName {
		return
	}

	p.log.Infof("A node '%s' has been deleted", node.Name)
	p.count++
	if p.count >= p.replicas {
		p.log.Infof("Required replicas was met")
		p.completed = true
		close(p.completedCh)
	}
}

func (p *DelNodeTrigger) WaitOn() (forceClosed bool) {
	p.log.Debug("Trigger waiting")

	if p.watchChannels() {
		p.log.Debug("Del Node Trigger was force closed")
		return true
	}

	p.log.Debug("Del Node Trigger completed")
	return false
}

func (p *DelNodeTrigger) watchChannels() (forceClose bool) {
	select {
	case <-p.stopCh:
		return true
	case <-p.completedCh:
		return false
	}
}

func (p *DelNodeTrigger) Activate() {
	p.log.Debug("Del Node Trigger Activated")

	go p.informer.Informer().Run(make(chan struct{}))

	return
}

func (p *DelNodeTrigger) Completed() bool {
	return p.completed
}

func (p *DelNodeTrigger) Delete() error {
	close(p.stopCh)
	return nil
}

func (p *DelNodeTrigger) Replicas() int {
	return p.replicas
}
