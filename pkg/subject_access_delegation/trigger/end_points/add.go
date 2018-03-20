package end_points

import (
	"github.com/sirupsen/logrus"
	informer "k8s.io/client-go/informers/core/v1"
	"k8s.io/client-go/tools/cache"

	authzv1alpha1 "github.com/joshvanl/k8s-subject-access-delegation/pkg/apis/authz/v1alpha1"
	"github.com/joshvanl/k8s-subject-access-delegation/pkg/interfaces"
	"github.com/joshvanl/k8s-subject-access-delegation/pkg/subject_access_delegation/utils"
)

type AddEndPoints struct {
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

var _ interfaces.Trigger = &AddEndPoints{}

const AddEndPointsKind = "AddEndPoints"

func NewAddEndPoints(sad interfaces.SubjectAccessDelegation, trigger *authzv1alpha1.EventTrigger) (*AddEndPoints, error) {
	endPointsTrigger := &AddEndPoints{
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
		//TODO: Fix this
		AddFunc: endPointsTrigger.addFunc,
	})

	return endPointsTrigger, nil
}

func (n *AddEndPoints) addFunc(obj interface{}) {

	endPoints, err := utils.GetEndPointsObject(obj)
	if err != nil {
		n.log.Errorf("failed to get added endPoints object: %v", err)
		return
	}
	if endPoints == nil {
		n.log.Error("failed to get endPoints, received nil object")
	}

	if endPoints.Name != n.endPointsName || n.sad.SeenUid(endPoints.UID) {
		return
	}

	n.sad.AddUid(endPoints.UID)

	n.log.Infof("A new endPoints '%s' has been added", endPoints.Name)
	n.count++
	if n.count >= n.replicas {
		n.log.Infof("Required replicas was met")
		n.completed = true
		close(n.completedCh)
	}
}

func (n *AddEndPoints) WaitOn() (forceClosed bool) {
	n.log.Debug("Trigger waiting")

	if n.watchChannels() {
		n.log.Debug("Add EndPoints Trigger was force closed")
		return true
	}

	n.log.Debug("Add EndPoints Trigger completed")
	return false
}

func (n *AddEndPoints) watchChannels() (forceClose bool) {
	select {
	case <-n.stopCh:
		return true
	case <-n.completedCh:
		return false
	}
}

func (n *AddEndPoints) Activate() {
	n.log.Debug("Add EndPoints Trigger Activated")

	go n.informer.Informer().Run(make(chan struct{}))

	return
}

func (n *AddEndPoints) Completed() bool {
	return n.completed
}

func (n *AddEndPoints) Delete() error {
	close(n.stopCh)
	return nil
}

func (n *AddEndPoints) Replicas() int {
	return n.replicas
}

func (n *AddEndPoints) Kind() string {
	return AddEndPointsKind
}