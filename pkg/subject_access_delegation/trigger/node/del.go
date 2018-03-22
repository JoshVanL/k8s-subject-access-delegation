package node

import (
	"fmt"

	"github.com/sirupsen/logrus"
	informer "k8s.io/client-go/informers/core/v1"
	"k8s.io/client-go/tools/cache"

	authzv1alpha1 "github.com/joshvanl/k8s-subject-access-delegation/pkg/apis/authz/v1alpha1"
	"github.com/joshvanl/k8s-subject-access-delegation/pkg/interfaces"
	"github.com/joshvanl/k8s-subject-access-delegation/pkg/subject_access_delegation/utils"
)

const DelNodeKind = "DelNode"

type DelNode struct {
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

var _ interfaces.Trigger = &DelNode{}

func NewDelNode(sad interfaces.SubjectAccessDelegation, trigger *authzv1alpha1.EventTrigger) (*DelNode, error) {

	if !utils.ValidName(trigger.Value) {
		return nil, fmt.Errorf("not a valid name '%s', must contain only alphanumerics, '-', '.' and '*'", trigger.Value)
	}

	nodeTrigger := &DelNode{
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

func (p *DelNode) delFunc(obj interface{}) {

	node, err := utils.GetNodeObject(obj)
	if err != nil {
		p.log.Errorf("failed to get deleted node object: %v", err)
		return
	}
	if node == nil {
		p.log.Error("failed to get node, received nil object")
	}

	match, err := utils.MatchName(node.Name, p.nodeName)
	if err != nil {
		p.log.Error("failed to match node name: %v", err)
		return
	}

	if !match || p.sad.DeletedUid(node.UID) {
		return
	}

	p.sad.DeleteUid(node.UID)

	p.log.Infof("A node '%s' has been deleted", node.Name)
	p.count++
	if p.count >= p.replicas {
		p.log.Infof("Required replicas was met")
		p.completed = true
		close(p.completedCh)
	}
}

func (p *DelNode) WaitOn() (forceClosed bool) {
	p.log.Debug("Trigger waiting")

	if p.watchChannels() {
		p.log.Debug("Del Node Trigger was force closed")
		return true
	}

	p.log.Debug("Del Node Trigger completed")
	return false
}

func (p *DelNode) watchChannels() (forceClose bool) {
	select {
	case <-p.stopCh:
		return true
	case <-p.completedCh:
		return false
	}
}

func (p *DelNode) Activate() {
	p.log.Debug("Del Node Trigger Activated")

	go p.informer.Informer().Run(make(chan struct{}))

	return
}

func (p *DelNode) Completed() bool {
	return p.completed
}

func (p *DelNode) Delete() error {
	close(p.stopCh)
	return nil
}

func (p *DelNode) Replicas() int {
	return p.replicas
}

func (p *DelNode) Kind() string {
	return DelNodeKind
}
