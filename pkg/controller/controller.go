package controller

import (
	"fmt"
	"time"

	"github.com/hashicorp/go-multierror"
	"github.com/sirupsen/logrus"
	apiextv1beta1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
	apiextcs "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
	kubeinformers "k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/util/workqueue"

	authzv1alpha1 "github.com/joshvanl/k8s-subject-access-delegation/pkg/apis/authz/v1alpha1"
	clientset "github.com/joshvanl/k8s-subject-access-delegation/pkg/client/clientset/versioned"
	sadscheme "github.com/joshvanl/k8s-subject-access-delegation/pkg/client/clientset/versioned/scheme"
	informers "github.com/joshvanl/k8s-subject-access-delegation/pkg/client/informers/externalversions"
	listers "github.com/joshvanl/k8s-subject-access-delegation/pkg/client/listers/authz/v1alpha1"
	"github.com/joshvanl/k8s-subject-access-delegation/pkg/interfaces"
	"github.com/joshvanl/k8s-subject-access-delegation/pkg/ntp_client"
	"github.com/joshvanl/k8s-subject-access-delegation/pkg/subject_access_delegation"
	"github.com/joshvanl/k8s-subject-access-delegation/pkg/subject_access_delegation/utils"
)

const controllerAgentName = "SAD-controller"

type Controller struct {
	kubeclientset kubernetes.Interface
	sadclientset  clientset.Interface

	sadsLister          listers.SubjectAccessDelegationLister
	sadsSynced          cache.InformerSynced
	kubeInformerFactory kubeinformers.SharedInformerFactory

	workqueue workqueue.RateLimitingInterface

	apiserverURL string
	log          *logrus.Entry
	ntpClient    *ntp_client.NTPClient
	clockOffset  time.Duration

	delegations map[string]interfaces.SubjectAccessDelegation

	seenUids    map[types.UID]bool
	deletedUids map[types.UID]bool
}

var _ interfaces.Controller = &Controller{}

func NewController(
	kubeclientset kubernetes.Interface,
	sadclientset clientset.Interface,
	kubeInformerFactory kubeinformers.SharedInformerFactory,
	sadInformerFactory informers.SharedInformerFactory,
	hosts []string,
	log *logrus.Entry) *Controller {

	log.Infof("Initialising Subject Access Delegation Controller...")

	sadInformer := sadInformerFactory.Authz().V1alpha1().SubjectAccessDelegations()

	sadscheme.AddToScheme(scheme.Scheme)

	controller := &Controller{
		kubeclientset:       kubeclientset,
		kubeInformerFactory: kubeInformerFactory,
		sadclientset:        sadclientset,
		sadsLister:          sadInformer.Lister(),
		sadsSynced:          sadInformer.Informer().HasSynced,
		workqueue:           workqueue.NewNamedRateLimitingQueue(workqueue.DefaultControllerRateLimiter(), "SubjectAccessDelegation"),
		log:                 log,
		ntpClient:           ntp_client.NewNTPClient(hosts),
		delegations:         make(map[string]interfaces.SubjectAccessDelegation),
		seenUids:            make(map[types.UID]bool),
		deletedUids:         make(map[types.UID]bool),
	}

	sadInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: controller.enqueueSad,
		UpdateFunc: func(old, new interface{}) {
			controller.updateSadObject(old, new)
		},
		DeleteFunc: controller.deleteSad,
	})

	return controller
}

func (c *Controller) Run(threadiness int, stopCh <-chan struct{}) error {
	defer runtime.HandleCrash()
	defer c.workqueue.ShutDown()

	if c.ntpClient != nil {
		if err := c.tryGetNtpTime(); err != nil {
			c.log.Errorf("Failed to set accurate time for controller: %v", err)
			c.log.Warn("Continuing without optimum clock accuracy.")
		}

	} else {
		c.log.Infof("No NTP server URLs specified.")
	}

	c.log.Infof("Using current time: %s", time.Now().Add(c.clockOffset).String())

	c.log.Info("Waiting for informer caches to sync...")
	if ok := cache.WaitForCacheSync(stopCh, c.sadsSynced); !ok {
		return fmt.Errorf("failed to wait for caches to sync")
	}

	c.log.Info("Starting Workers...")
	for i := 0; i < threadiness; i++ {
		go wait.Until(c.runWorker, time.Second, stopCh)
	}

	c.log.Infof("Syncing any Subject Access Delegations already created...")
	if err := c.updateStateFromAPI(); err != nil {
		c.log.Errorf("Error syncing with API server: %v", err)
	}

	c.log.Info("Controller Ready.")

	<-stopCh

	return nil
}

func (c *Controller) updateSadObject(old, new interface{}) {
	oldSad, err := utils.GetSubjectAccessDelegationObject(old)
	if err != nil {
		c.log.Errorf("unable to decode old sad from update: %v", err)
		return
	}
	newSad, err := utils.GetSubjectAccessDelegationObject(new)
	if err != nil {
		c.log.Errorf("unable to decode new sad from update: %v", err)
		return
	}

	delegation, ok := c.delegations[oldSad.Name]
	if !ok {
		c.log.Errorf("unable to find old delegation from old in update '%s'", oldSad.Name)
		return
	}

	notify, err := delegation.UpdateSadObject(newSad)
	if err != nil {
		c.log.Errorf("error during update of Subject Access Delegation object: %v", err)
		return
	}
	if notify {
		c.log.Infof("Subject Access Delegation '%s' Successfully Updated!", newSad.Name)
	}
}

func (c *Controller) updateStateFromAPI() error {
	options := metav1.ListOptions{}
	namespaces, err := c.kubeclientset.Core().Namespaces().List(options)
	if err != nil {
		return fmt.Errorf("failed to list namespaces in current cluster: %v", err)
	}

	var result *multierror.Error
	var sads []*authzv1alpha1.SubjectAccessDelegation

	for _, namespace := range namespaces.Items {
		delegations, err := c.sadsLister.SubjectAccessDelegations(namespace.Name).List(labels.Everything())
		if err != nil {
			result = multierror.Append(result, err)

		} else {

			for _, sad := range delegations {
				sads = append(sads, sad.DeepCopy())
			}
		}
	}

	for _, sad := range sads {
		if _, ok := c.delegations[sad.Name]; !ok {
			if err := c.ProcessDelegation(sad); err != nil {
				result = multierror.Append(result, fmt.Errorf("failed to process present delegation '%s': %v", sad.Name, err))
			}
		}
	}

	return result.ErrorOrNil()
}

func (c *Controller) tryGetNtpTime() error {
	c.log.Info("Getting current time fromm NTP server(s)...")
	if err := c.getOffSet(); err != nil {
		return err
	}
	c.log.Infof("Controller/system clock offset: %s", c.clockOffset.String())

	return nil
}

func (c *Controller) runWorker() {
	for c.processNextWorkItem() {
	}
}

func (c *Controller) processNextWorkItem() bool {
	obj, shutdown := c.workqueue.Get()

	if shutdown {
		return false
	}

	err := func(obj interface{}) error {
		defer c.workqueue.Done(obj)
		var key string
		var ok bool
		if key, ok = obj.(string); !ok {
			c.workqueue.Forget(obj)
			runtime.HandleError(fmt.Errorf("expected string in workqueue but got %#v", obj))
			return nil
		}
		if err := c.syncHandler(key); err != nil {
			return fmt.Errorf("error syncing '%s': %s", key, err.Error())
		}
		c.workqueue.Forget(obj)
		return nil
	}(obj)

	if err != nil {
		runtime.HandleError(err)
		return true
	}

	return true
}

func (c *Controller) getOffSet() (err error) {
	if c.clockOffset, err = c.ntpClient.GetOffset(); err != nil {
		return err
	}

	return nil
}

func (c *Controller) syncHandler(key string) error {
	namespace, name, err := cache.SplitMetaNamespaceKey(key)
	if err != nil {
		runtime.HandleError(fmt.Errorf("invalid resource key: %s", key))
		return nil
	}

	sad, err := c.sadsLister.SubjectAccessDelegations(namespace).Get(name)
	if err != nil {
		if apierrors.IsNotFound(err) {
			runtime.HandleError(fmt.Errorf("sad '%s' in work queue no longer exists", key))
			return nil
		}

		return err
	}

	if _, ok := c.delegations[sad.Name]; !ok && !sad.Status.Processed {
		if err := c.ProcessDelegation(sad); err != nil {
			c.log.Errorf("Failed to process Subject Access Delegation: %v", err)
			return err
		}
	}

	return nil
}

func (c *Controller) ProcessDelegation(sad *authzv1alpha1.SubjectAccessDelegation) error {
	c.log.Infof("New Subject Access Delegation '%s'", sad.Name)

	delegation := subject_access_delegation.New(c, sad, c.log, c.kubeInformerFactory, c.kubeclientset, c.sadclientset, c.clockOffset)
	if err := c.appendDelegation(delegation, sad); err != nil {
		return err
	}

	go func() {
		closed, err := delegation.Delegate()
		if err != nil {
			c.log.Errorf("Error processing Subject Access Delegation '%s': %v", delegation.Name(), err)
		}

		delegation.DeleteRoleBindings()

		if !closed {
			c.manuallyDeleteSad(sad)
		}

		return
	}()

	return nil
}

func (c *Controller) manuallyDeleteSad(sad *authzv1alpha1.SubjectAccessDelegation) {
	options := &metav1.DeleteOptions{}
	err := c.sadclientset.Authz().SubjectAccessDelegations(sad.Namespace).Delete(sad.Name, options)
	if err != nil {
		c.log.Errorf("Failed to delete Subject Access Delegation '%s' after completion: %v", sad.Name, err)
		return
	}
}

func (c *Controller) enqueueSad(obj interface{}) {
	var key string
	var err error
	if key, err = cache.MetaNamespaceKeyFunc(obj); err != nil {
		runtime.HandleError(err)
		return
	}
	c.workqueue.AddRateLimited(key)

}

func (c *Controller) handleObject(obj interface{}) {
	sad, err := c.getSADObject(obj)
	if err != nil {
		c.log.Errorf("Error handling object: %v", err)
	}

	c.enqueueSad(sad)
}

func (c *Controller) getSADObject(obj interface{}) (sad *authzv1alpha1.SubjectAccessDelegation, err error) {

	var object metav1.Object
	var ok bool
	if object, ok = obj.(metav1.Object); !ok {
		tombstone, ok := obj.(cache.DeletedFinalStateUnknown)
		if !ok {
			return nil, fmt.Errorf("error decoding object, invalid type")
		}
		object, ok = tombstone.Obj.(metav1.Object)
		if !ok {
			return nil, fmt.Errorf("error decoding object tombstone, invalid type")
		}
		c.log.Infof("Recovered deleted object '%s' from tombstone", object.GetName())
	}
	c.log.Infof("Processing object: %s", object.GetName())
	if ownerRef := metav1.GetControllerOf(object); ownerRef != nil {
		// If this object is not owned by a Sad, we should not do anything more
		// with it.
		if ownerRef.Kind != "SubjectAccessDelegation" {
			return
		}

		sad, err := c.sadsLister.SubjectAccessDelegations(object.GetNamespace()).Get(ownerRef.Name)
		if err != nil {
			c.log.Infof("ignoring orphaned object '%s' of foo '%s'", object.GetSelfLink(), ownerRef.Name)
			return nil, nil
		}

		return sad, nil
	}

	return nil, nil
}

func (c *Controller) deleteSad(obj interface{}) {
	object, ok := obj.(metav1.Object)
	if !ok {
		c.log.Warn("unable to retrieve object for deletion")
	}

	name := object.GetName()

	delegation, ok := c.delegations[name]
	if !ok {
		c.log.Errorf("unable to delete delegation '%s': no longer exists in controller", name)
		return
	}

	delete(c.delegations, name)

	if err := delegation.Delete(); err != nil {
		c.log.Errorf("error deleting Subject Access Delegation: %v", err)
		return
	}

	c.log.Infof("Subject Access Delegation '%s' has been deleted", name)
}

func (c *Controller) appendDelegation(delegation interfaces.SubjectAccessDelegation, sad *authzv1alpha1.SubjectAccessDelegation) error {
	if _, ok := c.delegations[sad.Name]; ok {
		return fmt.Errorf("Subject Access Delegation '%s' already exists.", sad.Name)
	}

	c.delegations[sad.Name] = delegation

	return nil
}

func (c *Controller) EnsureCRD(clientset apiextcs.Interface) error {
	c.log.Info("Ensuring Subject Access Delegation Custom Resource Definition...")

	crd := &apiextv1beta1.CustomResourceDefinition{
		ObjectMeta: metav1.ObjectMeta{
			Name: "subjectaccessdelegations.authz.sad",
		},
		Spec: apiextv1beta1.CustomResourceDefinitionSpec{
			Group:   "authz.sad",
			Version: "v1alpha1",
			Names: apiextv1beta1.CustomResourceDefinitionNames{
				Plural:     "subjectaccessdelegations",
				Singular:   "subjectaccessdelegation",
				Kind:       "SubjectAccessDelegation",
				ShortNames: []string{"sad"},
			},
			Scope: "Namespaced",
		},
	}

	crd.APIVersion = "apiextensions.k8s.io/v1beta1"
	crd.Kind = "CustomResourceDefinition"

	_, err := clientset.ApiextensionsV1beta1().CustomResourceDefinitions().Create(crd)
	if err != nil {
		if apierrors.IsAlreadyExists(err) {
			c.log.Info("Custom Resource Definition Already Exists.")
			return nil

		} else {
			return err
		}
	}

	// Ensure that the custom resource definition has been created before continuing
	for trys := 0; trys < 3; trys++ {

		crd, err := clientset.ApiextensionsV1beta1().CustomResourceDefinitions().Get("subjectaccessdelegations.authz.sad", metav1.GetOptions{})
		if err != nil {
			if apierrors.IsNotFound(err) {
				c.log.Infof("Custom resource not yet found, retrying (%d/3)..", trys+1)
			} else {
				c.log.Warnf("Error ensuring crd was created: %v", err)
			}

			continue
		}

		if crd != nil {
			c.log.Infof("Custom Resource Definition Successfully Created.")
			return nil
		}

		time.Sleep(time.Second)
	}

	return fmt.Errorf("unable to find SAD custom resource definition from Kubetnetes")
}

func (c *Controller) AddUid(uid types.UID) {
	c.seenUids[uid] = true
}

func (c *Controller) DeleteUid(uid types.UID) {
	c.deletedUids[uid] = true
}

func (c *Controller) SeenUid(uid types.UID) bool {
	return c.seenUids[uid]
}
func (c *Controller) DeletedUid(uid types.UID) bool {
	return c.deletedUids[uid]
}
