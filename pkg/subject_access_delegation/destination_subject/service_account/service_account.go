package service_account

import (
	"errors"
	"fmt"
	"strings"

	"github.com/hashicorp/go-multierror"
	"github.com/sirupsen/logrus"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"

	"github.com/joshvanl/k8s-subject-access-delegation/pkg/interfaces"
	"github.com/joshvanl/k8s-subject-access-delegation/pkg/subject_access_delegation/utils"
)

type ServiceAccount struct {
	log    *logrus.Entry
	client kubernetes.Interface
	sad    interfaces.SubjectAccessDelegation

	namespace      string
	name           string
	serviceAccount *corev1.ServiceAccount
}

var _ interfaces.DestinationSubject = &ServiceAccount{}

func New(sad interfaces.SubjectAccessDelegation, name string) ([]interfaces.DestinationSubject, error) {
	if strings.Contains(name, "*") {
		if !utils.ValidName(name) {
			return nil, fmt.Errorf("Not a valid regular expression for Service Account '%s'", name)
		}

		options := metav1.ListOptions{}
		sas, err := sad.Client().Core().ServiceAccounts(sad.Namespace()).List(options)
		if err != nil {
			return nil, fmt.Errorf("failed to list service accounts: %v", err)
		}

		var result *multierror.Error
		var serviceAccounts []interfaces.DestinationSubject

		for _, sa := range sas.Items {
			match, err := utils.MatchName(sa.Name, name)

			if err != nil {
				result = multierror.Append(result, fmt.Errorf("error matching service account regex: %v", err))
			} else if match {

				serviceAccounts = append(serviceAccounts, &ServiceAccount{
					log:       sad.Log(),
					client:    sad.Client(),
					sad:       sad,
					namespace: sad.Namespace(),
					name:      sa.Name,
				})
			}
		}

		return serviceAccounts, result.ErrorOrNil()
	}

	return []interfaces.DestinationSubject{
		&ServiceAccount{
			log:       sad.Log(),
			client:    sad.Client(),
			sad:       sad,
			namespace: sad.Namespace(),
			name:      name,
		},
	}, nil
}

func (d *ServiceAccount) getServiceAccount() error {
	options := metav1.GetOptions{}

	sa, err := d.client.Core().ServiceAccounts(d.Namespace()).Get(d.Name(), options)
	if err != nil {
		return fmt.Errorf("failed to get service account '%s': %v", d.Name(), err)
	}
	d.serviceAccount = sa

	return nil
}

func (d *ServiceAccount) ResolveDestination() error {
	if err := d.getServiceAccount(); err != nil {
		return err
	}

	if d.serviceAccount == nil {
		return errors.New("service account is nil")
	}

	return nil
}

func (d *ServiceAccount) Namespace() string {
	return d.namespace
}

func (d *ServiceAccount) Name() string {
	return d.name
}

func (d *ServiceAccount) Kind() string {
	return "ServiceAccount"
}
