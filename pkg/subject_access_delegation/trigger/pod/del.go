package pod

import (
	"github.com/sirupsen/logrus"
	informer "k8s.io/client-go/informers/core/v1"
	"k8s.io/client-go/tools/cache"

	authzv1alpha1 "github.com/joshvanl/k8s-subject-access-delegation/pkg/apis/authz/v1alpha1"
	"github.com/joshvanl/k8s-subject-access-delegation/pkg/interfaces"
	"github.com/joshvanl/k8s-subject-access-delegation/pkg/subject_access_delegation/utils"
)

type DelPod struct {
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

var _ interfaces.Trigger = &DelPod{}

func NewDelPod(sad interfaces.SubjectAccessDelegation, trigger *authzv1alpha1.EventTrigger) (*DelPod, error) {
	podTrigger := &DelPod{
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
		DeleteFunc: podTrigger.delFunc,
	})

	return podTrigger, nil
}

func (p *DelPod) delFunc(obj interface{}) {

	pod, err := utils.GetPodObject(obj)
	if err != nil {
		p.log.Errorf("failed to get deleted pod object: %v", err)
		return
	}
	if pod == nil {
		p.log.Error("failed to get pod, received nil object")
	}

	if pod.Name != p.podName {
		return
	}

	p.log.Infof("A pod '%s' has been deleted", pod.Name)
	p.count++
	if p.count >= p.replicas {
		p.log.Infof("Required replicas was met")
		p.completed = true
		close(p.completedCh)
	}
}

func (p *DelPod) WaitOn() (forceClosed bool) {
	p.log.Debug("Trigger waiting")

	if p.watchChannels() {
		p.log.Debug("Del Pod Trigger was force closed")
		return true
	}

	p.log.Debug("Del Pod Trigger completed")
	return false
}

func (p *DelPod) watchChannels() (forceClose bool) {
	select {
	case <-p.stopCh:
		return true
	case <-p.completedCh:
		return false
	}
}

func (p *DelPod) Activate() {
	p.log.Debug("Del Pod Trigger Activated")

	go p.informer.Informer().Run(make(chan struct{}))

	return
}

func (p *DelPod) Completed() bool {
	return p.completed
}

func (p *DelPod) Delete() error {
	close(p.stopCh)
	return nil
}

func (p *DelPod) Replicas() int {
	return p.replicas
}
