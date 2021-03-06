package pod

import (
	"fmt"

	"github.com/sirupsen/logrus"
	informer "k8s.io/client-go/informers/core/v1"
	"k8s.io/client-go/tools/cache"

	authzv1alpha1 "github.com/joshvanl/k8s-subject-access-delegation/pkg/apis/authz/v1alpha1"
	"github.com/joshvanl/k8s-subject-access-delegation/pkg/interfaces"
	"github.com/joshvanl/k8s-subject-access-delegation/pkg/subject_access_delegation/utils"
)

const DelPodKind = "DelPod"

type DelPod struct {
	log *logrus.Entry

	sad      interfaces.SubjectAccessDelegation
	podName  string
	replicas int
	uid      int

	stopCh      chan struct{}
	completedCh chan struct{}

	count     int
	completed bool
	informer  informer.PodInformer
}

var _ interfaces.Trigger = &DelPod{}

func NewDelPod(sad interfaces.SubjectAccessDelegation, trigger *authzv1alpha1.EventTrigger) (*DelPod, error) {

	if !utils.ValidName(trigger.Value) {
		return nil, fmt.Errorf("not a valid name '%s', must contain only alphanumerics, '-', '.' and '*'", trigger.Value)
	}

	podTrigger := &DelPod{
		log:         sad.Log(),
		sad:         sad,
		podName:     trigger.Value,
		replicas:    trigger.Replicas,
		stopCh:      make(chan struct{}),
		completedCh: make(chan struct{}),
		count:       0,
		completed:   trigger.Triggered,
		uid:         trigger.UID,
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

	match, err := utils.MatchName(pod.Name, p.podName)
	if err != nil {
		p.log.Error("failed to match pod name: %v", err)
		return
	}

	if !match || p.sad.DeletedUid(pod.UID) || p.completed {
		return
	}

	p.sad.DeleteUid(pod.UID)

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

	if err := p.sad.UpdateTriggerFired(p.uid, true); err != nil {
		p.log.Errorf("error updating delete pod trigger status: %v", err)
	}

	return false
}

func (p *DelPod) watchChannels() (forceClose bool) {
	for {
		select {
		case <-p.stopCh:
			return true
		case <-p.completedCh:
			return false
		}
	}
}

func (p *DelPod) Activate() {
	p.log.Debug("Del Pod Trigger Activated")
	p.completed = false

	go p.informer.Informer().Run(p.completedCh)

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

func (p *DelPod) Kind() string {
	return DelPodKind
}
