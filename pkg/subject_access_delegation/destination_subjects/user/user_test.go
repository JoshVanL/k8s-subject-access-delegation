package user

import (
	"errors"
	"testing"

	"github.com/golang/mock/gomock"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/joshvanl/k8s-subject-access-delegation/pkg/subject_access_delegation/mocks"
)

type fakeUser struct {
	*User
	ctrl *gomock.Controller

	fakeClient                  *mocks.MockInterface
	fakeCore                    *mocks.MockCoreV1Interface
	fakeServiceAccountInterface *mocks.MockServiceAccountInterface
}

func newFakeUser(t *testing.T) *fakeUser {

	p := &fakeUser{
		ctrl: gomock.NewController(t),
		User: &User{
			namespace: "fakeNamespace",
			name:      "fakeName",
		},
	}

	p.fakeClient = mocks.NewMockInterface(p.ctrl)
	p.fakeCore = mocks.NewMockCoreV1Interface(p.ctrl)
	p.fakeServiceAccountInterface = mocks.NewMockServiceAccountInterface(p.ctrl)
	p.User.client = p.fakeClient

	p.fakeClient.EXPECT().Core().Times(1).Return(p.fakeCore)
	p.fakeCore.EXPECT().ServiceAccounts(p.User.namespace).Times(1).Return(p.fakeServiceAccountInterface)

	return p
}

func TestUser_ResolveDestination_Nil(t *testing.T) {
	p := newFakeUser(t)
	defer p.ctrl.Finish()

	p.fakeServiceAccountInterface.EXPECT().Get(p.User.name, metav1.GetOptions{}).Times(1).Return(nil, nil)

	if err := p.ResolveDestination(); err == nil {
		t.Errorf("expected error but got none - pod is nil")
	}
}

func TestUser_ResolveDestination_Error(t *testing.T) {
	p := newFakeUser(t)
	defer p.ctrl.Finish()

	p.fakeServiceAccountInterface.EXPECT().Get(p.User.name, metav1.GetOptions{}).Times(1).Return(&corev1.ServiceAccount{}, errors.New("this is an error"))

	if err := p.ResolveDestination(); err == nil {
		t.Errorf("expected error but got none - returned error")
	}
}

func TestUser_ResolveDestination_Successful(t *testing.T) {
	p := newFakeUser(t)
	defer p.ctrl.Finish()

	p.fakeServiceAccountInterface.EXPECT().Get(p.User.name, metav1.GetOptions{}).Times(1).Return(&corev1.ServiceAccount{}, nil)

	if err := p.ResolveDestination(); err != nil {
		t.Errorf("unexpected error: %v")
	}
}
