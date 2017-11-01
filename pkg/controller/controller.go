package controller

import (
	"fmt"
	"reflect"
	"time"

	//"github.com/hashicorp/go-multierror"
	"github.com/sirupsen/logrus"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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
	"github.com/joshvanl/k8s-subject-access-delegation/pkg/subject_access_delegation"
)

// TODO: Add a name to a subject access delegatio
// When a delete happens of a SAD remove the permissions of that delegation
// Add a start and stop time/duration of the permissions
// Don't stop the delegation repeating if it wasn't ablt to apply, just wait again?? (Not sure, repeat on failure)
// Support more resources and origin subejects, not just role bindings. e.g. get the role bindings of a user and use that to apply permissions

//TODO How to map the sads with the triggers? Maybe string map or something, Doesn't work!

const controllerAgentName = "SAD-controller"

const (
	SuccessSynced         = "Synced"
	ErrResourceExists     = "ErrResourceExists"
	MessageResourceExists = "Resource %q already exists and is not managed by Subject Access Delegation"
	MessageResourceSynced = "Subject Access Delegation synced successfully"
)

type Controller struct {
	kubeclientset kubernetes.Interface
	sadclientset  clientset.Interface

	sadsLister listers.SubjectAccessDelegationLister
	sadsSynced cache.InformerSynced

	workqueue workqueue.RateLimitingInterface

	apiserverURL string
	log          *logrus.Entry

	delegations map[string]*subject_access_delegation.SubjectAccessDelegation
}

var stopCh = make(chan struct{})

func NewController(
	kubeclientset kubernetes.Interface,
	sadclientset clientset.Interface,
	kubeInformerFactory kubeinformers.SharedInformerFactory,
	sadInformerFactory informers.SharedInformerFactory,
	log *logrus.Entry) *Controller {

	sadInformer := sadInformerFactory.Authz().V1alpha1().SubjectAccessDelegations()

	sadscheme.AddToScheme(scheme.Scheme)

	controller := &Controller{
		kubeclientset: kubeclientset,
		sadclientset:  sadclientset,
		sadsLister:    sadInformer.Lister(),
		sadsSynced:    sadInformer.Informer().HasSynced,
		workqueue:     workqueue.NewNamedRateLimitingQueue(workqueue.DefaultControllerRateLimiter(), "SubjectAccessDelegation"),
		log:           log,
	}

	log.Info("Setting up event handlers")
	sadInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: controller.enqueueSad,
		UpdateFunc: func(old, new interface{}) {
			if !reflect.DeepEqual(old, new) {
				controller.enqueueSad(new)
			}
		},
		DeleteFunc: controller.deleteSad,
	})

	controller.delegations = make(map[string]*subject_access_delegation.SubjectAccessDelegation)

	return controller
}

func (c *Controller) Run(threadiness int, stopCh <-chan struct{}) error {
	defer runtime.HandleCrash()
	defer c.workqueue.ShutDown()

	c.log.Info("Starting SAD controller")

	c.log.Info("Waiting for informer caches to sync")
	if ok := cache.WaitForCacheSync(stopCh, c.sadsSynced); !ok {
		return fmt.Errorf("failed to wait for caches to sync")
	}

	c.log.Info("Starting workers")
	for i := 0; i < threadiness; i++ {
		go wait.Until(c.runWorker, time.Second, stopCh)
	}

	c.log.Info("Started workers")
	<-stopCh
	c.log.Info("Shutting down workers")

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
		c.log.Infof("Successfully synced '%s'", key)
		return nil
	}(obj)

	if err != nil {
		runtime.HandleError(err)
		return true
	}

	return true
}

func (c *Controller) syncHandler(key string) error {
	namespace, name, err := cache.SplitMetaNamespaceKey(key)
	if err != nil {
		runtime.HandleError(fmt.Errorf("invalid resource key: %s", key))
		return nil
	}

	sad, err := c.sadsLister.SubjectAccessDelegations(namespace).Get(name)
	if err != nil {
		if errors.IsNotFound(err) {
			runtime.HandleError(fmt.Errorf("sad '%s' in work queue no longer exists", key))
			return nil
		}

		return err
	}

	if !sad.Status.Processed {
		if err := c.ProcessDelegation(sad); err != nil {
			c.log.Errorf("failed to process Subject Access Delegation: %v", err)
			return err
		}
	}

	if err := c.updateSadStatus(sad); err != nil {
		return err
	}

	//c.recorder.Event(foo, corev1.EventTypeNormal, SuccessSynced, MessageResourceSynced)
	return nil
}

func (c *Controller) ProcessDelegation(sad *authzv1alpha1.SubjectAccessDelegation) error {
	c.log.Infof("New Subject Access Delegation '%s'", sad.Name)

	delegation := subject_access_delegation.New(sad, c.kubeclientset, c.log)
	if err := c.appendDelegation(delegation, sad); err != nil {
		return err
	}

	go func() {
		err := delegation.Delegate()
		if err != nil {
			// ----> If it fails here, delete from queue etc. <----
			c.log.Errorf("error during Subject Access Delegation '%s': %v", delegation.Name(), err)
		}
	}()

	return nil
}

func (c *Controller) updateSadStatus(sad *authzv1alpha1.SubjectAccessDelegation) error {
	sadCopy := sad.DeepCopy()
	sadCopy.Status.Processed = true
	_, err := c.sadclientset.AuthzV1alpha1().SubjectAccessDelegations(sad.Namespace).Update(sadCopy)
	return err
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
		c.log.Error(err)
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
		c.log.Errorf("unable to delete delegation '%s': no longer exists in controller")
		return
	}

	if err := delegation.Delete(); err != nil {
		c.log.Errorf("error deleting Subject Access Delegation: %v", err)
		return
	}

	c.delegations[name] = nil

	c.log.Infof("Subject Access Delegation '%s' has been deleted", name)
}

func (c *Controller) appendDelegation(delegation *subject_access_delegation.SubjectAccessDelegation, sad *authzv1alpha1.SubjectAccessDelegation) error {
	if _, ok := c.delegations[sad.Name]; ok {
		return fmt.Errorf("Subject Access Delegation '%s' already exists.", sad.Name)
	}

	c.delegations[sad.Name] = delegation

	return nil
}
