package origin_subject

import (
	"fmt"

	rbacv1 "k8s.io/api/rbac/v1"

	"github.com/joshvanl/k8s-subject-access-delegation/pkg/subject_access_delegation/interfaces"
	"github.com/joshvanl/k8s-subject-access-delegation/pkg/subject_access_delegation/origin_subject/role"
	"github.com/joshvanl/k8s-subject-access-delegation/pkg/subject_access_delegation/origin_subject/service_account"
	"github.com/joshvanl/k8s-subject-access-delegation/pkg/subject_access_delegation/origin_subject/user"
)

const RoleKind = "Role"

func New(sad interfaces.SubjectAccessDelegation) (interfaces.OriginSubject, error) {
	var originSubject interfaces.OriginSubject

	switch sad.OriginKind() {
	case RoleKind:
		originSubject = role.New(sad)
		return originSubject, nil

	case rbacv1.ServiceAccountKind:
		originSubject = service_account.New(sad)
		return originSubject, nil

	case rbacv1.UserKind:
		originSubject = user.New(sad)
		return originSubject, nil
	}

	return nil, fmt.Errorf("Subject Accesss Deletgation does not support Origin Subject Kind '%s'", sad.OriginKind())
}
