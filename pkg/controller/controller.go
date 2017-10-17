package controller

import (
	"errors"
	"fmt"
	"log"
	"reflect"
	"time"

	"github.com/sirupsen/logrus"
	"k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/util/workqueue"

	"github.com/joshvanl/k8s-subject-access-delegation/pkg/apis/authz/v1alpha1"
	"github.com/joshvanl/k8s-subject-access-delegation/pkg/client"
	factory "github.com/joshvanl/k8s-subject-access-delegation/pkg/informers/externalversions"
)

type Controller struct {
	apiserverURL  string
	queue         workqueue.RateLimitingInterface
	client        *client.Clientset
	sharedFactory factory.SharedInformerFactory
	log           *logrus.Entry
	informer      cache.SharedIndexInformer
}

var stopCh = make(chan struct{})

func New(apiserverURL string, log *logrus.Entry) (c *Controller, err error) {
	if log == nil {
		return nil, errors.New("parameter logrus Entry is nil")
	}

	c = &Controller{
		log: log,
	}

	c.client, err = client.NewForConfig(&rest.Config{
		Host: apiserverURL,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create api client: %v", err)
	}

	c.sharedFactory = factory.NewSharedInformerFactory(c.client, time.Second*30)

	c.informer = c.sharedFactory.Authz().V1alpha1().SubjectAccessDelegations().Informer()
	c.informer.AddEventHandler(
		cache.ResourceEventHandlerFuncs{
			AddFunc: c.enqueue,
			UpdateFunc: func(old, cur interface{}) {
				if !reflect.DeepEqual(old, cur) {
					c.enqueue(cur)
				}
			},
			DeleteFunc: foo,
		},
	)

	c.sharedFactory.Start(stopCh)

	if !cache.WaitForCacheSync(stopCh, c.informer.HasSynced) {
		return nil, errors.New("error waiting for informer cache to sync")
	}

	c.apiserverURL = apiserverURL
	c.queue = workqueue.NewRateLimitingQueue(workqueue.NewItemExponentialFailureRateLimiter(time.Second*5, time.Minute))

	return c, nil
}

func (c *Controller) enqueue(obj interface{}) {
	key, err := cache.DeletionHandlingMetaNamespaceKeyFunc(obj)
	if err != nil {
		runtime.HandleError(fmt.Errorf("error obtaining key for object being enqueue: %s", err.Error()))
		return
	}

	c.queue.Add(key)
}

func foo(obj interface{}) {
	fmt.Println("This has been deleted: %v", obj)
}

func (c *Controller) Work() error {
	for {
		key, shutdown := c.queue.Get()

		if shutdown {
			stopCh <- struct{}{}
			return nil
		}

		var strKey string
		var ok bool
		if strKey, ok = key.(string); !ok {
			rerr := fmt.Errorf("key in queue should be of type string but got %T. discarding", key)
			runtime.HandleError(rerr)
			return rerr
		}

		err := func(key string) error {
			defer c.queue.Done(key)

			namespace, name, err := cache.SplitMetaNamespaceKey(strKey)

			if err != nil {
				rerr := fmt.Errorf("error splitting meta namespace key into parts: %v", err)
				runtime.HandleError(rerr)
				return rerr
			}

			c.log.Infof("Read item '%s/%s' off workqueue. Processing...", namespace, name)

			// retrieve the latest version in the cache of this message
			obj, err := c.sharedFactory.Authz().V1alpha1().SubjectAccessDelegations().Lister().SubjectAccessDelegations(namespace).Get(name)
			if err != nil {
				rerr := fmt.Errorf("error getting object '%s/%s' from api: %v", namespace, name, err)
				runtime.HandleError(rerr)
				return rerr
			}

			c.log.Infof("Got most up to date version of '%s/%s'. Syncing...", namespace, name)

			if err := c.sync(obj); err != nil {
				rerr := fmt.Errorf("error processing item '%s/%s': %v", namespace, name, err)
				runtime.HandleError(rerr)
				return rerr
			}

			c.log.Infof("Finished processing '%s/%s' successfully! Removing from queue.", namespace, name)

			c.queue.Forget(key)
			return nil
		}(strKey)

		if err != nil {
			return err
		}
	}

	return nil
}

func (c *Controller) sync(al *v1alpha1.SubjectAccessDelegation) error {
	if al.Status.Processed {
		log.Printf("Skipping already Processed Subject Access '%s/%s'", al.Namespace, al.Name)
		return nil
	}

	note := fmt.Sprintf("%s\n%s\n%s", al.Spec.User, al.Spec.Duration, al.Spec.SomeStuff)
	c.log.Infof("Processed Subject Access!\n> %s\n", note)

	al.Status.Processed = true
	if _, err := c.client.AuthzV1alpha1().SubjectAccessDelegations(al.Namespace).Update(al); err != nil {
		return fmt.Errorf("error saving update to authz Subject Access resource: %v", err)
	}
	c.log.Infof("Finished saving update to authz Subject Access resource '%s/%s'", al.Namespace, al.Name)

	return nil
}
