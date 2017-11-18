package utils

import (
	"fmt"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	lister "k8s.io/client-go/listers/core/v1"
	"k8s.io/client-go/tools/cache"
)

func GetPodObject(lister lister.PodLister, obj interface{}) (pod *corev1.Pod, err error) {
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
		//c.log.Infof("Recovered deleted object '%s' from tombstone", object.GetName())
	}
	//c.log.Infof("Processing object: %s", object.GetName())
	if ownerRef := metav1.GetControllerOf(object); ownerRef != nil {
		if ownerRef.Kind != "Pod" {
			return
		}

		pod, err := lister.Pods(object.GetNamespace()).Get(ownerRef.Name)
		if err != nil {
			//c.log.Infof("ignoring orphaned object '%s' of foo '%s'", object.GetSelfLink(), ownerRef.Name)
			return nil, nil
		}

		return pod, nil
	}

	return nil, nil
}
