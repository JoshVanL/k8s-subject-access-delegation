package cluster_role

import (
	"errors"
	"reflect"
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

func TestPod_ResolveOrigin_Nil(t *testing.T) {
	r := newFakeRole(t)
	defer r.ctrl.Finish()

	r.fakeRoleInterface.EXPECT().Get(r.Role.name, metav1.GetOptions{}).Times(1).Return(nil, nil)

	if err := r.ResolveOrigin(); err == nil {
		t.Error("expected error but got none - role is nil")
	}
}

func TestPod_ResolveOrigin_Error(t *testing.T) {
	r := newFakeRole(t)
	defer r.ctrl.Finish()

	r.fakeRoleInterface.EXPECT().Get(r.Role.name, metav1.GetOptions{}).Times(1).Return(&rbacv1.Role{}, errors.New("this is an error"))

	if err := r.ResolveOrigin(); err == nil {
		t.Error("expected error but got none - returned error")
	}
}

func TestPod_ResolveOrigin_Successful(t *testing.T) {
	r := newFakeRole(t)
	defer r.ctrl.Finish()

	r.fakeRoleInterface.EXPECT().Get(r.Role.name, metav1.GetOptions{}).Times(1).Return(&rbacv1.Role{}, nil)

	if err := r.ResolveOrigin(); err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestPod_ResolveOrigin_RoleRefs(t *testing.T) {
	r := newFakeRole(t)
	defer r.ctrl.Finish()

	r.fakeRoleInterface.EXPECT().Get(r.Role.name, metav1.GetOptions{}).Times(1).Return(&rbacv1.Role{}, nil)
	if err := r.ResolveOrigin(); err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	roleRefs := []*rbacv1.RoleRef{
		&rbacv1.RoleRef{
			Kind: "Role",
			Name: r.name,
		}}

	refs, err := r.RoleRefs()
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	if !reflect.DeepEqual(refs, roleRefs) {
		t.Errorf("unexpected role refs exp=%+v got=%+v", roleRefs, refs)
	}
}
