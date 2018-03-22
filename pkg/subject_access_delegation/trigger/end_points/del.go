package end_points

import (
	"fmt"

	"github.com/sirupsen/logrus"
	informer "k8s.io/client-go/informers/core/v1"
	"k8s.io/client-go/tools/cache"

	authzv1alpha1 "github.com/joshvanl/k8s-subject-access-delegation/pkg/apis/authz/v1alpha1"
	"github.com/joshvanl/k8s-subject-access-delegation/pkg/interfaces"
	"github.com/joshvanl/k8s-subject-access-delegation/pkg/subject_access_delegation/utils"
)

const DelEndPointsKind = "DelEndPoints"

type DelEndPoints struct {
	log *logrus.Entry

	sad           interfaces.SubjectAccessDelegation
	endPointsName string
	replicas      int

	stopCh      chan struct{}
	completedCh chan struct{}

	count     int
	completed bool
	informer  informer.EndpointsInformer
}

var _ interfaces.Trigger = &DelEndPoints{}

func NewDelEndPoints(sad interfaces.SubjectAccessDelegation, trigger *authzv1alpha1.EventTrigger) (*DelEndPoints, error) {

	if !utils.ValidName(trigger.Value) {
		return nil, fmt.Errorf("not a valid name '%s', must contain only alphanumerics, '-', '.' and '*'", trigger.Value)
	}

	endPointsTrigger := &DelEndPoints{
		log:           sad.Log(),
		sad:           sad,
		endPointsName: trigger.Value,
		replicas:      trigger.Replicas,
		stopCh:        make(chan struct{}),
		completedCh:   make(chan struct{}),
		count:         0,
		completed:     false,
		informer:      sad.KubeInformerFactory().Core().V1().Endpoints(),
	}

	endPointsTrigger.informer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		DeleteFunc: endPointsTrigger.delFunc,
	})

	return endPointsTrigger, nil
}

func (n *DelEndPoints) delFunc(obj interface{}) {

	endPoints, err := utils.GetEndPointsObject(obj)
	if err != nil {
		n.log.Errorf("failed to get deleted endPoints object: %v", err)
		return
	}
	if endPoints == nil {
		n.log.Error("failed to get endPoints, received nil object")
	}

	match, err := utils.MatchName(endPoints.Name, n.endPointsName)
	if err != nil {
		n.log.Error("failed to match endpoints name: %v", err)
		return
	}

	if !match || n.sad.DeletedUid(endPoints.UID) {
		return
	}

	n.sad.DeleteUid(endPoints.UID)

	n.log.Infof("A endPoints '%s' has been deleted", endPoints.Name)
	n.count++
	if n.count >= n.replicas {
		n.log.Infof("Required replicas was met")
		n.completed = true
		close(n.completedCh)
	}
}

func (n *DelEndPoints) WaitOn() (forceClosed bool) {
	n.log.Debug("Trigger waiting")

	if n.watchChannels() {
		n.log.Debug("Del EndPoints Trigger was force closed")
		return true
	}

	n.log.Debug("Del EndPoints Trigger completed")
	return false
}

func (n *DelEndPoints) watchChannels() (forceClose bool) {
	select {
	case <-n.stopCh:
		return true
	case <-n.completedCh:
		return false
	}
}

func (n *DelEndPoints) Activate() {
	n.log.Debug("Del EndPoints Trigger Activated")

	go n.informer.Informer().Run(make(chan struct{}))

	return
}

func (n *DelEndPoints) Completed() bool {
	return n.completed
}

func (n *DelEndPoints) Delete() error {
	close(n.stopCh)
	return nil
}

func (n *DelEndPoints) Replicas() int {
	return n.replicas
}

func (n *DelEndPoints) Kind() string {
	return DelEndPointsKind
}
