package service_account

import (
	"github.com/joshvanl/k8s-subject-access-delegation/pkg/subject_access_delegation/utils"
)

// TODO: if this is active then this needs to chaneg the permissions on the
// destination subject
func (s *ServiceAccount) addFuncRoleBinding(obj interface{}) {
	binding, err := utils.GetRolebindingObject(obj)
	if err != nil {
		s.log.Errorf("failed to get added rolebinding object: %v", err)
		return
	}
	if binding == nil {
		s.log.Error("failed to get clusterbinding, received nil object")
	}

	if !s.bindingContainsSubject(binding) || !s.seenUID(binding.UID) {
		return
	}

	s.log.Infof("A new rolebinding referencing '%s' has been added. Updating SAD", s.Name())

	s.addUID(binding.UID)
	s.bindings = append(s.bindings, binding)
}

func (s *ServiceAccount) delFuncRoleBinding(obj interface{}) {

}

func (s *ServiceAccount) addFuncClusterRoleBinding(obj interface{}) {
	binding, err := utils.GetClusterRolebindingObject(obj)
	if err != nil {
		s.log.Errorf("failed to get added cluster rolebinding object: %v", err)
		return
	}
	if binding == nil {
		s.log.Error("failed to get cluster rolebinding, received nil object")
	}

	if !s.clusterBindingContainsSubject(binding) || !s.seenUID(binding.UID) {
		return
	}

	s.log.Infof("A new cluster rolebinding referencing '%s' has been added. Updating SAD", s.Name())

	s.addUID(binding.UID)
	s.clusterBindings = append(s.clusterBindings, binding)

}
