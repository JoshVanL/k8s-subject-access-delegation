package deployment

import (
	"fmt"

	"github.com/sirupsen/logrus"
	informer "k8s.io/client-go/informers/apps/v1"
	"k8s.io/client-go/tools/cache"

	authzv1alpha1 "github.com/joshvanl/k8s-subject-access-delegation/pkg/apis/authz/v1alpha1"
	"github.com/joshvanl/k8s-subject-access-delegation/pkg/interfaces"
	"github.com/joshvanl/k8s-subject-access-delegation/pkg/subject_access_delegation/utils"
)

const UpdateDeploymentKind = "UpdateDeployment"

type UpdateDeployment struct {
	log *logrus.Entry

	sad            interfaces.SubjectAccessDelegation
	deploymentName string
	replicas       int
	uid            int

	stopCh      chan struct{}
	completedCh chan struct{}

	count     int
	completed bool
	informer  informer.DeploymentInformer
}

var _ interfaces.Trigger = &UpdateDeployment{}

func NewUpdateDeployment(sad interfaces.SubjectAccessDelegation, trigger *authzv1alpha1.EventTrigger) (*UpdateDeployment, error) {

	if !utils.ValidName(trigger.Value) {
		return nil, fmt.Errorf("not a valid name '%s', must contain only alphanumerics, '-', '.' and '*'", trigger.Value)
	}

	deploymentTrigger := &UpdateDeployment{
		log:            sad.Log(),
		sad:            sad,
		deploymentName: trigger.Value,
		replicas:       trigger.Replicas,
		stopCh:         make(chan struct{}),
		completedCh:    make(chan struct{}),
		count:          0,
		completed:      trigger.Triggered,
		uid:            trigger.UID,
		informer:       sad.KubeInformerFactory().Apps().V1().Deployments(),
	}

	deploymentTrigger.informer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		UpdateFunc: deploymentTrigger.updateFunc,
	})

	return deploymentTrigger, nil
}

func (p *UpdateDeployment) updateFunc(oldObj, newObj interface{}) {

	old, err := utils.GetDeploymentObject(oldObj)
	if err != nil {
		p.log.Errorf("failed to get old updated deployment object: %v", err)
		return
	}

	new, err := utils.GetDeploymentObject(newObj)
	if err != nil {
		p.log.Errorf("failed to get new updated deployment object: %v", err)
		return
	}
	if new == nil || old == nil {
		p.log.Error("failed to get deployment, received nil object")
	}

	match, err := utils.MatchName(old.Name, p.deploymentName)
	if err != nil {
		p.log.Error("failed to match deployment name: %v", err)
		return
	}

	if !match || p.sad.DeletedUid(old.UID) || p.completed {
		return
	}

	p.sad.AddUid(new.UID)

	p.log.Infof("A deployment '%s' has been updated", new.Name)
	p.count++
	if p.count >= p.replicas {
		p.log.Infof("Required replicas was met")
		p.completed = true
		close(p.completedCh)
	}
}

func (p *UpdateDeployment) WaitOn() (forceClosed bool) {
	p.log.Debug("Trigger waiting")

	if p.watchChannels() {
		p.log.Debug("Update Deployment Trigger was force closed")
		return true
	}

	p.log.Debug("Update Deployment Trigger completed")

	if err := p.sad.UpdateTriggerFired(p.uid, true); err != nil {
		p.log.Errorf("error updating update deployment trigger status: %v", err)
	}

	return false
}

func (p *UpdateDeployment) watchChannels() (forceClose bool) {
	for {
		select {
		case <-p.stopCh:
			return true
		case <-p.completedCh:
			return false
		}
	}
}

func (p *UpdateDeployment) Activate() {
	p.log.Debug("Update Deployment Trigger Activated")
	p.completed = false

	go p.informer.Informer().Run(p.completedCh)

	return
}

func (p *UpdateDeployment) Completed() bool { return p.completed }

func (p *UpdateDeployment) Delete() error {
	close(p.stopCh)
	return nil
}

func (p *UpdateDeployment) Replicas() int { return p.replicas }

func (p *UpdateDeployment) Kind() string { return UpdateDeploymentKind }
