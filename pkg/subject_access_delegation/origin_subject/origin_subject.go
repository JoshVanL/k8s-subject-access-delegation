package origin_subject

import (
	"fmt"

	rbacv1 "k8s.io/api/rbac/v1"

	"github.com/joshvanl/k8s-subject-access-delegation/pkg/interfaces"
	//"github.com/joshvanl/k8s-subject-access-delegation/pkg/subject_access_delegation/origin_subject/group"
	"github.com/joshvanl/k8s-subject-access-delegation/pkg/subject_access_delegation/origin_subject/cluster_role"
	"github.com/joshvanl/k8s-subject-access-delegation/pkg/subject_access_delegation/origin_subject/role"
	"github.com/joshvanl/k8s-subject-access-delegation/pkg/subject_access_delegation/origin_subject/service_account"
	//"github.com/joshvanl/k8s-subject-access-delegation/pkg/subject_access_delegation/origin_subject/user"
)

const RoleKind = "Role"
const ClusterRoleType = "ClusterRole"

func New(sad interfaces.SubjectAccessDelegation, name, kind string) (interfaces.OriginSubject, error) {
	var originSubject interfaces.OriginSubject

	switch kind {
	case role.RoleKind:
		originSubject = role.New(sad, name)
		return originSubject, nil

	case cluster_role.ClusterRoleKind:
		originSubject = cluster_role.New(sad, name)
		return originSubject, nil

	case rbacv1.ServiceAccountKind:
		originSubject = service_account.New(sad, name)
		return originSubject, nil

		//	case rbacv1.UserKind:
		//		originSubject = user.New(sad, name)
		//		return originSubject, nil

		//	case rbacv1.GroupKind:
		//		originSubject = group.New(sad, name)
		//		return originSubject, nil

	default:
		return nil, fmt.Errorf("Subject Accesss Deletgation does not support Origin Subject Kind '%s'", kind)
	}
}
