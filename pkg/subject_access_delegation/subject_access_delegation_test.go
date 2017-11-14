package subject_access_delegation

import (
	"fmt"
	"testing"
	//"time"

	"github.com/golang/mock/gomock"
	"github.com/sirupsen/logrus"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	authzv1alpha1 "github.com/joshvanl/k8s-subject-access-delegation/pkg/apis/authz/v1alpha1"
	"github.com/joshvanl/k8s-subject-access-delegation/pkg/subject_access_delegation/mocks"
)

var (
	timeTrigger1 = authzv1alpha1.EventTrigger{
		Kind:  "Time",
		Value: "1s",
	}
	timeTrigger2 = authzv1alpha1.EventTrigger{
		Kind:  "Time",
		Value: "1s",
	}

	originSubjectRole = authzv1alpha1.OriginSubject{
		Kind: "Role",
		Name: "RoleRef",
	}
	originSubjectSA = authzv1alpha1.OriginSubject{
		Kind: "ServiceAccount",
		Name: "TestServiceAccount",
	}
	originSubjectUser = authzv1alpha1.OriginSubject{
		Kind: "User",
		Name: "TestUser",
	}

	destinationSubjects = []authzv1alpha1.DestinationSubject{
		authzv1alpha1.DestinationSubject{
			Kind: "Pod",
			Name: "TargetPod",
		},
		authzv1alpha1.DestinationSubject{
			Kind: "ServiceAccount",
			Name: "TargetServiceAccount",
		},
		authzv1alpha1.DestinationSubject{
			Kind: "User",
			Name: "TargetUser",
		},
	}
	bindingSubjects = []rbacv1.Subject{
		rbacv1.Subject{
			Kind: "Pod",
			Name: "TargetPod",
		},
		rbacv1.Subject{
			Kind: "ServiceAccount",
			Name: "TargetServiceAccount",
		},
		rbacv1.Subject{
			Kind: "User",
			Name: "TargetUser",
		},
	}

	roleRef = rbacv1.RoleRef{
		Kind: "Role",
		Name: "RoleRef",
	}
)

type fakeSubjectAccessDelegation struct {
	*SubjectAccessDelegation
	ctrl *gomock.Controller

	fakeClient         *mocks.MockInterface
	fakeRbac           *mocks.MockRbacV1Interface
	fakeRoleInterface  *mocks.MockRoleInterface
	fakeRoleBindingsIn *mocks.MockRoleBindingInterface
	fakeCore           *mocks.MockCoreV1Interface
	fakePodInterface   *mocks.MockPodInterface
	fakeSAInterface    *mocks.MockServiceAccountInterface
}

func newFakeSAD(t *testing.T) *fakeSubjectAccessDelegation {
	s := &fakeSubjectAccessDelegation{
		ctrl: gomock.NewController(t),
		SubjectAccessDelegation: &SubjectAccessDelegation{
			sad: &authzv1alpha1.SubjectAccessDelegation{},
			log: logrus.NewEntry(logrus.New()),
		},
	}

	s.sad.Name = "sadName"
	s.sad.Namespace = "sadNamespace"

	s.fakeClient = mocks.NewMockInterface(s.ctrl)
	s.fakeRbac = mocks.NewMockRbacV1Interface(s.ctrl)
	s.fakeRoleInterface = mocks.NewMockRoleInterface(s.ctrl)
	s.fakeCore = mocks.NewMockCoreV1Interface(s.ctrl)
	s.fakePodInterface = mocks.NewMockPodInterface(s.ctrl)
	s.SubjectAccessDelegation.client = s.fakeClient
	s.fakeSAInterface = mocks.NewMockServiceAccountInterface(s.ctrl)
	s.fakeRoleBindingsIn = mocks.NewMockRoleBindingInterface(s.ctrl)

	s.fakeClient.EXPECT().Rbac().AnyTimes().Return(s.fakeRbac)
	s.fakeClient.EXPECT().Core().AnyTimes().Return(s.fakeCore)
	s.fakeRbac.EXPECT().RoleBindings(s.sad.Namespace).AnyTimes().Return(s.fakeRoleBindingsIn)
	s.fakeRbac.EXPECT().Roles(s.sad.Namespace).AnyTimes().Return(s.fakeRoleInterface)
	s.fakeCore.EXPECT().Pods(s.sad.Namespace).AnyTimes().Return(s.fakePodInterface)
	s.fakeCore.EXPECT().ServiceAccounts(s.sad.Namespace).AnyTimes().Return(s.fakeSAInterface)

	return s
}

func returnPod() *corev1.Pod {
	returnPod := &corev1.Pod{}
	returnPod.Name = "TargetPod"
	returnPod.Kind = "Pod"

	return returnPod
}

func returnServiceAccount() *corev1.ServiceAccount {
	returnServiceAccount := &corev1.ServiceAccount{}
	returnServiceAccount.Name = "TargetServiceAccount"
	returnServiceAccount.Kind = "ServiceAccount"

	return returnServiceAccount
}

func returnUser() *corev1.ServiceAccount {
	returnUser := &corev1.ServiceAccount{}
	returnUser.Name = "TargetUser"
	returnUser.Kind = "User"

	return returnUser
}

func TestSAD_Delegate_Nill_NoRepeat(t *testing.T) {
	s := newFakeSAD(t)
	defer s.ctrl.Finish()

	closed, err := s.Delegate()
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if closed {
		t.Error("SAD delegation returned closed, expected false")
	}
}

func TestSAD_Delegate_Nill_Repeat_NoTime_NoOrigin(t *testing.T) {
	s := newFakeSAD(t)
	defer s.ctrl.Finish()

	s.SubjectAccessDelegation.sad.Spec.Repeat = 3

	closed, err := s.Delegate()
	if err == nil {
		t.Errorf("expected error but returned none")
	}
	if closed {
		t.Error("SAD delegation returned closed, expected false")
	}
}

func TestSAD_Delegate_Nill_Repeat_Time_NoOrigin(t *testing.T) {
	s := newFakeSAD(t)
	defer s.ctrl.Finish()

	s.SubjectAccessDelegation.sad.Spec.Repeat = 3

	timeTrigger := []authzv1alpha1.EventTrigger{
		authzv1alpha1.EventTrigger{
			Kind:  "Time",
			Value: "10h 20m",
		},
	}
	s.SubjectAccessDelegation.sad.Spec.EventTriggers = timeTrigger
	s.sad.Spec.DeletionTime = "5m"

	closed, err := s.Delegate()
	if err == nil {
		t.Errorf("expected error but returned none")
	}
	if closed {
		t.Error("SAD delegation returned closed, expected false")
	}
}

func TestSAD_Delegate_Nill_Repeat_NoTime_Origin(t *testing.T) {
	s := newFakeSAD(t)
	defer s.ctrl.Finish()

	s.SubjectAccessDelegation.sad.Spec.Repeat = 3

	originSubject := authzv1alpha1.OriginSubject{
		Kind: "Role",
		Name: "TestRole",
	}
	s.SubjectAccessDelegation.sad.Spec.OriginSubject = originSubject

	s.fakeRoleInterface.EXPECT().Get(s.sad.Spec.OriginSubject.Name, metav1.GetOptions{}).Times(1).Return(&rbacv1.Role{}, nil)

	closed, err := s.Delegate()
	if err == nil {
		t.Errorf("expected error but returned none")
	}
	if closed {
		t.Error("SAD delegation returned closed, expected false")
	}
}

func TestSAD_Delegate_Nill_Repeat_Time_Origin(t *testing.T) {
	s := newFakeSAD(t)
	defer s.ctrl.Finish()

	s.SubjectAccessDelegation.sad.Spec.Repeat = 3

	s.SubjectAccessDelegation.sad.Spec.EventTriggers = []authzv1alpha1.EventTrigger{timeTrigger1}
	s.sad.Spec.DeletionTime = "1s"

	originSubject := authzv1alpha1.OriginSubject{
		Kind: "Role",
		Name: "TestRole",
	}
	s.SubjectAccessDelegation.sad.Spec.OriginSubject = originSubject

	s.fakeRoleInterface.EXPECT().Get(s.sad.Spec.OriginSubject.Name, metav1.GetOptions{}).Times(1).Return(&rbacv1.Role{}, nil)

	closed, err := s.Delegate()
	if err == nil {
		t.Errorf("expected error but returned none")
	}
	if closed {
		t.Error("SAD delegation returned closed, expected false")
	}
}

func TestSAD_Delegate_Nill_Repeat_Time_OriginRole_Successful(t *testing.T) {
	s := newFakeSAD(t)
	defer s.ctrl.Finish()

	repeat := 3
	s.SubjectAccessDelegation.sad.Spec.Repeat = repeat
	s.SubjectAccessDelegation.sad.Spec.EventTriggers = []authzv1alpha1.EventTrigger{timeTrigger1}
	s.sad.Spec.DeletionTime = "1s"
	s.SubjectAccessDelegation.sad.Spec.OriginSubject = originSubjectRole
	s.sad.Spec.DestinationSubjects = destinationSubjects

	createBinding := &rbacv1.RoleBinding{}
	createBinding.Name = fmt.Sprintf("%s-%s-%s", s.sad.Name, s.sad.Namespace, roleRef.Name)
	createBinding.Subjects = bindingSubjects
	createBinding.Namespace = s.sad.Namespace
	createBinding.RoleRef = roleRef

	//timestamp := metav1.Time{
	//	Time: time.Now(),
	//}
	//createBinding.CreationTimestamp = timestamp

	s.fakeRoleInterface.EXPECT().Get(s.sad.Spec.OriginSubject.Name, metav1.GetOptions{}).Times(repeat).Return(&rbacv1.Role{}, nil)
	s.fakePodInterface.EXPECT().Get("TargetPod", gomock.Any()).Times(repeat).Return(returnPod(), nil)
	s.fakeSAInterface.EXPECT().Get("TargetServiceAccount", gomock.Any()).Times(repeat).Return(returnServiceAccount(), nil)
	s.fakeSAInterface.EXPECT().Get("TargetUser", gomock.Any()).Times(repeat).Return(returnUser(), nil)
	s.fakeRoleBindingsIn.EXPECT().Create(createBinding).Times(repeat).Return(nil, nil)
	s.fakeRoleBindingsIn.EXPECT().Delete(createBinding.Name, gomock.Any()).Times(repeat).Return(nil)

	closed, err := s.Delegate()
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if closed {
		t.Error("SAD delegation returned closed, expected false")
	}
}
