package controller

import (
	"fmt"
	"reflect"
	"time"

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
	"github.com/joshvanl/k8s-subject-access-delegation/pkg/trigger"
)

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
		c.log.Infof("Here is some stuff:\n%s\n%s\n%s\n%s", sad.Spec.OriginSubject, sad.Spec.Duration, sad.Spec.Repeat, sad.Spec.DestinationSubject)

		timeTrigger := trigger.New(c.log, sad, c.kubeclientset)
		if err := timeTrigger.Validate(); err != nil {
			c.log.Infof("THIS IS AN ERROR: %v", err)
		}
	}

	//// Get the deployment with the name specified in Foo.spec
	//sadc.sadsLister.SubjectAccessDelegations(
	//deployment, err := c.deploymentsLister.Deployments(sad.Namespace).Get(deploymentName)
	//// If the resource doesn't exist, we'll create it
	//if errors.IsNotFound(err) {
	//	deployment, err = c.kubeclientset.AppsV1beta2().Deployments(sa.Namespace).Create(newDeployment(sad))
	//}

	// If an error occurs during Get/Create, we'll requeue the item so we can
	// attempt processing again later. This could have been caused by a
	// temporary network failure, or any other transient reason.
	//if err != nil {
	//	return err
	//}

	// If the Deployment is not controlled by this Foo resource, we should log
	// a warning to the event recorder and ret
	//if !metav1.IsControlledBy(deployment, sad) {
	//	msg := fmt.Sprintf(MessageResourceExists, deployment.Name)
	//	c.recorder.Event(sad, corev1.EventTypeWarning, ErrResourceExists, msg)
	//	return fmt.Errorf(msg)
	//}

	//// If this number of the replicas on the Foo resource is specified, and the
	//// number does not equal the current desired replicas on the Deployment, we
	//// should update the Deployment resource.
	//if sad.Spec.Replicas != nil && *foo.Spec.Replicas != *deployment.Spec.Replicas {
	//	glog.V(4).Infof("Foor: %d, deplR: %d", *foo.Spec.Replicas, *deployment.Spec.Replicas)
	//	deployment, err = c.kubeclientset.AppsV1beta2().Deployments(foo.Namespace).Update(newDeployment(foo))
	//}

	//// If an error occurs during Update, we'll requeue the item so we can
	//// attempt processing again later. THis could have been caused by a
	//// temporary network failure, or any other transient reason.
	//if err != nil {
	//	return err
	//}

	// Finally, we update the status block of the Sad resource to reflect the
	// current state of the world
	err = c.updateSadStatus(sad)
	if err != nil {
		return err
	}

	//c.recorder.Event(foo, corev1.EventTypeNormal, SuccessSynced, MessageResourceSynced)
	return nil
}

func (c *Controller) updateSadStatus(sad *authzv1alpha1.SubjectAccessDelegation) error {
	sadCopy := sad.DeepCopy()
	sadCopy.Status.Processed = true
	_, err := c.sadclientset.AuthzV1alpha1().SubjectAccessDelegations(sad.Namespace).Update(sadCopy)
	return err
}

// enqueueSad takes a Sad resource and converts it into a namespace/name
// string which is then put onto the work queue. This method should *not* be
// passed resources of any type other than Foo.
func (c *Controller) enqueueSad(obj interface{}) {
	var key string
	var err error
	if key, err = cache.MetaNamespaceKeyFunc(obj); err != nil {
		runtime.HandleError(err)
		return
	}
	c.workqueue.AddRateLimited(key)
}

// handleObject will take any resource implementing metav1.Object and attempt
// to find the Sad resource that 'owns' it. It does this by looking at the
// objects metadata.ownerReferences field for an appropriate OwnerReference.
// It then enqueues that Foo resource to be processed. If the object does not
// have an appropriate OwnerReference, it will simply be skipped.
func (c *Controller) handleObject(obj interface{}) {
	var object metav1.Object
	var ok bool
	if object, ok = obj.(metav1.Object); !ok {
		tombstone, ok := obj.(cache.DeletedFinalStateUnknown)
		if !ok {
			runtime.HandleError(fmt.Errorf("error decoding object, invalid type"))
			return
		}
		object, ok = tombstone.Obj.(metav1.Object)
		if !ok {
			runtime.HandleError(fmt.Errorf("error decoding object tombstone, invalid type"))
			return
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
			return
		}

		c.enqueueSad(sad)
		return
	}
}

func (c *Controller) deleteSad(obj interface{}) {
	c.log.Info("You have deleted a sad")
}
