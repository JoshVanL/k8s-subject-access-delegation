package group

import (
	"errors"
	"fmt"

	rbacv1 "k8s.io/api/rbac/v1"
	"k8s.io/apimachinery/pkg/types"

	"github.com/joshvanl/k8s-subject-access-delegation/pkg/subject_access_delegation/role_binding"
	"github.com/joshvanl/k8s-subject-access-delegation/pkg/subject_access_delegation/utils"
)

func (g *Group) addFuncRoleBinding(obj interface{}) {
	roleBinding, err := g.getRoleBindingObject(obj)
	if err != nil {
		g.log.Errorf("failed to decode newly added rolebinding: %v", err)
		return
	}

	// if we arn't referenced or we've seen this binding before, return
	if !g.bindingContainsSubject(roleBinding) || g.seenUID(roleBinding.UID) {
		return
	}

	g.log.Debugf("A new rolebinding referencing '%s' has been added. Updating SAD", g.Name())

	g.addUID(roleBinding.UID)
	g.bindings = append(g.bindings, roleBinding)

	binding := role_binding.NewFromRoleBinding(g.sad, roleBinding)

	if err := g.sad.AddRoleBinding(binding); err != nil {
		g.log.Errorf("Failed to add new rolebinding: %v", err)
	}
}

func (g *Group) delFuncRoleBinding(obj interface{}) {
	roleBinding, err := g.getRoleBindingObject(obj)
	if err != nil {
		g.log.Errorf("failed to decode newly deleted rolebinding: %v", err)
		return
	}

	// if we arn't referenced or haven't seen this binding before, return
	if !g.bindingContainsSubject(roleBinding) || !g.seenUID(roleBinding.UID) {
		return
	}

	g.log.Debugf("A RoleBinding referencing '%s' has been deleted. Updating SAD", g.Name())

	if !g.deleteRoleBinding(roleBinding.UID) {
		g.log.Errorf("Didn't find the deleted rolbinding in SAD references. Something has gone very wrong.")
	}

	binding := role_binding.NewFromRoleBinding(g.sad, roleBinding)

	if err := g.sad.DeleteRoleBinding(binding); err != nil {
		g.log.Errorf("Failed to delete rolebinding '%s': %v", binding.Name(), err)
	}
}

func (g *Group) updateRoleBindingOject(oldObj, newObj interface{}) {
	oldRoleBinding, err := g.getRoleBindingObject(oldObj)
	if err != nil {
		g.log.Error(err)
		return
	}

	newRoleBinding, err := g.getRoleBindingObject(newObj)
	if err != nil {
		g.log.Error(err)
		return
	}

	// if we arn't referenced or haven't seen this binding before, return
	if !g.bindingContainsSubject(oldRoleBinding) || !g.seenUID(oldRoleBinding.UID) || !g.changedBinding(oldRoleBinding, newRoleBinding) {
		return
	}

	g.log.Debugf("A RoleBinding referencing '%s' has been updated. Updating SAD", g.Name())

	if !g.deleteRoleBinding(oldRoleBinding.UID) {
		g.log.Errorf("Didn't find the deleted rolbinding in SAD references. Something has gone very wrong.")
	}

	g.addUID(newRoleBinding.UID)
	g.bindings = append(g.bindings, newRoleBinding)

	oldBinding := role_binding.NewFromRoleBinding(g.sad, oldRoleBinding)
	newBinding := role_binding.NewFromRoleBinding(g.sad, newRoleBinding)

	if err := g.sad.UpdateRoleBinding(oldBinding, newBinding); err != nil {
		g.log.Errorf("error during updating SAD rolebindings: %v", err)
	}
}

func (g *Group) addFuncClusterRoleBinding(obj interface{}) {
	clusterRoleBinding, err := g.getClusterRoleBindingObject(obj)
	if err != nil {
		g.log.Errorf("failed to decode newly added cluster rolebinding: %v", err)
	}

	// if we arn't referenced or have seen this binding before, return
	if !g.clusterBindingContainsSubject(clusterRoleBinding) || g.seenUID(clusterRoleBinding.UID) {
		return
	}

	g.log.Debugf("A new cluster rolebinding referencing '%s' has been added. Updating SAD", g.Name())

	g.addUID(clusterRoleBinding.UID)
	g.clusterBindings = append(g.clusterBindings, clusterRoleBinding)

	binding := role_binding.NewFromClusterRoleBinding(g.sad, clusterRoleBinding)

	if err := g.sad.AddRoleBinding(binding); err != nil {
		g.log.Errorf("Failed to add new cluster rolebinding: %v", err)
	}
}

func (g *Group) delFuncClusterRoleBinding(obj interface{}) {
	clusterRoleBinding, err := g.getClusterRoleBindingObject(obj)
	if err != nil {
		g.log.Errorf("failed to decode newly deleted cluster rolebinding: %v", err)
		return
	}

	// if we arn't referenced or haven't seen this binding before, return
	if !g.clusterBindingContainsSubject(clusterRoleBinding) || !g.seenUID(clusterRoleBinding.UID) {
		return
	}

	g.log.Debugf("A Cluster RoleBinding referencing '%s' has been deleted. Updating SAD", g.Name())

	if !g.deleteClusterRoleBinding(clusterRoleBinding.UID) {
		g.log.Errorf("Didn't find the deleted cluster rolbinding in SAD references. Something has gone very wrong.")
	}

	binding := role_binding.NewFromClusterRoleBinding(g.sad, clusterRoleBinding)

	if err := g.sad.DeleteRoleBinding(binding); err != nil {
		g.log.Errorf("Failed to delete cluster rolebinding '%s': %v", binding.Name(), err)
	}
}

func (g *Group) updateClusterRoleBindingOject(oldObj, newObj interface{}) {
	oldClusterRoleBinding, err := g.getClusterRoleBindingObject(oldObj)
	if err != nil {
		g.log.Error(err)
		return
	}

	newClusterRoleBinding, err := g.getClusterRoleBindingObject(newObj)
	if err != nil {
		g.log.Error(err)
		return
	}

	// if we arn't referenced or haven't seen this binding before, return
	if !g.clusterBindingContainsSubject(oldClusterRoleBinding) || !g.seenUID(oldClusterRoleBinding.UID) || !g.changedClusterBinding(oldClusterRoleBinding, newClusterRoleBinding) {
		return
	}

	g.log.Debugf("A Cluster RoleBinding referencing '%s' has been updated. Updating SAD", g.Name())

	g.addUID(newClusterRoleBinding.UID)
	g.clusterBindings = append(g.clusterBindings, newClusterRoleBinding)

	oldBinding := role_binding.NewFromClusterRoleBinding(g.sad, oldClusterRoleBinding)
	newBinding := role_binding.NewFromClusterRoleBinding(g.sad, newClusterRoleBinding)

	if err := g.sad.UpdateRoleBinding(oldBinding, newBinding); err != nil {
		g.log.Errorf("error during updating SAD cluster rolebindings: %v", err)
	}
}

func (g *Group) getRoleBindingObject(obj interface{}) (*rbacv1.RoleBinding, error) {
	binding, err := utils.GetRoleBindingObject(obj)
	if err != nil {
		return nil, fmt.Errorf("failed to get rolebinding object: %v", err)
	}
	if binding == nil {
		return nil, errors.New("failed to get rolebinding, received nil object")
	}

	return binding, nil
}

func (g *Group) getClusterRoleBindingObject(obj interface{}) (*rbacv1.ClusterRoleBinding, error) {
	binding, err := utils.GetClusterRoleBindingObject(obj)
	if err != nil {
		return nil, fmt.Errorf("failed to get cluster rolebinding object: %v", err)
	}
	if binding == nil {
		return nil, errors.New("failed to get cluster rolebinding, received nil object")
	}

	return binding, nil
}

func (g *Group) deleteRoleBinding(uid types.UID) bool {
	for i, binding := range g.bindings {
		if binding.UID == uid {

			copy(g.bindings[i:], g.bindings[i+1:])
			g.bindings[len(g.bindings)-1] = nil
			g.bindings = g.bindings[:len(g.bindings)-1]

			return true
		}
	}

	return false
}

func (g *Group) deleteClusterRoleBinding(uid types.UID) bool {
	for i, binding := range g.clusterBindings {
		if binding.UID == uid {

			copy(g.clusterBindings[i:], g.clusterBindings[i+1:])
			g.clusterBindings[len(g.clusterBindings)-1] = nil
			g.clusterBindings = g.clusterBindings[:len(g.clusterBindings)-1]

			return true
		}
	}

	return false
}

func (g *Group) seenUID(uid types.UID) bool {
	b, ok := g.uids[uid]
	if !ok {
		return false
	}

	return b
}

func (g *Group) addUID(uid types.UID) {
	g.uids[uid] = true
}

func (g *Group) changedBinding(old, new *rbacv1.RoleBinding) bool {
	if old.Name != new.Name || old.Namespace != new.Namespace {
		return false
	}

	if old.RoleRef.Name != new.RoleRef.Name || old.RoleRef.Kind != new.RoleRef.Kind {
		return false
	}

	var changed bool
	for _, oldS := range old.Subjects {
		changed = true
		for _, newS := range new.Subjects {
			if oldS.Name == newS.Name && oldS.Kind == newS.Kind {
				changed = false
				break
			}
		}

		if changed {
			break
		}
	}

	return changed
}

func (g *Group) changedClusterBinding(old, new *rbacv1.ClusterRoleBinding) bool {
	if old.Name != new.Name {
		return false
	}

	if old.RoleRef.Name != new.RoleRef.Name || old.RoleRef.Kind != new.RoleRef.Kind {
		return false
	}

	var changed bool
	for _, oldS := range old.Subjects {
		changed = true
		for _, newS := range new.Subjects {
			if oldS.Name == newS.Name && oldS.Kind == newS.Kind {
				changed = false
				break
			}
		}

		if changed {
			break
		}
	}

	return changed
}
