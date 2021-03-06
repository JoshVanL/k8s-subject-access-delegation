package cluster_role

import (
	"errors"
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
	fakeClusterRoleInterface *mocks.MockClusterRoleInterface
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
	r.fakeClusterRoleInterface = mocks.NewMockClusterRoleInterface(r.ctrl)
	r.client = r.fakeClient

	r.fakeClient.EXPECT().Rbac().Times(1).Return(r.fakeRbac)
	r.fakeRbac.EXPECT().ClusterRoles().Times(1).Return(r.fakeClusterRoleInterface)

	return r
}

func TestPod_ResolveOrigin_Error(t *testing.T) {
	r := newFakeClusterRole(t)
	defer r.ctrl.Finish()

	r.fakeClusterRoleInterface.EXPECT().Get(r.name, metav1.GetOptions{}).Times(1).Return(nil, errors.New("an error"))

	if err := r.ResolveOrigin(); err == nil {
		t.Error("expected error but got none - cluster role is nil")
	}

	if r.role != nil {
		t.Errorf("expected cluster role to be nil, got=%+v", r.role)
	}
}

func TestPod_ResolveOrigin_Nil(t *testing.T) {
	r := newFakeClusterRole(t)
	defer r.ctrl.Finish()

	r.fakeClusterRoleInterface.EXPECT().Get(r.name, metav1.GetOptions{}).Times(1).Return(nil, nil)

	if err := r.ResolveOrigin(); err == nil {
		t.Error("expected error but got none - returned error")
	}

	if r.role != nil {
		t.Errorf("expected cluster role to be nil, got=%+v", r.role)
	}
}

func TestPod_ResolveOrigin_Successful(t *testing.T) {
	r := newFakeClusterRole(t)
	defer r.ctrl.Finish()

	aClusterRole := new(rbacv1.ClusterRole)
	aClusterRole.Name = "me"

	r.fakeClusterRoleInterface.EXPECT().Get(r.name, metav1.GetOptions{}).Times(1).Return(aClusterRole, nil)

	if err := r.ResolveOrigin(); err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	if r.role.Name != "me" {
		t.Errorf("unexpected cluster role name, expected=me, got=%s", r.role.Name)
	}
}

func TestPod_RoleRefs(t *testing.T) {
	r := newFakeClusterRole(t)
	defer r.ctrl.Finish()

	aClusterRole := new(rbacv1.ClusterRole)
	aClusterRole.Name = "me"

	r.fakeClusterRoleInterface.EXPECT().Get(r.name, metav1.GetOptions{}).Times(1).Return(aClusterRole, nil)

	if err := r.ResolveOrigin(); err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	if r.role.Name != "me" {
		t.Errorf("unexpected cluster role name, expected=me, got=%s", r.role.Name)
	}

	refs, clusterRefs := r.RoleRefs()
	if len(refs) != 0 {
		t.Errorf("unexpected role refs %+v", clusterRefs)
	}

	if len(clusterRefs) != 1 {
		t.Errorf("unexpected number of cluster refs: %+v", refs)
		return
	}

	if clusterRefs[0].Name != "fakeName" {
		t.Errorf("unexpected cluster role ref name, expected=fakeName, got=%s", refs[0].Name)
	}

	if clusterRefs[0].Kind != "ClusterRole" {
		t.Errorf("unexpected cluster role ref kind, expected=Role, got=%s", refs[0].Kind)
	}
}
