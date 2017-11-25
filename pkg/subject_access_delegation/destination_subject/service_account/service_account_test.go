package service_account

import (
	"errors"
	"testing"

	"github.com/golang/mock/gomock"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/joshvanl/k8s-subject-access-delegation/pkg/subject_access_delegation/mocks"
)

type fakeServiceAccount struct {
	*ServiceAccount
	ctrl *gomock.Controller

	fakeClient                  *mocks.MockInterface
	fakeCore                    *mocks.MockCoreV1Interface
	fakeServiceAccountInterface *mocks.MockServiceAccountInterface
}

func newFakeServiceAccount(t *testing.T) *fakeServiceAccount {
	s := &fakeServiceAccount{
		ctrl: gomock.NewController(t),
		ServiceAccount: &ServiceAccount{
			namespace: "fakeNamespace",
			name:      "fakeName",
		},
	}

	s.fakeClient = mocks.NewMockInterface(s.ctrl)
	s.fakeCore = mocks.NewMockCoreV1Interface(s.ctrl)
	s.fakeServiceAccountInterface = mocks.NewMockServiceAccountInterface(s.ctrl)
	s.ServiceAccount.client = s.fakeClient

	s.fakeClient.EXPECT().Core().Times(1).Return(s.fakeCore)
	s.fakeCore.EXPECT().ServiceAccounts(s.ServiceAccount.namespace).Times(1).Return(s.fakeServiceAccountInterface)

	return s
}

func TestServiceAccount_ResolveDestination_Nil(t *testing.T) {
	s := newFakeServiceAccount(t)
	defer s.ctrl.Finish()

	s.fakeServiceAccountInterface.EXPECT().Get(s.ServiceAccount.name, metav1.GetOptions{}).Times(1).Return(nil, nil)

	if err := s.ResolveDestination(); err == nil {
		t.Errorf("expected error but got none - pod is nil")
	}
}

func TestServiceAccount_ResolveDestination_Error(t *testing.T) {
	s := newFakeServiceAccount(t)
	defer s.ctrl.Finish()

	s.fakeServiceAccountInterface.EXPECT().Get(s.ServiceAccount.name, metav1.GetOptions{}).Times(1).Return(&corev1.ServiceAccount{}, errors.New("this is an error"))

	if err := s.ResolveDestination(); err == nil {
		t.Errorf("expected error but got none - returned error")
	}
}

func TestServiceAccount_ResolveDestination_Successful(t *testing.T) {
	s := newFakeServiceAccount(t)
	defer s.ctrl.Finish()

	s.fakeServiceAccountInterface.EXPECT().Get(s.ServiceAccount.name, metav1.GetOptions{}).Times(1).Return(&corev1.ServiceAccount{}, nil)

	if err := s.ResolveDestination(); err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}
