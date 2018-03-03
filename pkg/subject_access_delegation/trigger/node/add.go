package node

import (
	"github.com/sirupsen/logrus"
	informer "k8s.io/client-go/informers/core/v1"
	"k8s.io/client-go/tools/cache"

	authzv1alpha1 "github.com/joshvanl/k8s-subject-access-delegation/pkg/apis/authz/v1alpha1"
	"github.com/joshvanl/k8s-subject-access-delegation/pkg/interfaces"
	"github.com/joshvanl/k8s-subject-access-delegation/pkg/subject_access_delegation/utils"
)

type AddNode struct {
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

var _ interfaces.Trigger = &AddNode{}

func NewAddNode(sad interfaces.SubjectAccessDelegation, trigger *authzv1alpha1.EventTrigger) (*AddNode, error) {
	nodeTrigger := &AddNode{
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
		//TODO: Fix this
		AddFunc: nodeTrigger.addFunc,
	})

	return nodeTrigger, nil
}

func (p *AddNode) addFunc(obj interface{}) {

	node, err := utils.GetNodeObject(obj)
	if err != nil {
		p.log.Errorf("failed to get added node object: %v", err)
		return
	}
	if node == nil {
		p.log.Error("failed to get node, received nil object")
	}

	if node.Name != p.nodeName || p.sad.SeenUid(node.UID) {
		return
	}

	p.sad.AddUid(node.UID)

	p.log.Infof("A new node '%s' has been added", node.Name)
	p.count++
	if p.count >= p.replicas {
		p.log.Infof("Required replicas was met")
		p.completed = true
		close(p.completedCh)
	}
}

func (p *AddNode) WaitOn() (forceClosed bool) {
	p.log.Debug("Trigger waiting")

	if p.watchChannels() {
		p.log.Debug("Add Node Trigger was force closed")
		return true
	}

	p.log.Debug("Add Node Trigger completed")
	return false
}

func (p *AddNode) watchChannels() (forceClose bool) {
	select {
	case <-p.stopCh:
		return true
	case <-p.completedCh:
		return false
	}
}

func (p *AddNode) Activate() {
	p.log.Debug("Add Node Trigger Activated")

	go p.informer.Informer().Run(make(chan struct{}))

	return
}

func (p *AddNode) Completed() bool {
	return p.completed
}

func (p *AddNode) Delete() error {
	close(p.stopCh)
	return nil
}

func (p *AddNode) Replicas() int {
	return p.replicas
}
