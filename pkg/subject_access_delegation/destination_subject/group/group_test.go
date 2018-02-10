package group

import (
	"errors"
	"testing"

	"github.com/golang/mock/gomock"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/joshvanl/k8s-subject-access-delegation/pkg/subject_access_delegation/mocks"
)

type fakeGroup struct {
	*Group
	ctrl *gomock.Controller

	fakeClient                  *mocks.MockInterface
	fakeCore                    *mocks.MockCoreV1Interface
	fakeServiceAccountInterface *mocks.MockServiceAccountInterface
}

func newFakeGroup(t *testing.T) *fakeGroup {
	u := &fakeGroup{
		ctrl: gomock.NewController(t),
		Group: &Group{
			namespace: "fakeNamespace",
			name:      "fakeName",
		},
	}

	u.fakeClient = mocks.NewMockInterface(u.ctrl)
	u.fakeCore = mocks.NewMockCoreV1Interface(u.ctrl)
	u.fakeServiceAccountInterface = mocks.NewMockServiceAccountInterface(u.ctrl)
	u.Group.client = u.fakeClient

	u.fakeClient.EXPECT().Core().Times(1).Return(u.fakeCore)
	u.fakeCore.EXPECT().ServiceAccounts(u.Group.namespace).Times(1).Return(u.fakeServiceAccountInterface)

	return u
}

func TestGroup_ResolveDestination_Nil(t *testing.T) {
	u := newFakeGroup(t)
	defer u.ctrl.Finish()

	u.fakeServiceAccountInterface.EXPECT().Get(u.Group.name, metav1.GetOptions{}).Times(1).Return(nil, nil)

	if err := u.ResolveDestination(); err == nil {
		t.Errorf("expected error but got none - user is nil")
	}
}

func TestGroup_ResolveDestination_Error(t *testing.T) {
	u := newFakeGroup(t)
	defer u.ctrl.Finish()

	u.fakeServiceAccountInterface.EXPECT().Get(u.Group.name, metav1.GetOptions{}).Times(1).Return(&corev1.ServiceAccount{}, errors.New("this is an error"))

	if err := u.ResolveDestination(); err == nil {
		t.Errorf("expected error but got none - returned error")
	}
}

func TestGroup_ResolveDestination_Successful(t *testing.T) {
	u := newFakeGroup(t)
	defer u.ctrl.Finish()

	u.fakeServiceAccountInterface.EXPECT().Get(u.Group.name, metav1.GetOptions{}).Times(1).Return(&corev1.ServiceAccount{}, nil)

	if err := u.ResolveDestination(); err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}
