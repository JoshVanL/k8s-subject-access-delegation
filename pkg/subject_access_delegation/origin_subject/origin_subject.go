package origin_subject

import (
	"fmt"

	"github.com/joshvanl/k8s-subject-access-delegation/pkg/subject_access_delegation/interfaces"
	"github.com/joshvanl/k8s-subject-access-delegation/pkg/subject_access_delegation/origin_subject/role"
	"github.com/joshvanl/k8s-subject-access-delegation/pkg/subject_access_delegation/origin_subject/service_account"
)

const RoleKind = "Role"
const ServiceAccountKind = "ServiceAccount"

func New(sad interfaces.SubjectAccessDelegation) (interfaces.OriginSubject, error) {
	var originSubject interfaces.OriginSubject

	switch sad.OriginKind() {
	case RoleKind:
		originSubject = role.New(sad)
		return originSubject, nil

	case ServiceAccountKind:
		originSubject = service_account.New(sad)
		return originSubject, nil
	}

	return nil, fmt.Errorf("Subject Accesss Deletgation does not support Origin Subject Kind '%s'", sad.OriginKind())
}
