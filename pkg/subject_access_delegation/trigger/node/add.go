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

type AddNode struct {
	log *logrus.Entry

	sad      interfaces.SubjectAccessDelegation
	nodeName string
	replicas int
	uid      int

	stopCh      chan struct{}
	completedCh chan struct{}

	count     int
	completed bool
	informer  informer.NodeInformer
}

var _ interfaces.Trigger = &AddNode{}

const AddNodeKind = "AddNode"

func NewAddNode(sad interfaces.SubjectAccessDelegation, trigger *authzv1alpha1.EventTrigger) (*AddNode, error) {

	if !utils.ValidName(trigger.Value) {
		return nil, fmt.Errorf("not a valid name '%s', must contain only alphanumerics, '-', '.' and '*'", trigger.Value)
	}

	nodeTrigger := &AddNode{
		log:         sad.Log(),
		sad:         sad,
		nodeName:    trigger.Value,
		replicas:    trigger.Replicas,
		stopCh:      make(chan struct{}),
		completedCh: make(chan struct{}),
		count:       0,
		completed:   trigger.Triggered,
		uid:         trigger.UID,
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

	match, err := utils.MatchName(node.Name, p.nodeName)
	if err != nil {
		p.log.Error("failed to match node name: %v", err)
		return
	}

	if !match || p.sad.SeenUid(node.UID) || p.completed {
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

	if err := p.sad.UpdateTriggerFired(p.uid, true); err != nil {
		p.log.Errorf("error updating add node trigger status: %v", err)
	}

	return false
}

func (p *AddNode) watchChannels() (forceClose bool) {
	for {
		select {
		case <-p.stopCh:
			return true
		case <-p.completedCh:
			return false
		}
	}
}

func (p *AddNode) Activate() {
	p.log.Debug("Add Node Trigger Activated")
	p.completed = false

	go p.informer.Informer().Run(p.completedCh)

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

func (p *AddNode) Kind() string {
	return AddNodeKind
}
