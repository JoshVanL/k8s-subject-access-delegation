package user

import (
	"testing"

	"github.com/golang/mock/gomock"

	"github.com/joshvanl/k8s-subject-access-delegation/pkg/subject_access_delegation/mocks"
)

type fakeUser struct {
	*User
	ctrl *gomock.Controller

	fakeClient *mocks.MockInterface
	fakeCore   *mocks.MockCoreV1Interface
}

func newFakeUser(t *testing.T) *fakeUser {
	u := &fakeUser{
		ctrl: gomock.NewController(t),
		User: &User{
			namespace: "fakeNamespace",
			name:      "fakeName",
		},
	}

	u.fakeClient = mocks.NewMockInterface(u.ctrl)
	u.fakeCore = mocks.NewMockCoreV1Interface(u.ctrl)
	u.User.client = u.fakeClient

	return u
}

func TestUser_ResolveDestination(t *testing.T) {
	u := newFakeUser(t)
	defer u.ctrl.Finish()

	if err := u.ResolveDestination(); err != nil {
		t.Errorf("expected errro to always be nil, it's not?!: %v", err)
	}
}
