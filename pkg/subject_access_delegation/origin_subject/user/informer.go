package user

import (
	"errors"
	"fmt"

	rbacv1 "k8s.io/api/rbac/v1"
	"k8s.io/apimachinery/pkg/types"

	"github.com/joshvanl/k8s-subject-access-delegation/pkg/subject_access_delegation/role_binding"
	"github.com/joshvanl/k8s-subject-access-delegation/pkg/subject_access_delegation/utils"
)

func (u *User) addFuncRoleBinding(obj interface{}) {
	roleBinding, err := u.getRoleBindingObject(obj)
	if err != nil {
		u.log.Errorf("failed to decode newly added rolebinding: %v", err)
		return
	}

	// if we arn't referenced or we've seen this binding before, return
	if u.bindingContainsSubject(roleBinding) || u.seenUID(roleBinding.UID) {
		return
	}

	u.log.Infof("A new rolebinding referencing '%s' has been added. Updating SAD", u.Name())

	u.addUID(roleBinding.UID)
	u.bindings = append(u.bindings, roleBinding)

	binding := role_binding.NewFromRoleBinding(u.sad, roleBinding)

	if err := u.sad.AddRoleBinding(binding); err != nil {
		u.log.Errorf("Failed to add new rolebinding: %v", err)
	}
}

// TODO: We need to tell the controller to update it's referenced rolebindings
func (u *User) delFuncRoleBinding(obj interface{}) {
	roleBinding, err := u.getRoleBindingObject(obj)
	if err != nil {
		u.log.Errorf("failed to decode newly deleted rolebinding: %v", err)
		return
	}

	// if we arn't referenced or haven't seen this binding before, return
	if !u.bindingContainsSubject(roleBinding) || !u.seenUID(roleBinding.UID) {
		return
	}

	u.log.Infof("A RoleBinding referencing '%s' has been deleted. Updating SAD", u.Name())

	if !u.deleteRoleBinding(roleBinding.UID) {
		u.log.Errorf("Didn't find the deleted rolbinding in SAD references. Something has gone very wrong.")
	}

	binding := role_binding.NewFromRoleBinding(u.sad, roleBinding)

	if !u.sad.DeleteRoleBinding(binding) {
		u.log.Errorf("Failed to delete rolebinding '%s'. It did not exist.", binding.Name)
	}
}

// TODO: we need to tell the controller to update it's replicated binding to
// newObj
func (u *User) updateRoleBindingOject(oldObj, newObj interface{}) {
	oldRoleBinding, err := u.getRoleBindingObject(oldObj)
	if err != nil {
		u.log.Error(err)
		return
	}

	newRoleBinding, err := u.getRoleBindingObject(newObj)
	if err != nil {
		u.log.Error(err)
		return
	}

	// if we arn't referenced or haven't seen this binding before, return
	if !u.bindingContainsSubject(oldRoleBinding) || !u.seenUID(oldRoleBinding.UID) {
		return
	}

	u.log.Infof("A RoleBinding referencing '%s' has been updated. Updating SAD", u.Name())

	if !u.deleteRoleBinding(oldRoleBinding.UID) {
		u.log.Errorf("Didn't find the deleted rolbinding in SAD references. Something has gone very wrong.")
	}

	u.addUID(newRoleBinding.UID)
	u.bindings = append(u.bindings, newRoleBinding)

	oldBinding := role_binding.NewFromRoleBinding(u.sad, oldRoleBinding)
	newBinding := role_binding.NewFromRoleBinding(u.sad, newRoleBinding)

	if err := u.sad.UpdateRoleBinding(oldBinding, newBinding); err != nil {
		u.log.Errorf("error during updating SAD rolebindings: %v", err)
	}
}

// TODO: if this is active then this needs to change the permissions on the
// destination subject
func (u *User) addFuncClusterRoleBinding(obj interface{}) {
	clusterRoleBinding, err := u.getClusterRoleBindingObject(obj)
	if err != nil {
		u.log.Errorf("failed to decode newly added cluster rolebinding: %v", err)
	}

	// if we arn't referenced or have seen this binding before, return
	if !u.clusterBindingContainsSubject(clusterRoleBinding) || u.seenUID(clusterRoleBinding.UID) {
		return
	}

	u.log.Infof("A new cluster rolebinding referencing '%s' has been added. Updating SAD", u.Name())

	u.addUID(clusterRoleBinding.UID)
	u.clusterBindings = append(u.clusterBindings, clusterRoleBinding)

	binding := role_binding.NewFromClusterRoleBinding(u.sad, clusterRoleBinding)

	if err := u.sad.AddRoleBinding(binding); err != nil {
		u.log.Errorf("Failed to add new cluster rolebinding: %v", err)
	}
}

// TODO: We need to tell the controller to update it's referenced rolebindings
func (u *User) delFuncClusterRoleBinding(obj interface{}) {
	clusterRoleBinding, err := u.getClusterRoleBindingObject(obj)
	if err != nil {
		u.log.Errorf("failed to decode newly deleted cluster rolebinding: %v", err)
		return
	}

	// if we arn't referenced or haven't seen this binding before, return
	if !u.clusterBindingContainsSubject(clusterRoleBinding) || !u.seenUID(clusterRoleBinding.UID) {
		return
	}

	u.log.Infof("A Cluster RoleBinding referencing '%s' has been deleted. Updating SAD", u.Name())

	if !u.deleteClusterRoleBinding(clusterRoleBinding.UID) {
		u.log.Errorf("Didn't find the deleted cluster rolbinding in SAD references. Something has gone very wrong.")
	}

	binding := role_binding.NewFromClusterRoleBinding(u.sad, clusterRoleBinding)

	if !u.sad.DeleteRoleBinding(binding) {
		u.log.Errorf("Failed to delete cluster rolebinding '%s'. It did not exist.", binding.Name)
	}
}

// TODO: we need to tell the controller to update it's replicated binding to
// newObj
func (u *User) updateClusterRoleBindingOject(oldObj, newObj interface{}) {
	oldClusterRoleBinding, err := u.getClusterRoleBindingObject(oldObj)
	if err != nil {
		u.log.Error(err)
		return
	}

	newClusterRoleBinding, err := u.getClusterRoleBindingObject(newObj)
	if err != nil {
		u.log.Error(err)
		return
	}

	// if we arn't referenced or haven't seen this binding before, return
	if !u.clusterBindingContainsSubject(oldClusterRoleBinding) || !u.seenUID(oldClusterRoleBinding.UID) {
		return
	}

	u.log.Infof("A Cluster RoleBinding referencing '%s' has been updated. Updating SAD", u.Name())

	u.addUID(newClusterRoleBinding.UID)
	u.clusterBindings = append(u.clusterBindings, newClusterRoleBinding)

	oldBinding := role_binding.NewFromClusterRoleBinding(u.sad, oldClusterRoleBinding)
	newBinding := role_binding.NewFromClusterRoleBinding(u.sad, newClusterRoleBinding)

	if err := u.sad.UpdateRoleBinding(oldBinding, newBinding); err != nil {
		u.log.Errorf("error during updating SAD cluster rolebindings: %v", err)
	}
}

func (u *User) getRoleBindingObject(obj interface{}) (*rbacv1.RoleBinding, error) {
	binding, err := utils.GetRoleBindingObject(obj)
	if err != nil {
		return nil, fmt.Errorf("failed to get rolebinding object: %v", err)
	}
	if binding == nil {
		return nil, errors.New("failed to get rolebinding, received nil object")
	}

	return binding, nil
}

func (u *User) getClusterRoleBindingObject(obj interface{}) (*rbacv1.ClusterRoleBinding, error) {
	binding, err := utils.GetClusterRoleBindingObject(obj)
	if err != nil {
		return nil, fmt.Errorf("failed to get cluster rolebinding object: %v", err)
	}
	if binding == nil {
		return nil, errors.New("failed to get cluster rolebinding, received nil object")
	}

	return binding, nil
}

func (u *User) deleteRoleBinding(uid types.UID) bool {
	for i, binding := range u.bindings {
		if binding.UID == uid {

			copy(u.bindings[i:], u.bindings[i+1:])
			u.bindings[len(u.bindings)-1] = nil
			u.bindings = u.bindings[:len(u.bindings)-1]

			return true
		}
	}

	return false
}

func (u *User) deleteClusterRoleBinding(uid types.UID) bool {
	for i, binding := range u.clusterBindings {
		if binding.UID == uid {

			copy(u.clusterBindings[i:], u.clusterBindings[i+1:])
			u.clusterBindings[len(u.clusterBindings)-1] = nil
			u.clusterBindings = u.clusterBindings[:len(u.clusterBindings)-1]

			return true
		}
	}

	return false
}

func (u *User) seenUID(uid types.UID) bool {
	b, ok := u.uids[uid]
	if !ok {
		return false
	}

	return b
}

func (u *User) addUID(uid types.UID) {
	u.uids[uid] = true
}
