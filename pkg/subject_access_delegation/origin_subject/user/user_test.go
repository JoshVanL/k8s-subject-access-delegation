package user

import (
	"errors"
	"reflect"
	"testing"

	"github.com/golang/mock/gomock"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"

	"github.com/joshvanl/k8s-subject-access-delegation/pkg/subject_access_delegation/mocks"
)

type fakeUser struct {
	*User
	ctrl *gomock.Controller

	fakeClient                  *mocks.MockInterface
	fakeRbac                    *mocks.MockRbacV1Interface
	fakeRoleBindingsInterface   *mocks.MockRoleBindingInterface
	fakeCore                    *mocks.MockCoreV1Interface
	fakeServiceAccountInterface *mocks.MockServiceAccountInterface
}

func newFakeUser(t *testing.T) *fakeUser {
	u := &fakeUser{
		ctrl: gomock.NewController(t),
		User: &User{
			namespace: "fakeNamespace",
			name:      "me",
		},
	}

	u.fakeClient = mocks.NewMockInterface(u.ctrl)
	u.fakeRbac = mocks.NewMockRbacV1Interface(u.ctrl)
	u.fakeRoleBindingsInterface = mocks.NewMockRoleBindingInterface(u.ctrl)
	u.fakeCore = mocks.NewMockCoreV1Interface(u.ctrl)
	u.fakeServiceAccountInterface = mocks.NewMockServiceAccountInterface(u.ctrl)

	u.User.client = u.fakeClient

	return u
}

func TestUser_ResolveDestination_Nil(t *testing.T) {
	u := newFakeUser(t)
	defer u.ctrl.Finish()

	u.fakeClient.EXPECT().Core().Times(1).Return(u.fakeCore)
	u.fakeCore.EXPECT().ServiceAccounts(u.User.namespace).Times(1).Return(u.fakeServiceAccountInterface)
	u.fakeServiceAccountInterface.EXPECT().Get(u.User.name, gomock.Any()).Times(1).Return(nil, nil)

	if err := u.ResolveOrigin(); err == nil {
		t.Error("expected error but got none - rolebindings is nil")
	}
}

func TestUser_ResolveOrigin_Error(t *testing.T) {
	u := newFakeUser(t)
	defer u.ctrl.Finish()

	u.fakeClient.EXPECT().Core().Times(1).Return(u.fakeCore)
	u.fakeCore.EXPECT().ServiceAccounts(u.User.namespace).Times(1).Return(u.fakeServiceAccountInterface)
	u.fakeServiceAccountInterface.EXPECT().Get(u.User.name, gomock.Any()).Times(1).Return(&corev1.ServiceAccount{}, errors.New("this is an error"))

	if err := u.ResolveOrigin(); err == nil {
		t.Error("expected error but got none - returned error")
	}
}

func TestUser_ResolveOrigin_Successful(t *testing.T) {
	u := newFakeUser(t)
	defer u.ctrl.Finish()

	u.fakeClient.EXPECT().Core().Times(1).Return(u.fakeCore)
	u.fakeCore.EXPECT().ServiceAccounts(u.User.namespace).Times(1).Return(u.fakeServiceAccountInterface)
	u.fakeServiceAccountInterface.EXPECT().Get(u.User.name, gomock.Any()).Times(1).Return(&corev1.ServiceAccount{}, nil)

	if err := u.ResolveOrigin(); err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestUser_RoleRefs_Nil(t *testing.T) {
	u := newFakeUser(t)
	defer u.ctrl.Finish()

	u.fakeClient.EXPECT().Rbac().Times(1).Return(u.fakeRbac)
	u.fakeRbac.EXPECT().RoleBindings(u.namespace).Times(1).Return(u.fakeRoleBindingsInterface)
	u.fakeRoleBindingsInterface.EXPECT().List(gomock.Any()).Return(nil, nil)

	_, err := u.RoleRefs()
	if err == nil {
		t.Error("expected error but got none - returned nil")
	}
}

func TestUser_RoleRefs_Error(t *testing.T) {
	u := newFakeUser(t)
	defer u.ctrl.Finish()

	u.fakeClient.EXPECT().Rbac().Times(1).Return(u.fakeRbac)
	u.fakeRbac.EXPECT().RoleBindings(u.namespace).Times(1).Return(u.fakeRoleBindingsInterface)
	u.fakeRoleBindingsInterface.EXPECT().List(gomock.Any()).Return(nil, errors.New("this is an error"))

	_, err := u.RoleRefs()
	if err == nil {
		t.Error("expected error but got none - returned error")
	}
}

func TestUser_RoleRefs_Successful_None(t *testing.T) {
	u := newFakeUser(t)
	defer u.ctrl.Finish()

	roleRefsReturn := &rbacv1.RoleBindingList{}

	u.fakeClient.EXPECT().Rbac().Times(1).Return(u.fakeRbac)
	u.fakeRbac.EXPECT().RoleBindings(u.namespace).Times(1).Return(u.fakeRoleBindingsInterface)
	u.fakeRoleBindingsInterface.EXPECT().List(gomock.Any()).Return(roleRefsReturn, nil)

	refs, err := u.RoleRefs()
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	if len(refs) != 0 {
		t.Errorf("unexpected role refs exp=nothing got=%+v", refs)
	}
}

func TestUser_RoleRefs_Successful_All(t *testing.T) {
	u := newFakeUser(t)
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
			Kind: "User",
			Name: "me",
		},
		rbacv1.Subject{
			Kind: "Pod",
			Name: "me",
		},
		rbacv1.Subject{
			Kind: "User",
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

func TestUser_RoleRefs_Successful_Some(t *testing.T) {
	u := newFakeUser(t)
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
			Kind: "User",
			Name: "foo",
		},
		rbacv1.Subject{
			Kind: "Pod",
			Name: "me",
		},
		rbacv1.Subject{
			Kind: "User",
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
		Kind: "User",
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

func TestUser_RoleRefs_Successful_NoRef(t *testing.T) {
	u := newFakeUser(t)
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
			Kind: "User",
			Name: "bee",
		},
		rbacv1.Subject{
			Kind: "User",
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
