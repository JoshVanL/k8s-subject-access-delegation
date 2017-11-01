package origin_subject

import (
	"github.com/joshvanl/k8s-subject-access-delegation/pkg/subject_access_delegation/interfaces"
	"github.com/joshvanl/k8s-subject-access-delegation/pkg/subject_access_delegation/origin_subject/role"
)

const RoleKind = "Role"
const ServiceAccountKind = "ServiceAccount"

func New(sad interfaces.SubjectAccessDelegation) (interfaces.OriginSubject, error) {
	var originSubject interfaces.OriginSubject

	if sad.Kind() == RoleKind {
		originSubject = role.New(sad)
		return originSubject, nil
	}

	return nil, nil
}
