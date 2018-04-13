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

const DelDeploymentKind = "DelDeployment"

type DelDeployment struct {
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

var _ interfaces.Trigger = &DelDeployment{}

func NewDelDeployment(sad interfaces.SubjectAccessDelegation, trigger *authzv1alpha1.EventTrigger) (*DelDeployment, error) {

	if !utils.ValidName(trigger.Value) {
		return nil, fmt.Errorf("not a valid name '%s', must contain only alphanumerics, '-', '.' and '*'", trigger.Value)
	}

	deploymentTrigger := &DelDeployment{
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
		DeleteFunc: deploymentTrigger.delFunc,
	})

	return deploymentTrigger, nil
}

func (p *DelDeployment) delFunc(obj interface{}) {

	deployment, err := utils.GetDeploymentObject(obj)
	if err != nil {
		p.log.Errorf("failed to get deleted deployment object: %v", err)
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

	if !match || p.sad.DeletedUid(deployment.UID) || p.completed {
		return
	}

	p.sad.DeleteUid(deployment.UID)

	p.log.Infof("A deployment '%s' has been deleted", deployment.Name)
	p.count++
	if p.count >= p.replicas {
		p.log.Infof("Required replicas was met")
		p.completed = true
		close(p.completedCh)
	}
}

func (p *DelDeployment) WaitOn() (forceClosed bool) {
	p.log.Debug("Trigger waiting")

	if p.watchChannels() {
		p.log.Debug("Del Deployment Trigger was force closed")
		return true
	}

	p.log.Debug("Del Deployment Trigger completed")

	if err := p.sad.UpdateTriggerFired(p.uid, true); err != nil {
		p.log.Errorf("error updating delete deployment trigger status: %v", err)
	}

	return false
}

func (p *DelDeployment) watchChannels() (forceClose bool) {
	for {
		select {
		case <-p.stopCh:
			return true
		case <-p.completedCh:
			return false
		}
	}
}

func (p *DelDeployment) Activate() {
	p.log.Debug("Del Deployment Trigger Activated")
	p.completed = false

	go p.informer.Informer().Run(p.completedCh)

	return
}

func (p *DelDeployment) Completed() bool {
	return p.completed
}

func (p *DelDeployment) Delete() error {
	close(p.stopCh)
	return nil
}

func (p *DelDeployment) Replicas() int {
	return p.replicas
}

func (p *DelDeployment) Kind() string {
	return DelDeploymentKind
}
