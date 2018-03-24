package group

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

type Group struct {
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

var _ interfaces.OriginSubject = &Group{}

func New(sad interfaces.SubjectAccessDelegation, name string) *Group {
	return &Group{
		log:                    sad.Log(),
		client:                 sad.Client(),
		sad:                    sad,
		namespace:              sad.Namespace(),
		name:                   name,
		bindingInformer:        sad.KubeInformerFactory().Rbac().V1().RoleBindings(),
		clusterBindingInformer: sad.KubeInformerFactory().Rbac().V1().ClusterRoleBindings(),
	}
}

func (g *Group) ResolveOrigin() error {
	if err := g.roleBindings(); err != nil {
		return err
	}

	g.ListenRolebindings()

	return nil
}

func (g *Group) RoleRefs() (roleRefs []*rbacv1.RoleRef, clusterRoleRefs []*rbacv1.RoleRef) {
	for _, binding := range g.bindings {
		roleRefs = append(roleRefs, &binding.DeepCopy().RoleRef)
	}

	for _, binding := range g.clusterBindings {
		clusterRoleRefs = append(clusterRoleRefs, &binding.DeepCopy().RoleRef)
	}

	return roleRefs, clusterRoleRefs
}

func (g *Group) ListenRolebindings() {
	g.bindingInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc:    g.addFuncRoleBinding,
		UpdateFunc: g.updateRoleBindingOject,
		DeleteFunc: g.delFuncRoleBinding,
	})

	g.clusterBindingInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc:    g.addFuncClusterRoleBinding,
		UpdateFunc: g.updateClusterRoleBindingOject,
		DeleteFunc: g.delFuncClusterRoleBinding,
	})

	g.clusterBindingInformer.Informer().Run(make(chan struct{}))
}

func (g *Group) roleBindings() error {
	// make this more efficient
	options := metav1.ListOptions{}

	bindingsList, err := g.client.Rbac().RoleBindings(g.Namespace()).List(options)
	if err != nil {
		return fmt.Errorf("failed to retrieve Rolebindings of Group Account '%s': %v", g.Name(), err)
	}

	clusterBindingsList, err := g.client.Rbac().ClusterRoleBindings().List(options)
	if err != nil {
		return fmt.Errorf("failed to retrieve Cluster Rolebindings of Group Account '%s': %v", g.Name(), err)
	}

	g.bindings = make([]*rbacv1.RoleBinding, 0)
	g.clusterBindings = make([]*rbacv1.ClusterRoleBinding, 0)
	g.uids = make(map[types.UID]bool)

	for _, binding := range bindingsList.Items {
		if g.bindingContainsSubject(binding.DeepCopy()) {
			g.bindings = append(g.bindings, binding.DeepCopy())
			g.uids[binding.UID] = true
		}
	}

	for _, binding := range clusterBindingsList.Items {
		if g.clusterBindingContainsSubject(binding.DeepCopy()) {
			g.clusterBindings = append(g.clusterBindings, binding.DeepCopy())
			g.uids[binding.UID] = true
		}
	}

	return nil
}

func (g *Group) bindingContainsSubject(binding *rbacv1.RoleBinding) bool {
	for _, subject := range binding.Subjects {
		if subject.Kind == rbacv1.GroupKind && subject.Name == g.Name() {
			return true
		}
	}

	return false
}

func (g *Group) clusterBindingContainsSubject(binding *rbacv1.ClusterRoleBinding) bool {
	for _, subject := range binding.Subjects {
		if subject.Kind == rbacv1.GroupKind && subject.Name == g.Name() {
			return true
		}
	}

	return false
}

func (g *Group) Namespace() string {
	return g.namespace
}

func (g *Group) Name() string {
	return g.name
}

func (g *Group) Kind() string {
	return rbacv1.GroupKind
}
