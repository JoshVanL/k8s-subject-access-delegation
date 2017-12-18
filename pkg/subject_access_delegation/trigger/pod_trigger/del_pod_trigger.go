package pod_trigger

import (
	"fmt"
	//"reflect"

	"github.com/sirupsen/logrus"
	informer "k8s.io/client-go/informers/core/v1"
	"k8s.io/client-go/tools/cache"

	authzv1alpha1 "github.com/joshvanl/k8s-subject-access-delegation/pkg/apis/authz/v1alpha1"
	"github.com/joshvanl/k8s-subject-access-delegation/pkg/subject_access_delegation/interfaces"
	"github.com/joshvanl/k8s-subject-access-delegation/pkg/subject_access_delegation/utils"
)

type DelPodTrigger struct {
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

var _ interfaces.Trigger = &DelPodTrigger{}

func NewDelPodTrigger(sad interfaces.SubjectAccessDelegation, trigger *authzv1alpha1.EventTrigger) (podTrigger *DelPodTrigger, err error) {
	podTrigger = &DelPodTrigger{
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
		//AddFunc: podTrigger.addFunc,
		//UpdateFunc: nil,
		//UpdateFunc: func(old, new interface{}) {
		//	if !reflect.DeepEqual(old, new) {
		//		podTrigger.addFunc(new)
		//	}
		//},
		DeleteFunc: func(obj interface{}) {
			podTrigger.delFunc(obj)
		},
	})

	fmt.Printf("%v", podTrigger.podName)

	return podTrigger, nil
}

func (p *DelPodTrigger) delFunc(obj interface{}) {

	pod, err := utils.GetPodObject(p.informer.Lister(), obj)
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

func (p *DelPodTrigger) WaitOn() (forceClosed bool) {
	p.log.Debug("Trigger waiting")

	if p.watchChannels() {
		p.log.Debug("Del Pod Trigger was force closed")
		return true
	}

	p.log.Debug("Del Pod Trigger completed")
	return false
}

func (p *DelPodTrigger) watchChannels() (forceClose bool) {
	select {
	case <-p.stopCh:
		return true
	case <-p.completedCh:
		return false
	}
}

//func (p *AddPodTrigger) updateFunc(obj interface{}) {
//	p.log.Infof("updateFunc")
//}

//func (p *AddPodTrigger) deleteFunc(obj interface{}) {
//	p.log.Infof("deleteFunc")
//}

func (p *DelPodTrigger) Activate() {
	p.log.Debug("Del Pod Trigger Activated")
	p.informer.Informer().Run(p.stopCh)
	return
}

func (p *DelPodTrigger) Completed() bool {
	return p.completed
}

func (p *DelPodTrigger) Delete() error {
	close(p.stopCh)
	return nil
}

func (p *DelPodTrigger) Replicas() int {
	return p.replicas
}
