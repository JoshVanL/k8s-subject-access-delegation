package utils

import (
	"testing"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"

	authzv1alpha1 "github.com/joshvanl/k8s-subject-access-delegation/pkg/apis/authz/v1alpha1"
)

func TestPod_Pass(t *testing.T) {
	_, err := GetPodObject(new(corev1.Pod))
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestPod_Fail(t *testing.T) {
	_, err := GetPodObject(new(corev1.ServiceAccount))
	if err == nil {
		t.Errorf("expected error, got=none")
	}
}

func TestNode_Pass(t *testing.T) {
	_, err := GetNodeObject(new(corev1.Node))
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestNode_Failt(t *testing.T) {
	_, err := GetNodeObject(new(corev1.ServiceAccount))
	if err == nil {
		t.Errorf("expected error, got=none")
	}
}

func TestSecret_Pass(t *testing.T) {
	_, err := GetSecretObject(new(corev1.Secret))
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestSecret_Failt(t *testing.T) {
	_, err := GetSecretObject(new(corev1.ServiceAccount))
	if err == nil {
		t.Errorf("expected error, got=none")
	}
}

func TestService_Pass(t *testing.T) {
	_, err := GetServiceObject(new(corev1.Service))
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestService_Fail(t *testing.T) {
	_, err := GetServiceObject(new(corev1.ServiceAccount))
	if err == nil {
		t.Errorf("expected error, got=none")
	}
}

func TestEndPoints_Pass(t *testing.T) {
	_, err := GetEndPointsObject(new(corev1.Endpoints))
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestEndPoints_Fail(t *testing.T) {
	_, err := GetEndPointsObject(new(corev1.ServiceAccount))
	if err == nil {
		t.Errorf("expected error, got=none")
	}
}

func TestDeployments_Pass(t *testing.T) {
	_, err := GetDeploymentObject(new(appsv1.Deployment))
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestDeployments_Fail(t *testing.T) {
	_, err := GetDeploymentObject(new(corev1.ServiceAccount))
	if err == nil {
		t.Errorf("expected error, got=none")
	}
}

func TestRoleBinding_Pass(t *testing.T) {
	_, err := GetRoleBindingObject(new(rbacv1.RoleBinding))
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestRoleBinding_Pass_Fail(t *testing.T) {
	_, err := GetRoleBindingObject(new(corev1.ServiceAccount))
	if err == nil {
		t.Errorf("expected error, got=none")
	}
}

func TestClusterRoleBinding_Pass(t *testing.T) {
	_, err := GetClusterRoleBindingObject(new(rbacv1.ClusterRoleBinding))
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestClusterRoleBinding_Fail(t *testing.T) {
	_, err := GetClusterRoleBindingObject(new(corev1.ServiceAccount))
	if err == nil {
		t.Errorf("expected error, got=none")
	}
}

func TestSAD_Pass(t *testing.T) {
	_, err := GetSubjectAccessDelegationObject(new(authzv1alpha1.SubjectAccessDelegation))
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestSAD_Fail(t *testing.T) {
	_, err := GetClusterRoleBindingObject(new(corev1.ServiceAccount))
	if err == nil {
		t.Errorf("expected error, got=none")
	}
}
