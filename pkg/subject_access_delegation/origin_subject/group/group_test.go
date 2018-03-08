package group

import (
	"errors"
	"reflect"
	"testing"

	"github.com/golang/mock/gomock"
	rbacv1 "k8s.io/api/rbac/v1"

	"github.com/joshvanl/k8s-subject-access-delegation/pkg/subject_access_delegation/mocks"
)

type fakeGroup struct {
	*Group
	ctrl *gomock.Controller

	fakeClient                *mocks.MockInterface
	fakeRbac                  *mocks.MockRbacV1Interface
	fakeRoleBindingsInterface *mocks.MockRoleBindingInterface
	fakeCore                  *mocks.MockCoreV1Interface
}

func newFakeGroup(t *testing.T) *fakeGroup {
	u := &fakeGroup{
		ctrl: gomock.NewController(t),
		Group: &Group{
			namespace: "fakeNamespace",
			name:      "me",
			uid:       "me",
		},
	}

	u.fakeClient = mocks.NewMockInterface(u.ctrl)
	u.fakeRbac = mocks.NewMockRbacV1Interface(u.ctrl)
	u.fakeRoleBindingsInterface = mocks.NewMockRoleBindingInterface(u.ctrl)
	u.fakeCore = mocks.NewMockCoreV1Interface(u.ctrl)

	u.Group.client = u.fakeClient

	return u
}

func TestGroup_ResolveDestination(t *testing.T) {
	u := newFakeGroup(t)
	defer u.ctrl.Finish()

	if err := u.ResolveOrigin(); err != nil {
		t.Errorf("expected nil, go non-nil?!: %v", err)
	}
}

func TestGroup_GroupRefs_Nil(t *testing.T) {
	u := newFakeGroup(t)
	defer u.ctrl.Finish()

	u.fakeClient.EXPECT().Rbac().Times(1).Return(u.fakeRbac)
	u.fakeRbac.EXPECT().RoleBindings(u.namespace).Times(1).Return(u.fakeRoleBindingsInterface)
	u.fakeRoleBindingsInterface.EXPECT().List(gomock.Any()).Return(nil, nil)

	_, err := u.RoleRefs()
	if err == nil {
		t.Error("expected error but got none - returned nil")
	}
}

func TestGroup_GroupRefs_Error(t *testing.T) {
	u := newFakeGroup(t)
	defer u.ctrl.Finish()

	u.fakeClient.EXPECT().Rbac().Times(1).Return(u.fakeRbac)
	u.fakeRbac.EXPECT().RoleBindings(u.namespace).Times(1).Return(u.fakeRoleBindingsInterface)
	u.fakeRoleBindingsInterface.EXPECT().List(gomock.Any()).Return(nil, errors.New("this is an error"))

	_, err := u.RoleRefs()
	if err == nil {
		t.Error("expected error but got none - returned error")
	}
}

func TestGroup_GroupRefs_Successful_None(t *testing.T) {
	u := newFakeGroup(t)
	defer u.ctrl.Finish()

	groupRefsReturn := &rbacv1.RoleBindingList{}

	u.fakeClient.EXPECT().Rbac().Times(1).Return(u.fakeRbac)
	u.fakeRbac.EXPECT().RoleBindings(u.namespace).Times(1).Return(u.fakeRoleBindingsInterface)
	u.fakeRoleBindingsInterface.EXPECT().List(gomock.Any()).Return(groupRefsReturn, nil)

	refs, err := u.RoleRefs()
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	if len(refs) != 0 {
		t.Errorf("unexpected role refs exp=nothing got=%+v", refs)
	}
}

func TestGroup_GroupRefs_Successful_All(t *testing.T) {
	u := newFakeGroup(t)
	defer u.ctrl.Finish()

	roleBindingsReturn := &rbacv1.RoleBindingList{}

	var roleRef1, roleRef2, roleRef3 rbacv1.RoleRef
	roleRef1.Name = "roleRef1"
	roleRef1.Kind = "Role"
	roleRef2.Name = "roleRef2"
	roleRef2.Kind = "Role"
	roleRef3.Name = "roleRef3"
	roleRef3.Kind = "Role"

	subjects := []rbacv1.Subject{
		rbacv1.Subject{
			Kind: "Pod",
			Name: "foo",
		},
		rbacv1.Subject{
			Kind: "Group",
			Name: "me",
		},
		rbacv1.Subject{
			Kind: "Pod",
			Name: "me",
		},
		rbacv1.Subject{
			Kind: "Group",
			Name: "bar",
		},
	}

	items := []rbacv1.RoleBinding{
		rbacv1.RoleBinding{
			RoleRef:  roleRef1,
			Subjects: subjects,
		},
		rbacv1.RoleBinding{
			RoleRef:  roleRef2,
			Subjects: subjects,
		},
		rbacv1.RoleBinding{
			RoleRef:  roleRef3,
			Subjects: subjects,
		},
	}

	roleBindingsReturn.Items = items

	roleRefsExp := []*rbacv1.RoleRef{
		&rbacv1.RoleRef{
			Name: "roleRef1",
			Kind: "Role",
		},
		&rbacv1.RoleRef{
			Name: "roleRef2",
			Kind: "Role",
		},
		&rbacv1.RoleRef{
			Name: "roleRef3",
			Kind: "Role",
		},
	}

	u.fakeClient.EXPECT().Rbac().Times(1).Return(u.fakeRbac)
	u.fakeRbac.EXPECT().RoleBindings(u.namespace).Times(1).Return(u.fakeRoleBindingsInterface)
	u.fakeRoleBindingsInterface.EXPECT().List(gomock.Any()).Return(roleBindingsReturn, nil)

	refs, err := u.RoleRefs()
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	if !reflect.DeepEqual(refs, roleRefsExp) {
		t.Errorf("unexpected role refs\nexp=%+v\ngot=%+v", roleRefsExp, refs)
	}
}

func TestGroup_GroupRefs_Successful_Some(t *testing.T) {
	u := newFakeGroup(t)
	defer u.ctrl.Finish()

	roleBindingsReturn := &rbacv1.RoleBindingList{}

	var roleRef1, roleRef2, roleRef3 rbacv1.RoleRef
	roleRef1.Name = "roleRef1"
	roleRef1.Kind = "Role"
	roleRef2.Name = "roleRef2"
	roleRef2.Kind = "Role"
	roleRef3.Name = "roleRef3"
	roleRef3.Kind = "Role"

	subjects := []rbacv1.Subject{
		rbacv1.Subject{
			Kind: "Group",
			Name: "foo",
		},
		rbacv1.Subject{
			Kind: "Pod",
			Name: "me",
		},
		rbacv1.Subject{
			Kind: "Group",
			Name: "bar",
		},
	}

	items := []rbacv1.RoleBinding{
		rbacv1.RoleBinding{
			RoleRef:  roleRef1,
			Subjects: subjects,
		},
		rbacv1.RoleBinding{
			RoleRef:  roleRef2,
			Subjects: subjects,
		},
	}

	subjects = append(subjects, rbacv1.Subject{
		Kind: "Group",
		Name: "me",
	})

	items = append(items, rbacv1.RoleBinding{
		RoleRef:  roleRef3,
		Subjects: subjects,
	})

	roleBindingsReturn.Items = items

	roleRefsExp := []*rbacv1.RoleRef{
		&rbacv1.RoleRef{
			Name: "roleRef3",
			Kind: "Role",
		},
	}

	u.fakeClient.EXPECT().Rbac().Times(1).Return(u.fakeRbac)
	u.fakeRbac.EXPECT().RoleBindings(u.namespace).Times(1).Return(u.fakeRoleBindingsInterface)
	u.fakeRoleBindingsInterface.EXPECT().List(gomock.Any()).Return(roleBindingsReturn, nil)

	refs, err := u.RoleRefs()
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	if !reflect.DeepEqual(refs, roleRefsExp) {
		t.Errorf("unexpected role refs\nexp=%+v\ngot=%+v", roleRefsExp, refs)
	}
}

func TestGroup_GroupRefs_Successful_NoRef(t *testing.T) {
	u := newFakeGroup(t)
	defer u.ctrl.Finish()

	roleBindingsReturn := &rbacv1.RoleBindingList{}

	var roleRef1, roleRef2, roleRef3 rbacv1.RoleRef
	roleRef1.Name = "roleRef1"
	roleRef1.Kind = "Role"
	roleRef2.Name = "roleRef2"
	roleRef2.Kind = "Role"
	roleRef3.Name = "roleRef3"
	roleRef3.Kind = "Role"

	subjects := []rbacv1.Subject{
		rbacv1.Subject{
			Kind: "Pod",
			Name: "foo",
		},
		rbacv1.Subject{
			Kind: "Pod",
			Name: "me",
		},
		rbacv1.Subject{
			Kind: "Group",
			Name: "bee",
		},
		rbacv1.Subject{
			Kind: "Group",
			Name: "bar",
		},
	}

	items := []rbacv1.RoleBinding{
		rbacv1.RoleBinding{
			RoleRef:  roleRef1,
			Subjects: subjects,
		},
		rbacv1.RoleBinding{
			RoleRef:  roleRef2,
			Subjects: subjects,
		},
		rbacv1.RoleBinding{
			RoleRef:  roleRef3,
			Subjects: subjects,
		},
	}

	roleBindingsReturn.Items = items

	u.fakeClient.EXPECT().Rbac().Times(1).Return(u.fakeRbac)
	u.fakeRbac.EXPECT().RoleBindings(u.namespace).Times(1).Return(u.fakeRoleBindingsInterface)
	u.fakeRoleBindingsInterface.EXPECT().List(gomock.Any()).Return(roleBindingsReturn, nil)

	refs, err := u.RoleRefs()
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	if len(refs) != 0 {
		t.Errorf("unexpected role refs exp=nothing got=%+v", refs)
	}
}
