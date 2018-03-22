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

const UpdateNodeKind = "UpdateNode"

type UpdateNode struct {
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

var _ interfaces.Trigger = &UpdateNode{}

func NewUpdateNode(sad interfaces.SubjectAccessDelegation, trigger *authzv1alpha1.EventTrigger) (*UpdateNode, error) {

	if !utils.ValidName(trigger.Value) {
		return nil, fmt.Errorf("not a valid name '%s', must contain only alphanumerics, '-', '.' and '*'", trigger.Value)
	}

	nodeTrigger := &UpdateNode{
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
		UpdateFunc: nodeTrigger.updateFunc,
	})

	return nodeTrigger, nil
}

func (n *UpdateNode) updateFunc(oldObj, newObj interface{}) {

	old, err := utils.GetNodeObject(oldObj)
	if err != nil {
		n.log.Errorf("failed to get updated node object: %v", err)
		return
	}

	new, err := utils.GetNodeObject(newObj)
	if err != nil {
		n.log.Errorf("failed to get updated node object: %v", err)
		return
	}
	if new == nil || old == nil {
		n.log.Error("failed to get node, received nil object")
	}

	match, err := utils.MatchName(old.Name, n.nodeName)
	if err != nil {
		n.log.Error("failed to match node name: %v", err)
		return
	}

	if !match || n.sad.DeletedUid(new.UID) || n.completed {
		return
	}

	n.sad.AddUid(new.UID)

	n.log.Infof("A node '%s' has been updated", new.Name)
	n.count++
	if n.count >= n.replicas {
		n.log.Infof("Required replicas was met")
		n.completed = true
		close(n.completedCh)
	}
}

func (n *UpdateNode) WaitOn() (forceClosed bool) {
	n.log.Debug("Trigger waiting")

	if n.watchChannels() {
		n.log.Debug("Del Node Trigger was force closed")
		return true
	}

	n.log.Debug("Del Node Trigger completed")
	return false
}

func (n *UpdateNode) watchChannels() (forceClose bool) {
	select {
	case <-n.stopCh:
		return true
	case <-n.completedCh:
		return false
	}
}

func (n *UpdateNode) Activate() {
	n.log.Debug("Del Node Trigger Activated")

	go n.informer.Informer().Run(make(chan struct{}))

	return
}

func (n *UpdateNode) Completed() bool {
	return n.completed
}

func (n *UpdateNode) Delete() error {
	close(n.stopCh)
	return nil
}

func (n *UpdateNode) Replicas() int {
	return n.replicas
}

func (n *UpdateNode) Kind() string {
	return UpdateNodeKind
}
