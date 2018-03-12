package service_account

import (
	"errors"
	"fmt"

	rbacv1 "k8s.io/api/rbac/v1"
	"k8s.io/apimachinery/pkg/types"

	"github.com/joshvanl/k8s-subject-access-delegation/pkg/subject_access_delegation/role_binding"
	"github.com/joshvanl/k8s-subject-access-delegation/pkg/subject_access_delegation/utils"
)

func (s *ServiceAccount) addFuncRoleBinding(obj interface{}) {
	roleBinding, err := s.getRoleBindingObject(obj)
	if err != nil {
		s.log.Errorf("failed to decode newly added rolebinding: %v", err)
		return
	}

	// if we arn't referenced or we've seen this binding before, return
	if s.bindingContainsSubject(roleBinding) || s.seenUID(roleBinding.UID) {
		return
	}

	s.log.Infof("A new rolebinding referencing '%s' has been added. Updating SAD", s.Name())

	s.addUID(roleBinding.UID)
	s.bindings = append(s.bindings, roleBinding)

	binding := role_binding.NewFromRoleBinding(s.sad, roleBinding)

	if err := s.sad.AddRoleBinding(binding); err != nil {
		s.log.Errorf("Failed to add new rolebinding: %v", err)
	}
}

func (s *ServiceAccount) delFuncRoleBinding(obj interface{}) {
	roleBinding, err := s.getRoleBindingObject(obj)
	if err != nil {
		s.log.Errorf("failed to decode newly deleted rolebinding: %v", err)
		return
	}

	// if we arn't referenced or haven't seen this binding before, return
	if !s.bindingContainsSubject(roleBinding) || !s.seenUID(roleBinding.UID) {
		return
	}

	s.log.Infof("A RoleBinding referencing '%s' has been deleted. Updating SAD", s.Name())

	if !s.deleteRoleBinding(roleBinding.UID) {
		s.log.Errorf("Didn't find the deleted rolbinding in SAD references. Something has gone very wrong.")
	}

	binding := role_binding.NewFromRoleBinding(s.sad, roleBinding)

	if err := s.sad.DeleteRoleBinding(binding); err != nil {
		s.log.Errorf("Failed to delete rolebinding '%s': %v", binding.Name(), err)
	}
}

func (s *ServiceAccount) updateRoleBinding(oldObj, newObj interface{}) {
	oldRoleBinding, err := s.getRoleBindingObject(oldObj)
	if err != nil {
		s.log.Error(err)
		return
	}

	newRoleBinding, err := s.getRoleBindingObject(newObj)
	if err != nil {
		s.log.Error(err)
		return
	}

	// if we arn't referenced or haven't seen this binding before, return
	if !s.bindingContainsSubject(oldRoleBinding) || !s.seenUID(oldRoleBinding.UID) {
		return
	}

	s.log.Infof("A RoleBinding referencing '%s' has been updated. Updating SAD", s.Name())

	if !s.deleteRoleBinding(oldRoleBinding.UID) {
		s.log.Errorf("Didn't find the deleted rolbinding in SAD references. Something has gone very wrong.")
	}

	s.addUID(newRoleBinding.UID)
	s.bindings = append(s.bindings, newRoleBinding)

	oldBinding := role_binding.NewFromRoleBinding(s.sad, oldRoleBinding)
	newBinding := role_binding.NewFromRoleBinding(s.sad, newRoleBinding)

	if err := s.sad.UpdateRoleBinding(oldBinding, newBinding); err != nil {
		s.log.Errorf("error during updating SAD rolebindings: %v", err)
	}
}

func (s *ServiceAccount) addFuncClusterRoleBinding(obj interface{}) {
	clusterRoleBinding, err := s.getClusterRoleBindingObject(obj)
	if err != nil {
		s.log.Errorf("failed to decode newly added cluster rolebinding: %v", err)
	}

	// if we arn't referenced or have seen this binding before, return
	if !s.clusterBindingContainsSubject(clusterRoleBinding) || s.seenUID(clusterRoleBinding.UID) {
		return
	}

	s.log.Infof("A new cluster rolebinding referencing '%s' has been added. Updating SAD", s.Name())

	s.addUID(clusterRoleBinding.UID)
	s.clusterBindings = append(s.clusterBindings, clusterRoleBinding)

	binding := role_binding.NewFromClusterRoleBinding(s.sad, clusterRoleBinding)

	if err := s.sad.AddRoleBinding(binding); err != nil {
		s.log.Errorf("Failed to add new cluster rolebinding: %v", err)
	}
}

func (s *ServiceAccount) delFuncClusterRoleBinding(obj interface{}) {
	clusterRoleBinding, err := s.getClusterRoleBindingObject(obj)
	if err != nil {
		s.log.Errorf("failed to decode newly deleted cluster rolebinding: %v", err)
		return
	}

	// if we arn't referenced or haven't seen this binding before, return
	if !s.clusterBindingContainsSubject(clusterRoleBinding) || !s.seenUID(clusterRoleBinding.UID) {
		return
	}

	s.log.Infof("A Cluster RoleBinding referencing '%s' has been deleted. Updating SAD", s.Name())

	if !s.deleteClusterRoleBinding(clusterRoleBinding.UID) {
		s.log.Errorf("Didn't find the deleted cluster rolbinding in SAD references. Something has gone very wrong.")
	}

	binding := role_binding.NewFromClusterRoleBinding(s.sad, clusterRoleBinding)

	if err := s.sad.DeleteRoleBinding(binding); err != nil {
		s.log.Errorf("Failed to delete cluster rolebinding '%s': %v", binding.Name(), err)
	}
}

func (s *ServiceAccount) updateClusterRoleBinding(oldObj, newObj interface{}) {
	oldClusterRoleBinding, err := s.getClusterRoleBindingObject(oldObj)
	if err != nil {
		s.log.Error(err)
		return
	}

	newClusterRoleBinding, err := s.getClusterRoleBindingObject(newObj)
	if err != nil {
		s.log.Error(err)
		return
	}

	// if we arn't referenced or haven't seen this binding before, return
	if !s.clusterBindingContainsSubject(oldClusterRoleBinding) || !s.seenUID(oldClusterRoleBinding.UID) {
		return
	}

	s.log.Infof("A Cluster RoleBinding referencing '%s' has been updated. Updating SAD", s.Name())

	s.addUID(newClusterRoleBinding.UID)
	s.clusterBindings = append(s.clusterBindings, newClusterRoleBinding)

	oldBinding := role_binding.NewFromClusterRoleBinding(s.sad, oldClusterRoleBinding)
	newBinding := role_binding.NewFromClusterRoleBinding(s.sad, newClusterRoleBinding)

	if err := s.sad.UpdateRoleBinding(oldBinding, newBinding); err != nil {
		s.log.Errorf("error during updating SAD cluster rolebindings: %v", err)
	}
}

func (s *ServiceAccount) getRoleBindingObject(obj interface{}) (*rbacv1.RoleBinding, error) {
	binding, err := utils.GetRoleBindingObject(obj)
	if err != nil {
		return nil, fmt.Errorf("failed to get rolebinding object: %v", err)
	}
	if binding == nil {
		return nil, errors.New("failed to get rolebinding, received nil object")
	}

	return binding, nil
}

func (s *ServiceAccount) getClusterRoleBindingObject(obj interface{}) (*rbacv1.ClusterRoleBinding, error) {
	binding, err := utils.GetClusterRoleBindingObject(obj)
	if err != nil {
		return nil, fmt.Errorf("failed to get cluster rolebinding object: %v", err)
	}
	if binding == nil {
		return nil, errors.New("failed to get cluster rolebinding, received nil object")
	}

	return binding, nil
}

func (s *ServiceAccount) deleteRoleBinding(uid types.UID) bool {
	for i, binding := range s.bindings {
		if binding.UID == uid {

			copy(s.bindings[i:], s.bindings[i+1:])
			s.bindings[len(s.bindings)-1] = nil
			s.bindings = s.bindings[:len(s.bindings)-1]

			return true
		}
	}

	return false
}

func (s *ServiceAccount) deleteClusterRoleBinding(uid types.UID) bool {
	for i, binding := range s.clusterBindings {
		if binding.UID == uid {

			copy(s.clusterBindings[i:], s.clusterBindings[i+1:])
			s.clusterBindings[len(s.clusterBindings)-1] = nil
			s.clusterBindings = s.clusterBindings[:len(s.clusterBindings)-1]

			return true
		}
	}

	return false
}

func (s *ServiceAccount) seenUID(uid types.UID) bool {
	b, ok := s.uids[uid]
	if !ok {
		return false
	}

	return b
}

func (s *ServiceAccount) addUID(uid types.UID) {
	s.uids[uid] = true
}
