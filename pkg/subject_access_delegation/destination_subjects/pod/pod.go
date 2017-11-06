package pod

import (
	"fmt"

	"github.com/sirupsen/logrus"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"

	"github.com/joshvanl/k8s-subject-access-delegation/pkg/subject_access_delegation/interfaces"
)

type Pod struct {
	log    *logrus.Entry
	client kubernetes.Interface
	sad    interfaces.SubjectAccessDelegation

	namespace string
	name      string
	pod       *corev1.Pod
}

var _ interfaces.DestinationSubject = &Pod{}

func New(sad interfaces.SubjectAccessDelegation, name string) *Pod {
	return &Pod{
		log:       sad.Log(),
		client:    sad.Client(),
		sad:       sad,
		namespace: sad.Namespace(),
		name:      name,
	}
}

func (d *Pod) getPod() error {
	options := metav1.GetOptions{}

	pod, err := d.client.Core().Pods(d.Namespace()).Get(d.Name(), options)
	if err != nil {
		return fmt.Errorf("failed to get pod '%s': %v", d.Name(), err)
	}

	d.pod = pod

	return nil
}

func (d *Pod) ResolveDestination() error {
	return d.getPod()
}

func (d *Pod) Namespace() string {
	return d.namespace
}

func (d *Pod) Name() string {
	return d.name
}

func (d *Pod) Kind() string {
	return "Pod"
}
