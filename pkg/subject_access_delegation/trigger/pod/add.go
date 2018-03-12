package pod

import (
	"github.com/sirupsen/logrus"
	informer "k8s.io/client-go/informers/core/v1"
	"k8s.io/client-go/tools/cache"

	authzv1alpha1 "github.com/joshvanl/k8s-subject-access-delegation/pkg/apis/authz/v1alpha1"
	"github.com/joshvanl/k8s-subject-access-delegation/pkg/interfaces"
	"github.com/joshvanl/k8s-subject-access-delegation/pkg/subject_access_delegation/utils"
)

const AddPodKind = "AddPod"

type AddPod struct {
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

var _ interfaces.Trigger = &AddPod{}

func NewAddPod(sad interfaces.SubjectAccessDelegation, trigger *authzv1alpha1.EventTrigger) (*AddPod, error) {
	podTrigger := &AddPod{
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

func (p *AddPod) addFunc(obj interface{}) {

	pod, err := utils.GetPodObject(obj)
	if err != nil {
		p.log.Errorf("failed to get added pod object: %v", err)
		return
	}
	if pod == nil {
		p.log.Error("failed to get pod, received nil object")
	}

	if pod.Name != p.podName || p.sad.SeenUid(pod.UID) {
		return
	}

	p.sad.AddUid(pod.UID)

	p.log.Infof("A new pod '%s' has been added", pod.Name)
	p.count++
	if p.count >= p.replicas {
		p.log.Infof("Required replicas was met")
		p.completed = true
		close(p.completedCh)
	}
}

func (p *AddPod) WaitOn() (forceClosed bool) {
	p.log.Debug("Trigger waiting")

	if p.watchChannels() {
		p.log.Debug("Add Pod Trigger was force closed")
		return true
	}

	p.log.Debug("Add Pod Trigger completed")
	return false
}

func (p *AddPod) watchChannels() (forceClose bool) {
	select {
	case <-p.stopCh:
		return true
	case <-p.completedCh:
		return false
	}
}

func (p *AddPod) Activate() {
	p.log.Debug("Add Pod Trigger Activated")

	go p.informer.Informer().Run(make(chan struct{}))

	return
}

func (p *AddPod) Completed() bool {
	return p.completed
}

func (p *AddPod) Delete() error {
	close(p.stopCh)
	return nil
}

func (p *AddPod) Replicas() int {
	return p.replicas
}

func (p *AddPod) Kind() string {
	return AddPodKind
}
