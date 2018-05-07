package subject_access_delegation

import (
	"testing"

	"github.com/golang/mock/gomock"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	authzv1alpha1 "github.com/joshvanl/k8s-subject-access-delegation/pkg/apis/authz/v1alpha1"
	"github.com/joshvanl/k8s-subject-access-delegation/pkg/subject_access_delegation/mocks"
)

var (
	originSubject = authzv1alpha1.OriginSubject{
		Name: "name",
		Kind: "ServiceAccount",
	}

	destinationSubject = authzv1alpha1.DestinationSubject{
		Name: "name2",
		Kind: "User",
	}

	sa = corev1.ServiceAccount{}

	noRoleBindings        = new(rbacv1.RoleBindingList)
	noClusterRoleBindings = new(rbacv1.ClusterRoleBindingList)
	someBindings          = []*rbacv1.RoleBinding{
		&rbacv1.RoleBinding{
			RoleRef: rbacv1.RoleRef{
				Name: "name",
				Kind: "ServiceAccount",
			},
			Subjects: []rbacv1.Subject{
				rbacv1.Subject{
					Kind: "ServiceAccount",
					Name: "name",
				},
				rbacv1.Subject{
					Kind: "ServiceAccount",
					Name: "name",
				},
			},
		},
		&rbacv1.RoleBinding{
			RoleRef: rbacv1.RoleRef{
				Name: "name",
				Kind: "ServiceAccount",
			},
			Subjects: []rbacv1.Subject{
				rbacv1.Subject{
					Kind: "ServiceAccount",
					Name: "name",
				},
			},
		},
	}
	someClusterBindings = []*rbacv1.ClusterRoleBinding{
		&rbacv1.ClusterRoleBinding{
			RoleRef: rbacv1.RoleRef{
				Name: "name",
				Kind: "ServiceAccount",
			},
			Subjects: []rbacv1.Subject{
				rbacv1.Subject{
					Kind: "ServiceAccount",
					Name: "name",
				},
				rbacv1.Subject{
					Kind: "ServiceAccount",
					Name: "name",
				},
			},
		},
		&rbacv1.ClusterRoleBinding{
			RoleRef: rbacv1.RoleRef{
				Name: "name",
			},
			Subjects: []rbacv1.Subject{
				rbacv1.Subject{
					Kind: "ServiceAccount",
					Name: "name",
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

type fakeSubejectAcessDelegation struct {
	real *SubjectAccessDelegation
	ctrl *gomock.Controller

	fakeController              *mocks.MockController
	fakeClient                  *mocks.MockInterface
	fakeCore                    *mocks.MockCoreV1Interface
	fakeServiceAccountInterface *mocks.MockServiceAccountInterface

	fakeRbac                        *mocks.MockRbacV1Interface
	fakeRoleBindingsInterface       *mocks.MockRoleBindingInterface
	fakeClusterRoleBindingInterface *mocks.MockClusterRoleBindingInterface

	fakeSAD *mocks.MockSubjectAccessDelegation
}

func newSubjectAccessDelegation(t *testing.T) *fakeSubejectAcessDelegation {
	s := &fakeSubejectAcessDelegation{
		ctrl: gomock.NewController(t),
	}

	sa.Name = "name"
	sa.Namespace = "fakeNamespace"

	s.fakeController = mocks.NewMockController(s.ctrl)
	s.fakeClient = mocks.NewMockInterface(s.ctrl)
	s.fakeCore = mocks.NewMockCoreV1Interface(s.ctrl)
	s.fakeServiceAccountInterface = mocks.NewMockServiceAccountInterface(s.ctrl)
	s.fakeRbac = mocks.NewMockRbacV1Interface(s.ctrl)
	s.fakeRoleBindingsInterface = mocks.NewMockRoleBindingInterface(s.ctrl)
	s.fakeClusterRoleBindingInterface = mocks.NewMockClusterRoleBindingInterface(s.ctrl)

	s.fakeSAD = mocks.NewMockSubjectAccessDelegation(s.ctrl)

	s.fakeClient.EXPECT().Core().Return(s.fakeCore)
	s.fakeCore.EXPECT().ServiceAccounts(gomock.Any()).Return(s.fakeServiceAccountInterface)

	s.real = New(s.fakeController, nil, nil, nil, s.fakeClient, nil, 0)
	s.real.sad = testingSAD()

	return s
}

func TestSAD_UpdateSadObject_Success(t *testing.T) {
	s := newSubjectAccessDelegation(t)
	sad := testingSAD()

	options := metav1.ListOptions{}

	s.fakeServiceAccountInterface.EXPECT().Get(gomock.Any(), gomock.Any()).Return(&sa, nil)

	s.fakeClient.EXPECT().Rbac().Return(s.fakeRbac).AnyTimes()
	s.fakeRbac.EXPECT().RoleBindings(gomock.Any()).Return(s.fakeRoleBindingsInterface)
	s.fakeRbac.EXPECT().ClusterRoleBindings().Return(s.fakeClusterRoleBindingInterface)
	s.fakeRoleBindingsInterface.EXPECT().List(options).Times(1).Return(noRoleBindings, nil)
	s.fakeClusterRoleBindingInterface.EXPECT().List(options).Times(1).Return(noClusterRoleBindings, nil)

	b, err := s.real.UpdateSadObject(sad)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	expectBool(b, false, t)
}

func TestSAD_UpdateSadObject_SuccessChangeDestination(t *testing.T) {
	s := newSubjectAccessDelegation(t)
	sad := testingSAD()

	sad.Spec.DestinationSubjects[0].Name = "foo"

	options := metav1.ListOptions{}

	s.fakeServiceAccountInterface.EXPECT().Get(gomock.Any(), gomock.Any()).Return(&sa, nil)

	s.fakeClient.EXPECT().Rbac().Return(s.fakeRbac).AnyTimes()
	s.fakeRbac.EXPECT().RoleBindings(gomock.Any()).Return(s.fakeRoleBindingsInterface)
	s.fakeRbac.EXPECT().ClusterRoleBindings().Return(s.fakeClusterRoleBindingInterface)
	s.fakeRoleBindingsInterface.EXPECT().List(options).Times(1).Return(noRoleBindings, nil)
	s.fakeClusterRoleBindingInterface.EXPECT().List(options).Times(1).Return(noClusterRoleBindings, nil)

	b, err := s.real.UpdateSadObject(sad)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	expectBool(b, true, t)
}

func TestSAD_UpdateSadObject_SuccessChangeTrigger(t *testing.T) {
	s := newSubjectAccessDelegation(t)
	sad := testingSAD()

	sad.Spec.EventTriggers = []authzv1alpha1.EventTrigger{
		authzv1alpha1.EventTrigger{
			Value: "value",
			Kind:  "kind",
		},
	}

	options := metav1.ListOptions{}

	s.fakeServiceAccountInterface.EXPECT().Get(gomock.Any(), gomock.Any()).Return(&sa, nil)

	s.fakeClient.EXPECT().Rbac().Return(s.fakeRbac).AnyTimes()
	s.fakeRbac.EXPECT().RoleBindings(gomock.Any()).Return(s.fakeRoleBindingsInterface)
	s.fakeRbac.EXPECT().ClusterRoleBindings().Return(s.fakeClusterRoleBindingInterface)
	s.fakeRoleBindingsInterface.EXPECT().List(options).Times(1).Return(noRoleBindings, nil)
	s.fakeClusterRoleBindingInterface.EXPECT().List(options).Times(1).Return(noClusterRoleBindings, nil)

	b, err := s.real.UpdateSadObject(sad)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	expectBool(b, true, t)
}

func testingSAD() *authzv1alpha1.SubjectAccessDelegation {
	return &authzv1alpha1.SubjectAccessDelegation{
		Spec: authzv1alpha1.SubjectAccessDelegationSpec{
			EventTriggers:       nil,
			DeletionTriggers:    nil,
			OriginSubject:       originSubject,
			DestinationSubjects: []authzv1alpha1.DestinationSubject{destinationSubject},
		},
	}
}

func expectBool(got, exp bool, t *testing.T) {
	if got != exp {
		t.Errorf("unexpected changes needed. got=%v exp=%v", got, exp)
	}
}
