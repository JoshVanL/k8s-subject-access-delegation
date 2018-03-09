package service_account

import (
	"errors"
	"fmt"

	rbacv1 "k8s.io/api/rbac/v1"
	"k8s.io/apimachinery/pkg/types"

	"github.com/joshvanl/k8s-subject-access-delegation/pkg/subject_access_delegation/utils"
)

// TODO: if this is active then this needs to change the permissions on the
// destination subject
func (s *ServiceAccount) addFuncRoleBinding(obj interface{}) {
	binding, err := s.getRoleBindingObject(obj)
	if err != nil {
		s.log.Errorf("failed to decode newly added rolebinding: %v", err)
		return
	}

	// if we arn't referenced or we've seen this binding before, return
	if !s.bindingContainsSubject(binding) || s.seenUID(binding.UID) {
		return
	}

	s.log.Infof("A new rolebinding referencing '%s' has been added. Updating SAD", s.Name())

	s.addUID(binding.UID)
	s.bindings = append(s.bindings, binding)
}

// TODO: We need to tell the controller to update it's referenced rolebindings
func (s *ServiceAccount) delFuncRoleBinding(obj interface{}) {
	binding, err := s.getRoleBindingObject(obj)
	if err != nil {
		s.log.Errorf("failed to decode newly deleted rolebinding: %v", err)
		return
	}

	// if we arn't referenced or haven't seen this binding before, return
	if !s.bindingContainsSubject(binding) || !s.seenUID(binding.UID) {
		return
	}

	s.log.Infof("A RoleBinding referencing '%s' has been deleted. Updating SAD", s.Name())

	if !s.deleteRoleBinding(binding.UID) {
		s.log.Errorf("Didn't find the deleted rolbinding in SAD references. Something has gone very wrong.")
	}
}

// TODO: we need to tell the controller to update it's replicated binding to
// newObj
func (s *ServiceAccount) updateRoleBindingOject(oldObj, newObj interface{}) {
	binding, err := s.getRoleBindingObject(oldObj)
	if err != nil {
		s.log.Error(err)
		return
	}

	// if we arn't referenced or haven't seen this binding before, return
	if !s.bindingContainsSubject(binding) || !s.seenUID(binding.UID) {
		return
	}

	s.log.Infof("A RoleBinding referencing '%s' has been updated. Updating SAD", s.Name())
}

// TODO: if this is active then this needs to change the permissions on the
// destination subject
func (s *ServiceAccount) addFuncClusterRoleBinding(obj interface{}) {
	binding, err := s.getClusterRoleBindingObject(obj)
	if err != nil {
		s.log.Errorf("failed to decode newly added cluster rolebinding: %v", err)
	}

	// if we arn't referenced or have seen this binding before, return
	if !s.clusterBindingContainsSubject(binding) || s.seenUID(binding.UID) {
		return
	}

	s.log.Infof("A new cluster rolebinding referencing '%s' has been added. Updating SAD", s.Name())

	s.addUID(binding.UID)
	s.clusterBindings = append(s.clusterBindings, binding)
}

// TODO: We need to tell the controller to update it's referenced rolebindings
func (s *ServiceAccount) delFuncClusterRoleBinding(obj interface{}) {
	binding, err := s.getClusterRoleBindingObject(obj)
	if err != nil {
		s.log.Errorf("failed to decode newly deleted cluster rolebinding: %v", err)
		return
	}

	// if we arn't referenced or haven't seen this binding before, return
	if !s.clusterBindingContainsSubject(binding) || !s.seenUID(binding.UID) {
		return
	}

	s.log.Infof("A Cluster RoleBinding referencing '%s' has been deleted. Updating SAD", s.Name())

	if !s.deleteClusterRoleBinding(binding.UID) {
		s.log.Errorf("Didn't find the deleted cluster rolbinding in SAD references. Something has gone very wrong.")
	}
}

// TODO: we need to tell the controller to update it's replicated binding to
// newObj
func (s *ServiceAccount) updateClusterRoleBindingOject(oldObj, newObj interface{}) {
	binding, err := s.getClusterRoleBindingObject(oldObj)
	if err != nil {
		s.log.Error(err)
		return
	}

	// if we arn't referenced or haven't seen this binding before, return
	if !s.clusterBindingContainsSubject(binding) || !s.seenUID(binding.UID) {
		return
	}

	s.log.Infof("A Cluster RoleBinding referencing '%s' has been updated. Updating SAD", s.Name())
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
