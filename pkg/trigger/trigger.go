package trigger

import (
	"fmt"
	"time"

	"github.com/sirupsen/logrus"
	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	//rbacapi "k8s.io/kubernetes/pkg/apis/rbac"

	authzv1alpha1 "github.com/joshvanl/k8s-subject-access-delegation/pkg/apis/authz/v1alpha1"
)

type Trigger struct {
	log          *logrus.Entry
	creationTime *time.Time

	sad    *authzv1alpha1.SubjectAccessDelegation
	client kubernetes.Interface
}

func New(log *logrus.Entry, sad *authzv1alpha1.SubjectAccessDelegation, client kubernetes.Interface) *Trigger {
	now := time.Now()

	return &Trigger{
		log:          log,
		creationTime: &now,

		sad:    sad,
		client: client,
	}
}

func (t *Trigger) Validate() error {
	options := meta_v1.GetOptions{}
	roles := t.client.Rbac().Roles(t.sad.Spec.DestinationSubject.Namespace)

	role, err := roles.Get(t.sad.Spec.OriginSubject.Name, options)
	if err != nil {
		return fmt.Errorf("failed to get role %s: %v", t.sad.Spec.OriginSubject.Name, err)
	}

	fmt.Printf("%s", role.String())

	return nil
}

func (t *Trigger) TickTock() error {
	delta := time.Second * time.Duration(t.sad.Spec.Duration)
	ticker := time.NewTicker(delta)
	<-ticker.C

	//Get roles of origin subject
	// Update to origin of subject

	return nil
}

func (t *Trigger) Duration() int64 {
	return t.sad.Spec.Duration
}

func (t *Trigger) CreationTime() *time.Time {
	return t.creationTime
}

func (t *Trigger) Repeat() int {
	return t.sad.Spec.Repeat
}
