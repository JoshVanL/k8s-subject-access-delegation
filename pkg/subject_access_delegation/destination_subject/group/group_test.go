package group

import (
	"testing"

	"github.com/golang/mock/gomock"

	"github.com/joshvanl/k8s-subject-access-delegation/pkg/subject_access_delegation/mocks"
)

type fakeGroup struct {
	*Group
	ctrl *gomock.Controller

	fakeClient *mocks.MockInterface
	fakeCore   *mocks.MockCoreV1Interface
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
	u.Group.client = u.fakeClient

	return u
}

func TestGroup_ResolveDestination_Nil(t *testing.T) {
	u := newFakeGroup(t)
	defer u.ctrl.Finish()

	if err := u.ResolveDestination(); err != nil {
		t.Errorf("expected no error but got one?!: %v", err)
	}
}
