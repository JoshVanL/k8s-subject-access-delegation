package role

import (
	"errors"
	"testing"

	"github.com/golang/mock/gomock"
	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/joshvanl/k8s-subject-access-delegation/pkg/subject_access_delegation/mocks"
)

type fakeRole struct {
	*Role
	ctrl *gomock.Controller

	fakeClient        *mocks.MockInterface
	fakeRbac          *mocks.MockRbacV1Interface
	fakeRoleInterface *mocks.MockRoleInterface
}

func newFakeRole(t *testing.T) *fakeRole {
	r := &fakeRole{
		ctrl: gomock.NewController(t),
		Role: &Role{
			namespace: "fakeNamespace",
			name:      "fakeName",
		},
	}

	r.fakeClient = mocks.NewMockInterface(r.ctrl)
	r.fakeRbac = mocks.NewMockRbacV1Interface(r.ctrl)
	r.fakeRoleInterface = mocks.NewMockRoleInterface(r.ctrl)
	r.Role.client = r.fakeClient

	r.fakeClient.EXPECT().Rbac().Times(1).Return(r.fakeRbac)
	r.fakeRbac.EXPECT().Roles(r.Role.namespace).Times(1).Return(r.fakeRoleInterface)

	return r
}

func TestPod_ResolveOrigin_Error(t *testing.T) {
	r := newFakeRole(t)
	defer r.ctrl.Finish()

	r.fakeRoleInterface.EXPECT().Get(r.Role.name, metav1.GetOptions{}).Times(1).Return(nil, errors.New("an error"))

	if err := r.ResolveOrigin(); err == nil {
		t.Error("expected error but got none - role is nil")
	}

	if r.role != nil {
		t.Errorf("expected role to be nil, got=%+v", r.role)
	}
}

func TestPod_ResolveOrigin_Nil(t *testing.T) {
	r := newFakeRole(t)
	defer r.ctrl.Finish()

	r.fakeRoleInterface.EXPECT().Get(r.Role.name, metav1.GetOptions{}).Times(1).Return(nil, nil)

	if err := r.ResolveOrigin(); err == nil {
		t.Error("expected error but got none - returned error")
	}

	if r.role != nil {
		t.Errorf("expected role to be nil, got=%+v", r.role)
	}
}

func TestPod_ResolveOrigin_Successful(t *testing.T) {
	r := newFakeRole(t)
	defer r.ctrl.Finish()

	aRole := &rbacv1.Role{}
	aRole.Name = "me"

	r.fakeRoleInterface.EXPECT().Get(r.Role.name, metav1.GetOptions{}).Times(1).Return(aRole, nil)

	if err := r.ResolveOrigin(); err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	if r.role.Name != "me" {
		t.Errorf("unexpected role name, expected=me, got=%s", r.role.Name)
	}
}

func TestPod_RoleRefs(t *testing.T) {
	r := newFakeRole(t)
	defer r.ctrl.Finish()

	aRole := &rbacv1.Role{}
	aRole.Name = "me"

	r.fakeRoleInterface.EXPECT().Get(r.Role.name, metav1.GetOptions{}).Times(1).Return(aRole, nil)

	if err := r.ResolveOrigin(); err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	if r.role.Name != "me" {
		t.Errorf("unexpected role name, expected=me, got=%s", r.role.Name)
	}

	refs, clusterRefs := r.RoleRefs()
	if len(clusterRefs) != 0 {
		t.Errorf("unexpected cluster role refsL %+v", clusterRefs)
	}

	if len(refs) != 1 {
		t.Errorf("unexpected number of refs: %+v", refs)
		return
	}

	if refs[0].Name != "fakeName" {
		t.Errorf("unexpected role ref name, expected=fakeName, got=%s", refs[0].Name)
	}

	if refs[0].Kind != "Role" {
		t.Errorf("unexpected role ref kind, expected=Role, got=%s", refs[0].Kind)
	}
}
