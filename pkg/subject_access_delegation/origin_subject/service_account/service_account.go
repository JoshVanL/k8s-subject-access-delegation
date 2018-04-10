package service_account

import (
	"errors"
	"fmt"

	"github.com/sirupsen/logrus"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	informer "k8s.io/client-go/informers/rbac/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"

	"github.com/joshvanl/k8s-subject-access-delegation/pkg/interfaces"
)

type ServiceAccount struct {
	log    *logrus.Entry
	client kubernetes.Interface
	sad    interfaces.SubjectAccessDelegation

	namespace      string
	name           string
	serviceAccount *corev1.ServiceAccount

	bindings        []*rbacv1.RoleBinding
	clusterBindings []*rbacv1.ClusterRoleBinding
	uids            map[types.UID]bool
	stopCh          chan struct{}

	bindingInformer        informer.RoleBindingInformer
	clusterBindingInformer informer.ClusterRoleBindingInformer
}

var _ interfaces.OriginSubject = &ServiceAccount{}

func New(sad interfaces.SubjectAccessDelegation, name string) *ServiceAccount {
	return &ServiceAccount{
		log:                    sad.Log(),
		client:                 sad.Client(),
		sad:                    sad,
		namespace:              sad.Namespace(),
		name:                   name,
		bindingInformer:        sad.KubeInformerFactory().Rbac().V1().RoleBindings(),
		clusterBindingInformer: sad.KubeInformerFactory().Rbac().V1().ClusterRoleBindings(),
		stopCh:                 make(chan struct{}),
	}
}

func (s *ServiceAccount) ResolveOrigin() error {
	if err := s.serviceAccountObject(); err != nil {
		return err
	}

	if err := s.roleBindings(); err != nil {
		return err
	}

	s.ListenRolebindings()

	return nil
}

func (s *ServiceAccount) RoleRefs() (roleRefs []*rbacv1.RoleRef, clusterRoleRefs []*rbacv1.RoleRef) {
	for _, binding := range s.bindings {
		roleRefs = append(roleRefs, &binding.DeepCopy().RoleRef)
	}

	for _, binding := range s.clusterBindings {
		clusterRoleRefs = append(clusterRoleRefs, &binding.DeepCopy().RoleRef)
	}

	return roleRefs, clusterRoleRefs
}

func (s *ServiceAccount) roleBindings() error {
	// make this more efficient
	options := metav1.ListOptions{}

	bindingsList, err := s.client.Rbac().RoleBindings(s.Namespace()).List(options)
	if err != nil {
		return fmt.Errorf("failed to retrieve Rolebindings of Service Account '%s': %v", s.Name(), err)
	}

	clusterBindingsList, err := s.client.Rbac().ClusterRoleBindings().List(options)
	if err != nil {
		return fmt.Errorf("failed to retrieve Cluster Rolebindings of Service Account '%s': %v", s.Name(), err)
	}

	s.bindings = make([]*rbacv1.RoleBinding, 0)
	s.clusterBindings = make([]*rbacv1.ClusterRoleBinding, 0)
	s.uids = make(map[types.UID]bool)

	for _, binding := range bindingsList.Items {
		if s.bindingContainsSubject(binding.DeepCopy()) {
			s.bindings = append(s.bindings, binding.DeepCopy())
			s.uids[binding.UID] = true
		}
	}

	for _, binding := range clusterBindingsList.Items {
		if s.clusterBindingContainsSubject(binding.DeepCopy()) {
			s.clusterBindings = append(s.clusterBindings, binding.DeepCopy())
			s.uids[binding.UID] = true
		}
	}

	return nil
}

func (s *ServiceAccount) serviceAccountObject() error {
	options := metav1.GetOptions{}

	serviceAccount, err := s.client.Core().ServiceAccounts(s.Namespace()).Get(s.Name(), options)
	if err != nil {
		return fmt.Errorf("failed to get Service Account '%s': %v", s.Name(), err)
	}

	if serviceAccount == nil {
		return errors.New("service account is nil")
	}

	s.serviceAccount = serviceAccount

	return nil
}

func (s *ServiceAccount) ListenRolebindings() {
	s.bindingInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc:    s.addFuncRoleBinding,
		UpdateFunc: s.updateRoleBinding,
		DeleteFunc: s.delFuncRoleBinding,
	})

	s.clusterBindingInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc:    s.addFuncClusterRoleBinding,
		UpdateFunc: s.updateClusterRoleBinding,
		DeleteFunc: s.delFuncClusterRoleBinding,
	})

	go s.clusterBindingInformer.Informer().Run(s.stopCh)
}

func (s *ServiceAccount) bindingContainsSubject(binding *rbacv1.RoleBinding) bool {
	for _, subject := range binding.Subjects {
		if subject.Kind == rbacv1.ServiceAccountKind && subject.Name == s.Name() {
			return true
		}
	}

	return false
}

func (s *ServiceAccount) clusterBindingContainsSubject(binding *rbacv1.ClusterRoleBinding) bool {
	for _, subject := range binding.Subjects {
		if subject.Kind == rbacv1.ServiceAccountKind && subject.Name == s.Name() {
			return true
		}
	}

	return false
}

func (s *ServiceAccount) Namespace() string {
	return s.namespace
}

func (s *ServiceAccount) Name() string {
	return s.name
}

func (s *ServiceAccount) Kind() string {
	return rbacv1.ServiceAccountKind
}

func (s *ServiceAccount) Delete() {
	close(s.stopCh)
}
