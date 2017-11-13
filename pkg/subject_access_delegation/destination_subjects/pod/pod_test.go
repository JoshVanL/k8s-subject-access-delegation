package pod

import (
	"errors"
	"testing"

	"github.com/golang/mock/gomock"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/joshvanl/k8s-subject-access-delegation/pkg/subject_access_delegation/mocks"
)

type fakePod struct {
	*Pod
	ctrl *gomock.Controller

	fakeClient       *mocks.MockInterface
	fakeCore         *mocks.MockCoreV1Interface
	fakePodInterface *mocks.MockPodInterface
}

func newFakePod(t *testing.T) *fakePod {
	p := &fakePod{
		ctrl: gomock.NewController(t),
		Pod: &Pod{
			namespace: "fakeNamespace",
			name:      "fakeName",
		},
	}

	p.fakeClient = mocks.NewMockInterface(p.ctrl)
	p.fakeCore = mocks.NewMockCoreV1Interface(p.ctrl)
	p.fakePodInterface = mocks.NewMockPodInterface(p.ctrl)
	p.Pod.client = p.fakeClient

	p.fakeClient.EXPECT().Core().Times(1).Return(p.fakeCore)
	p.fakeCore.EXPECT().Pods(p.Pod.namespace).Times(1).Return(p.fakePodInterface)

	return p
}

func TestPod_ResolveDestination_Nil(t *testing.T) {
	p := newFakePod(t)
	defer p.ctrl.Finish()

	p.fakePodInterface.EXPECT().Get(p.Pod.name, metav1.GetOptions{}).Times(1).Return(nil, nil)

	if err := p.ResolveDestination(); err == nil {
		t.Errorf("expected error but got none - pod is nil")
	}
}

func TestPod_ResolveDestination_Error(t *testing.T) {
	p := newFakePod(t)
	defer p.ctrl.Finish()

	p.fakePodInterface.EXPECT().Get(p.Pod.name, metav1.GetOptions{}).Times(1).Return(&corev1.Pod{}, errors.New("this is an error"))

	if err := p.ResolveDestination(); err == nil {
		t.Errorf("expected error but got none - returned error")
	}
}

func TestPod_ResolveDestination_Successful(t *testing.T) {
	p := newFakePod(t)
	defer p.ctrl.Finish()

	p.fakePodInterface.EXPECT().Get(p.Pod.name, metav1.GetOptions{}).Times(1).Return(&corev1.Pod{}, nil)

	if err := p.ResolveDestination(); err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}
