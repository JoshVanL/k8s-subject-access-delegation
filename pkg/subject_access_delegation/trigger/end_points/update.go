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

const UpdateEndPointsKind = "UpdateEndPoints"

type UpdateEndPoints struct {
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

var _ interfaces.Trigger = &UpdateEndPoints{}

func NewUpdateEndPoints(sad interfaces.SubjectAccessDelegation, trigger *authzv1alpha1.EventTrigger) (*UpdateEndPoints, error) {

	if !utils.ValidName(trigger.Value) {
		return nil, fmt.Errorf("not a valid name '%s', must contain only alphanumerics, '-', '.' and '*'", trigger.Value)
	}

	endPointsTrigger := &UpdateEndPoints{
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
		UpdateFunc: endPointsTrigger.updateFunc,
	})

	return endPointsTrigger, nil
}

func (n *UpdateEndPoints) updateFunc(oldObj, newObj interface{}) {

	old, err := utils.GetEndPointsObject(oldObj)
	if err != nil {
		n.log.Errorf("failed to get updated endPoints object: %v", err)
		return
	}

	new, err := utils.GetEndPointsObject(newObj)
	if err != nil {
		n.log.Errorf("failed to get updated endPoints object: %v", err)
		return
	}
	if new == nil || old == nil {
		n.log.Error("failed to get endPoints, received nil object")
	}

	match, err := utils.MatchName(old.Name, n.endPointsName)
	if err != nil {
		n.log.Error("failed to match endpoints name: %v", err)
		return
	}

	if !match || n.sad.DeletedUid(new.UID) || n.completed {
		return
	}

	n.sad.AddUid(new.UID)

	n.log.Infof("A endPoints '%s' has been updated", new.Name)
	n.count++
	if n.count >= n.replicas {
		n.log.Infof("Required replicas was met")
		n.completed = true
		close(n.completedCh)
	}
}

func (n *UpdateEndPoints) WaitOn() (forceClosed bool) {
	n.log.Debug("Trigger waiting")

	if n.watchChannels() {
		n.log.Debug("Del EndPoints Trigger was force closed")
		return true
	}

	n.log.Debug("Del EndPoints Trigger completed")
	return false
}

func (n *UpdateEndPoints) watchChannels() (forceClose bool) {
	select {
	case <-n.stopCh:
		return true
	case <-n.completedCh:
		return false
	}
}

func (n *UpdateEndPoints) Activate() {
	n.log.Debug("Del EndPoints Trigger Activated")

	go n.informer.Informer().Run(make(chan struct{}))

	return
}

func (n *UpdateEndPoints) Completed() bool {
	return n.completed
}

func (n *UpdateEndPoints) Delete() error {
	close(n.stopCh)
	return nil
}

func (n *UpdateEndPoints) Replicas() int {
	return n.replicas
}

func (n *UpdateEndPoints) Kind() string {
	return UpdateEndPointsKind
}
