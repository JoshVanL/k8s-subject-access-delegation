package pod_trigger

//TODO: Have one parent listener for each research type e.g. one pod, deployment tigger listener that sends info to all relevant trigger children -- reduces api overhead

import (
	"github.com/sirupsen/logrus"
	informer "k8s.io/client-go/informers/core/v1"
	"k8s.io/client-go/tools/cache"

	authzv1alpha1 "github.com/joshvanl/k8s-subject-access-delegation/pkg/apis/authz/v1alpha1"
	"github.com/joshvanl/k8s-subject-access-delegation/pkg/subject_access_delegation/interfaces"
	"github.com/joshvanl/k8s-subject-access-delegation/pkg/subject_access_delegation/utils"
)

type AddPodTrigger struct {
	log *logrus.Entry

	sad      interfaces.SubjectAccessDelegation
	podName  string
	replicas int

	stopCh      chan struct{}
	completedCh chan struct{}

	count     int
	completed bool
	informer  informer.PodInformer
}

var _ interfaces.Trigger = &AddPodTrigger{}

func NewAddPodTrigger(sad interfaces.SubjectAccessDelegation, trigger *authzv1alpha1.EventTrigger) (podTrigger *AddPodTrigger, err error) {
	podTrigger = &AddPodTrigger{
		log:         sad.Log(),
		sad:         sad,
		podName:     trigger.Value,
		replicas:    trigger.Replicas,
		stopCh:      make(chan struct{}),
		completedCh: make(chan struct{}),
		count:       0,
		completed:   false,
		informer:    sad.KubeInformerFactory().Core().V1().Pods(),
	}

	podTrigger.informer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		//TODO: Fix this
		AddFunc: podTrigger.addFunc,
	})

	return podTrigger, nil
}

func (p *AddPodTrigger) addFunc(obj interface{}) {

	pod, err := utils.GetPodObject(p.informer.Lister(), obj)
	if err != nil {
		p.log.Errorf("failed to get added pod object: %v", err)
		return
	}
	if pod == nil {
		p.log.Error("failed to get pod, received nil object")
	}

	if pod.Name != p.podName {
		return
	}

	p.log.Infof("A new pod '%s' has been added", pod.Name)
	p.count++
	if p.count >= p.replicas {
		p.log.Infof("Required replicas was met")
		p.completed = true
		close(p.completedCh)
	}
}

func (p *AddPodTrigger) WaitOn() (forceClosed bool) {
	p.log.Debug("Trigger waiting")

	if p.watchChannels() {
		p.log.Debug("Add Pod Trigger was force closed")
		return true
	}

	p.log.Debug("Add Pod Trigger completed")
	return false
}

func (p *AddPodTrigger) watchChannels() (forceClose bool) {
	select {
	case <-p.stopCh:
		return true
	case <-p.completedCh:
		return false
	}
}

func (p *AddPodTrigger) Activate() {
	p.log.Debug("Add Pod Trigger Activated")

	var infStopCh chan struct{}

	go p.informer.Informer().Run(infStopCh)

	return
}

func (p *AddPodTrigger) Completed() bool {
	return p.completed
}

func (p *AddPodTrigger) Delete() error {
	close(p.stopCh)
	return nil
}

func (p *AddPodTrigger) Replicas() int {
	return p.replicas
}
