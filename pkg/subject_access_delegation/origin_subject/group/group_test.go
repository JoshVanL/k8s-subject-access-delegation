package group

import (
	"errors"
	"testing"

	"github.com/golang/mock/gomock"
	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/joshvanl/k8s-subject-access-delegation/pkg/subject_access_delegation/mocks"
)

var (
	noRoleBindings        = new(rbacv1.RoleBindingList)
	noClusterRoleBindings = new(rbacv1.ClusterRoleBindingList)
	someBindings          = []*rbacv1.RoleBinding{
		&rbacv1.RoleBinding{
			RoleRef: rbacv1.RoleRef{
				Name: "a-name-1",
				Kind: "Group",
			},
			Subjects: []rbacv1.Subject{
				rbacv1.Subject{
					Kind: "Group",
					Name: "me",
				},
				rbacv1.Subject{
					Kind: "Group",
					Name: "notme",
				},
			},
		},
		&rbacv1.RoleBinding{
			RoleRef: rbacv1.RoleRef{
				Name: "a-name-2",
				Kind: "Group",
			},
			Subjects: []rbacv1.Subject{
				rbacv1.Subject{
					Kind: "Group",
					Name: "notme",
				},
			},
		},
	}
	someClusterBindings = []*rbacv1.ClusterRoleBinding{
		&rbacv1.ClusterRoleBinding{
			RoleRef: rbacv1.RoleRef{
				Name: "cluster-name-1",
				Kind: "Group",
			},
			Subjects: []rbacv1.Subject{
				rbacv1.Subject{
					Kind: "Group",
					Name: "me",
				},
				rbacv1.Subject{
					Kind: "Group",
					Name: "notme",
				},
			},
		},
		&rbacv1.ClusterRoleBinding{
			RoleRef: rbacv1.RoleRef{
				Name: "cluster-name-2",
			},
			Subjects: []rbacv1.Subject{
				rbacv1.Subject{
					Kind: "Group",
					Name: "notme",
				},
			},
		},
	}

	bindingsList = &rbacv1.RoleBindingList{
		Items: []rbacv1.RoleBinding{
			*someBindings[0],
			*someBindings[1],
		},
	}

	clusterBindingList = &rbacv1.ClusterRoleBindingList{
		Items: []rbacv1.ClusterRoleBinding{
			*someClusterBindings[0],
			*someClusterBindings[1],
		},
	}
)

type fakeGroup struct {
	*Group
	ctrl *gomock.Controller

	fakeClient *mocks.MockInterface
	fakeRbac   *mocks.MockRbacV1Interface
	fakeCore   *mocks.MockCoreV1Interface

	fakeRoleBindingsInterface        *mocks.MockRoleBindingInterface
	fakeClusterRoleBindingsInterface *mocks.MockClusterRoleBindingInterface

	fakeBindingInformer        *mocks.MockRoleBindingInformer
	fakeClusterBindingInformer *mocks.MockClusterRoleBindingInformer
	fakeSharedIndexInformer    *mocks.MockSharedIndexInformer
}

func newFakeGroup(t *testing.T) *fakeGroup {
	u := &fakeGroup{
		ctrl: gomock.NewController(t),
		Group: &Group{
			namespace: "fakeNamespace",
			name:      "me",
		},
	}

	u.fakeClient = mocks.NewMockInterface(u.ctrl)
	u.fakeRbac = mocks.NewMockRbacV1Interface(u.ctrl)
	u.fakeRoleBindingsInterface = mocks.NewMockRoleBindingInterface(u.ctrl)
	u.fakeClusterRoleBindingsInterface = mocks.NewMockClusterRoleBindingInterface(u.ctrl)
	u.fakeCore = mocks.NewMockCoreV1Interface(u.ctrl)

	u.fakeBindingInformer = mocks.NewMockRoleBindingInformer(u.ctrl)
	u.fakeClusterBindingInformer = mocks.NewMockClusterRoleBindingInformer(u.ctrl)
	u.fakeSharedIndexInformer = mocks.NewMockSharedIndexInformer(u.ctrl)
	u.Group.bindingInformer = u.fakeBindingInformer
	u.Group.clusterBindingInformer = u.fakeClusterBindingInformer

	u.Group.client = u.fakeClient

	noRoleBindings.Items = make([]rbacv1.RoleBinding, 0)
	noClusterRoleBindings.Items = make([]rbacv1.ClusterRoleBinding, 0)

	return u
}

func TestGroup_GroupRefs_None(t *testing.T) {
	u := newFakeGroup(t)
	defer u.ctrl.Finish()

	refs, clusterRefs := u.RoleRefs()
	if len(refs) != 0 {
		t.Errorf("expected no refs, got=%+v", refs)
	}

	if len(clusterRefs) != 0 {
		t.Errorf("expected no clouster refs, got=%+v", clusterRefs)
	}

}

func TestGroup_GroupRefs_Some(t *testing.T) {
	u := newFakeGroup(t)
	defer u.ctrl.Finish()

	u.Group.bindings = someBindings
	u.Group.clusterBindings = someClusterBindings

	refs, clusterRefs := u.RoleRefs()
	if len(refs) != 2 {
		t.Errorf("expected 2 refs, got=%+v", refs)
	}

	if len(clusterRefs) != 2 {
		t.Errorf("expected 2 refs, got=%+v", clusterRefs)
	}
}

func TestGroup_RoleBindings_ErrorBinding(t *testing.T) {
	u := newFakeGroup(t)
	defer u.ctrl.Finish()

	options := metav1.ListOptions{}

	u.fakeClient.EXPECT().Rbac().Times(1).Return(u.fakeRbac)
	u.fakeRbac.EXPECT().RoleBindings(gomock.Any()).Return(u.fakeRoleBindingsInterface)
	u.fakeRoleBindingsInterface.EXPECT().List(options).Times(1).Return(nil, errors.New("this is an error"))

	if err := u.roleBindings(); err == nil {
		t.Error("expected error, got=none")
	}
}

func TestGroup_RoleBindings_ErrorClusterBinding(t *testing.T) {
	u := newFakeGroup(t)
	defer u.ctrl.Finish()

	options := metav1.ListOptions{}

	u.fakeClient.EXPECT().Rbac().Times(2).Return(u.fakeRbac)

	u.fakeRbac.EXPECT().RoleBindings(gomock.Any()).Return(u.fakeRoleBindingsInterface)
	u.fakeRoleBindingsInterface.EXPECT().List(options).Times(1).Return(nil, nil)

	u.fakeRbac.EXPECT().ClusterRoleBindings().Return(u.fakeClusterRoleBindingsInterface)
	u.fakeClusterRoleBindingsInterface.EXPECT().List(options).Times(1).Return(nil, errors.New("this is an error"))

	if err := u.roleBindings(); err == nil {
		t.Error("expected error, got=none")
	}
}

func TestGroup_RoleBindings_Success(t *testing.T) {
	u := newFakeGroup(t)
	defer u.ctrl.Finish()

	options := metav1.ListOptions{}

	u.fakeClient.EXPECT().Rbac().Times(2).Return(u.fakeRbac)

	u.fakeRbac.EXPECT().RoleBindings(gomock.Any()).Return(u.fakeRoleBindingsInterface)
	u.fakeRoleBindingsInterface.EXPECT().List(options).Times(1).Return(bindingsList, nil)

	u.fakeRbac.EXPECT().ClusterRoleBindings().Return(u.fakeClusterRoleBindingsInterface)
	u.fakeClusterRoleBindingsInterface.EXPECT().List(options).Times(1).Return(clusterBindingList, nil)

	if err := u.roleBindings(); err != nil {
		t.Errorf("unexpected error: %v", err)
		return
	}

	if len(u.bindings) != 1 || len(u.clusterBindings) != 1 {
		t.Errorf("unexpected number of bindings and cluster Rolebindings, expected=1, got=%d, %d", len(u.bindings), len(u.clusterBindings))
		return
	}

	if u.bindings[0].RoleRef.Name != "a-name-1" {
		t.Errorf("unexpected rolebinding name, expected=a-name-1, got=%s", u.bindings[0].RoleRef.Name)
	}

	if u.clusterBindings[0].RoleRef.Name != "cluster-name-1" {
		t.Errorf("unexpected rolebinding name, expected=a-name-1, got=%s", u.clusterBindings[0].RoleRef.Name)
	}
}

func TestGroup_ResolveDestination(t *testing.T) {
	u := newFakeGroup(t)
	defer u.ctrl.Finish()

	u.fakeClient.EXPECT().Rbac().AnyTimes().Return(u.fakeRbac)
	u.fakeRbac.EXPECT().RoleBindings(u.namespace).Times(1).Return(u.fakeRoleBindingsInterface)
	u.fakeRbac.EXPECT().ClusterRoleBindings().Times(1).Return(u.fakeClusterRoleBindingsInterface)
	u.fakeRoleBindingsInterface.EXPECT().List(gomock.Any()).Return(noRoleBindings, nil)
	u.fakeClusterRoleBindingsInterface.EXPECT().List(gomock.Any()).Return(noClusterRoleBindings, nil)

	u.fakeBindingInformer.EXPECT().Informer().AnyTimes().Return(u.fakeSharedIndexInformer)
	u.fakeClusterBindingInformer.EXPECT().Informer().AnyTimes().Return(u.fakeSharedIndexInformer)

	u.fakeSharedIndexInformer.EXPECT().AddEventHandler(gomock.Any()).Times(2)
	u.fakeSharedIndexInformer.EXPECT().Run(gomock.Any()).AnyTimes()

	if err := u.ResolveOrigin(); err != nil {
		t.Errorf("expected nil, go non-nil?!: %v", err)
	}
}
