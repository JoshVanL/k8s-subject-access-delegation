package user

import (
	"fmt"

	"github.com/sirupsen/logrus"
	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	informer "k8s.io/client-go/informers/rbac/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"

	"github.com/joshvanl/k8s-subject-access-delegation/pkg/interfaces"
)

type User struct {
	log    *logrus.Entry
	client kubernetes.Interface
	sad    interfaces.SubjectAccessDelegation

	namespace string
	name      string
	uids      map[types.UID]bool

	bindings        []*rbacv1.RoleBinding
	clusterBindings []*rbacv1.ClusterRoleBinding

	bindingInformer        informer.RoleBindingInformer
	clusterBindingInformer informer.ClusterRoleBindingInformer
}

var _ interfaces.OriginSubject = &User{}

func New(sad interfaces.SubjectAccessDelegation, name string) *User {
	return &User{
		log:                    sad.Log(),
		client:                 sad.Client(),
		sad:                    sad,
		namespace:              sad.Namespace(),
		name:                   name,
		bindingInformer:        sad.KubeInformerFactory().Rbac().V1().RoleBindings(),
		clusterBindingInformer: sad.KubeInformerFactory().Rbac().V1().ClusterRoleBindings(),
	}
}

func (u *User) ResolveOrigin() error {
	if err := u.roleBindings(); err != nil {
		return err
	}

	u.ListenRolebindings()

	return nil
}

func (u *User) RoleRefs() (roleRefs []*rbacv1.RoleRef, clusterRoleRefs []*rbacv1.RoleRef) {
	for _, binding := range u.bindings {
		roleRefs = append(roleRefs, &binding.DeepCopy().RoleRef)
	}

	for _, binding := range u.clusterBindings {
		clusterRoleRefs = append(clusterRoleRefs, &binding.DeepCopy().RoleRef)
	}

	return roleRefs, clusterRoleRefs
}

func (u *User) ListenRolebindings() {
	u.bindingInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc:    u.addFuncRoleBinding,
		UpdateFunc: u.updateRoleBinding,
		DeleteFunc: u.delFuncRoleBinding,
	})

	u.clusterBindingInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc:    u.addFuncClusterRoleBinding,
		UpdateFunc: u.updateClusterRoleBinding,
		DeleteFunc: u.delFuncClusterRoleBinding,
	})

	go u.clusterBindingInformer.Informer().Run(make(chan struct{}))
}

func (u *User) roleBindings() error {
	// make this more efficient
	options := metav1.ListOptions{}

	bindingsList, err := u.client.Rbac().RoleBindings(u.Namespace()).List(options)
	if err != nil {
		return fmt.Errorf("failed to retrieve Rolebindings of User Account '%s': %v", u.Name(), err)
	}

	clusterBindingsList, err := u.client.Rbac().ClusterRoleBindings().List(options)
	if err != nil {
		return fmt.Errorf("failed to retrieve Cluster Rolebindings of User Account '%s': %v", u.Name(), err)
	}

	u.bindings = make([]*rbacv1.RoleBinding, 0)
	u.clusterBindings = make([]*rbacv1.ClusterRoleBinding, 0)
	u.uids = make(map[types.UID]bool)

	for _, binding := range bindingsList.Items {
		if u.bindingContainsSubject(binding.DeepCopy()) {
			u.bindings = append(u.bindings, binding.DeepCopy())
			u.uids[binding.UID] = true
		}
	}

	for _, binding := range clusterBindingsList.Items {
		if u.clusterBindingContainsSubject(binding.DeepCopy()) {
			u.clusterBindings = append(u.clusterBindings, binding.DeepCopy())
			u.uids[binding.UID] = true
		}
	}

	return nil
}

func (u *User) bindingContainsSubject(binding *rbacv1.RoleBinding) bool {
	for _, subject := range binding.Subjects {
		if subject.Kind == rbacv1.UserKind && subject.Name == u.Name() {
			return true
		}
	}

	return false
}

func (u *User) clusterBindingContainsSubject(binding *rbacv1.ClusterRoleBinding) bool {
	for _, subject := range binding.Subjects {
		if subject.Kind == rbacv1.UserKind && subject.Name == u.Name() {
			return true
		}
	}

	return false
}

func (u *User) Namespace() string {
	return u.namespace
}

func (u *User) Name() string {
	return u.name
}

func (u *User) Kind() string {
	return rbacv1.UserKind
}
