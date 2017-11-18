package pod_trigger

//TODO: Have one parent listener for each research type e.g. one pod, deployment tigger listener that sends info to all relevant trigger children -- reduces api overhead

import (
	//"fmt"
	//"reflect"
	//"time"

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

func New(sad interfaces.SubjectAccessDelegation, trigger *authzv1alpha1.EventTrigger) (podTrigger *AddPodTrigger, err error) {
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
		AddFunc:    podTrigger.addFunc,
		UpdateFunc: nil,
		//UpdateFunc: func(old, new interface{}) {
		//	if !reflect.DeepEqual(old, new) {
		//		podTrigger.updateFunc(new)
		//	}
		//},
		DeleteFunc: nil,
	})

	return podTrigger, nil
}

func (p *AddPodTrigger) addFunc(obj interface{}) {
	pod, err := utils.GetPodObject(p.informer.Lister(), obj)
	if err != nil {
		p.log.Errorf("failed to get added pod object: %v", err)
	}

	if pod.Name == p.podName {
		p.log.Infof("A NEW POD WAS ADDED!!! HERE ########")
	}

	p.count++
	if p.replicas <= p.count {
		p.log.Infof("Required replicas was met")
		p.completed = true
		close(p.completedCh)
	}
}

func (p *AddPodTrigger) WaitOn() (forceClosed bool, err error) {
	//t.log.Debug("Trigger waiting")

	//if t.watchChannels() {
	//	t.log.Debug("Time Trigger was force closed")
	//	return true, nil
	//}

	//t.log.Debug("Time Trigger time expired")

	select {
	case <-p.stopCh:
		return true, nil
	case <-p.completedCh:
		return false, nil
	}

	return false, nil
}

//func (p *AddPodTrigger) updateFunc(obj interface{}) {
//	p.log.Infof("updateFunc")
//}

//func (p *AddPodTrigger) deleteFunc(obj interface{}) {
//	p.log.Infof("deleteFunc")
//}

func (p *AddPodTrigger) Activate() {
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
