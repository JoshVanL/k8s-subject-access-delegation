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

const serviceAccountKind = "ServiceAccount"

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

	bindingInformer        informer.RoleBindingInformer
	clusterbindingInformer informer.ClusterRoleBindingInformer
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
		clusterbindingInformer: sad.KubeInformerFactory().Rbac().V1().ClusterRoleBindings(),
	}
}

// TODO: this just needs to return the role refs of the rolebindings
func (s *ServiceAccount) RoleRefs() (roleRefs []*rbacv1.RoleRef, clusterRoleRefs []*rbacv1.RoleRef, err error) {
	for _, binding := range s.bindings {
		roleRef := binding.RoleRef
		roleRefs = append(roleRefs, &roleRef)
	}

	for _, binding := range s.clusterBindings {
		roleRef := binding.RoleRef
		clusterRoleRefs = append(clusterRoleRefs, &roleRef)
	}

	return roleRefs, clusterRoleRefs, nil
}

func (s *ServiceAccount) getRoleBindings() error {
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
		if s.bindingContainsSubject(&binding) {
			s.bindings = append(s.bindings, &binding)
			s.uids[binding.UID] = true
			break
		}
	}

	for _, binding := range clusterBindingsList.Items {
		if s.clusterBindingContainsSubject(&binding) {
			s.clusterBindings = append(s.clusterBindings, &binding)
			s.uids[binding.UID] = true
		}
	}

	return nil
}

func (s *ServiceAccount) getServiceAccount() error {
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

func (s *ServiceAccount) ResolveOrigin() error {
	if err := s.getServiceAccount(); err != nil {
		return err
	}

	if err := s.getRoleBindings(); err != nil {
		return err
	}

	s.ListenRolebindings()

	return nil
}

func (s *ServiceAccount) ListenRolebindings() {
	s.bindingInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc:    s.addFuncRoleBinding,
		UpdateFunc: s.updateRoleBindingOject,
		DeleteFunc: s.delFuncRoleBinding,
	})

	s.clusterbindingInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc:    s.addFuncClusterRoleBinding,
		UpdateFunc: s.updateClusterRoleBindingOject,
		DeleteFunc: s.delFuncClusterRoleBinding,
	})
}

func (s *ServiceAccount) bindingContainsSubject(binding *rbacv1.RoleBinding) bool {
	for _, subject := range binding.Subjects {
		if subject.Kind == serviceAccountKind && subject.Name == s.Name() {
			return true
		}
	}

	return false
}

func (s *ServiceAccount) clusterBindingContainsSubject(binding *rbacv1.ClusterRoleBinding) bool {
	for _, subject := range binding.Subjects {
		if subject.Kind == serviceAccountKind && subject.Name == s.Name() {
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
	return serviceAccountKind
}
