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

const UpdatePodKind = "UpdatePodKind"

type UpdatePod struct {
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

var _ interfaces.Trigger = &UpdatePod{}

func NewUpdatePod(sad interfaces.SubjectAccessDelegation, trigger *authzv1alpha1.EventTrigger) (*UpdatePod, error) {

	if !utils.ValidName(trigger.Value) {
		return nil, fmt.Errorf("not a valid name '%s', must contain only alphanumerics, '-', '.' and '*'", trigger.Value)
	}

	podTrigger := &UpdatePod{
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
		UpdateFunc: podTrigger.updateFunc,
	})

	return podTrigger, nil
}

func (p *UpdatePod) updateFunc(oldObj, newObj interface{}) {

	old, err := utils.GetPodObject(oldObj)
	if err != nil {
		p.log.Errorf("failed to get old updated pod object: %v", err)
		return
	}

	new, err := utils.GetPodObject(newObj)
	if err != nil {
		p.log.Errorf("failed to get new updated pod object: %v", err)
		return
	}
	if new == nil || old == nil {
		p.log.Error("failed to get pod, received nil object")
	}

	match, err := utils.MatchName(old.Name, p.podName)
	if err != nil {
		p.log.Error("failed to match pod name: %v", err)
		return
	}

	if !match || p.sad.DeletedUid(old.UID) || p.completed {
		return
	}

	p.sad.AddUid(new.UID)

	p.log.Infof("A pod '%s' has been updated", new.Name)
	p.count++
	if p.count >= p.replicas {
		p.log.Infof("Required replicas was met")
		p.completed = true
		close(p.completedCh)
	}
}

func (p *UpdatePod) WaitOn() (forceClosed bool) {
	p.log.Debug("Trigger waiting")

	if p.watchChannels() {
		p.log.Debug("Update Pod Trigger was force closed")
		return true
	}

	p.log.Debug("Update Pod Trigger completed")

	if err := p.sad.UpdateTriggerFired(p.uid, true); err != nil {
		p.log.Errorf("error updating update pod trigger status: %v", err)
	}

	return false
}

func (p *UpdatePod) watchChannels() (forceClose bool) {
	for {
		select {
		case <-p.stopCh:
			return true
		case <-p.completedCh:
			return false
		}
	}
}

func (p *UpdatePod) Activate() {
	p.log.Debug("Update Pod Trigger Activated")
	p.completed = false

	go p.informer.Informer().Run(p.completedCh)

	return
}

func (p *UpdatePod) Completed() bool {
	return p.completed
}

func (p *UpdatePod) Delete() error {
	close(p.stopCh)
	return nil
}

func (p *UpdatePod) Replicas() int {
	return p.replicas
}

func (p *UpdatePod) Kind() string {
	return UpdatePodKind
}
