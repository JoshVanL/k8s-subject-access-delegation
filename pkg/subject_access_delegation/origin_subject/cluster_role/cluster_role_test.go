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

type fakeClusterRole struct {
	*ClusterRole
	ctrl *gomock.Controller

	fakeClient               *mocks.MockInterface
	fakeRbac                 *mocks.MockRbacV1Interface
	fakeClusterRoleInterface *mocks.MockRoleInterface
}

func newFakeClusterRole(t *testing.T) *fakeClusterRole {
	r := &fakeClusterRole{
		ctrl: gomock.NewController(t),
		ClusterRole: &ClusterRole{
			name: "fakeName",
		},
	}

	r.fakeClient = mocks.NewMockInterface(r.ctrl)
	r.fakeRbac = mocks.NewMockRbacV1Interface(r.ctrl)
	r.fakeClusterRoleInterface = mocks.NewMockRoleInterface(r.ctrl)
	r.ClusterRole.client = r.fakeClient

	r.fakeClient.EXPECT().Rbac().Times(1).Return(r.fakeRbac)
	r.fakeRbac.EXPECT().ClusterRoles().Times(1).Return(r.ClusterRole)

	return r
}

func TestPod_ResolveOrigin_Nil(t *testing.T) {
	r := newFakeClusterRole(t)
	defer r.ctrl.Finish()

	r.fakeClusterRoleInterface.EXPECT().Get(r.ClusterRole.name, metav1.GetOptions{}).Times(1).Return(nil, nil)

	if err := r.ResolveOrigin(); err == nil {
		t.Error("expected error but got none - role is nil")
	}
}

func TestPod_ResolveOrigin_Error(t *testing.T) {
	r := newFakeClusterRole(t)
	defer r.ctrl.Finish()

	r.fakeClusterRoleInterface.EXPECT().Get(r.ClusterRole.name, metav1.GetOptions{}).Times(1).Return(&rbacv1.ClusterRole{}, errors.New("this is an error"))

	if err := r.ResolveOrigin(); err == nil {
		t.Error("expected error but got none - returned error")
	}
}

func TestPod_ResolveOrigin_Successful(t *testing.T) {
	r := newFakeClusterRole(t)
	defer r.ctrl.Finish()

	r.fakeClusterRoleInterface.EXPECT().Get(r.ClusterRole.name, metav1.GetOptions{}).Times(1).Return(&rbacv1.ClusterRole{}, nil)

	if err := r.ResolveOrigin(); err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestPod_ResolveOrigin_ClusterRoleRefs(t *testing.T) {
	r := newFakeClusterRole(t)
	defer r.ctrl.Finish()

	r.fakeClusterRoleInterface.EXPECT().Get(r.ClusterRole.name, metav1.GetOptions{}).Times(1).Return(&rbacv1.ClusterRole{}, nil)
	if err := r.ResolveOrigin(); err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	roleRefs := []*rbacv1.RoleRef{
		&rbacv1.RoleRef{
			Kind: "ClusterRole",
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
