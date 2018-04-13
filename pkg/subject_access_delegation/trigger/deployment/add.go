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

const AddDeploymentKind = "AddDeployment"

type AddDeployment struct {
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

var _ interfaces.Trigger = &AddDeployment{}

func NewAddDeployment(sad interfaces.SubjectAccessDelegation, trigger *authzv1alpha1.EventTrigger) (*AddDeployment, error) {

	if !utils.ValidName(trigger.Value) {
		return nil, fmt.Errorf("not a valid name '%s', must contain only alphanumerics, '-', '.' and '*'", trigger.Value)
	}

	deploymentTrigger := &AddDeployment{
		log:            sad.Log(),
		sad:            sad,
		deploymentName: trigger.Value,
		replicas:       trigger.Replicas,
		stopCh:         make(chan struct{}),
		completedCh:    make(chan struct{}),
		count:          0,
		completed:      trigger.Triggered,
		informer:       sad.KubeInformerFactory().Apps().V1().Deployments(),
		uid:            trigger.UID,
	}

	deploymentTrigger.informer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: deploymentTrigger.addFunc,
	})

	return deploymentTrigger, nil
}

func (p *AddDeployment) addFunc(obj interface{}) {

	deployment, err := utils.GetDeploymentObject(obj)
	if err != nil {
		p.log.Errorf("failed to get added deployment object: %v", err)
		return
	}
	if deployment == nil {
		p.log.Error("failed to get deployment, received nil object")
	}

	match, err := utils.MatchName(deployment.Name, p.deploymentName)
	if err != nil {
		p.log.Error("failed to match deployment name: %v", err)
		return
	}

	if !match || p.completed {
		return
	}

	p.sad.AddUid(deployment.UID)

	p.log.Infof("A new deployment '%s' has been added", deployment.Name)
	p.count++
	if p.count >= p.replicas {
		p.log.Infof("Required replicas was met")
		p.completed = true
		close(p.completedCh)
	}
}

func (p *AddDeployment) WaitOn() (forceClosed bool) {
	p.log.Debug("Trigger waiting")

	if p.completed {
		return false
	}

	if p.watchChannels() {
		p.log.Debug("Add Deployment Trigger was force closed")
		return true
	}

	p.log.Debug("Add Deployment Trigger completed")

	if err := p.sad.UpdateTriggerFired(p.uid, true); err != nil {
		p.log.Errorf("error updating add deployment trigger status: %v", err)
	}

	return false
}

func (p *AddDeployment) watchChannels() bool {
	for {
		select {
		case <-p.stopCh:
			p.log.Debug("Add Deployment Trigger was force closed")
			return true

		case <-p.completedCh:
			return false
		}
	}
}

func (p *AddDeployment) Activate() {
	p.log.Debug("Add Deployment Trigger Activated")
	p.completed = false

	go p.informer.Informer().Run(p.completedCh)

	return
}

func (p *AddDeployment) Completed() bool { return p.completed }

func (p *AddDeployment) Delete() error {
	close(p.stopCh)
	return nil
}

func (p *AddDeployment) Replicas() int { return p.replicas }

func (p *AddDeployment) Kind() string { return AddDeploymentKind }
