package utils

import (
	"fmt"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/tools/cache"
)

func GetPodObject(obj interface{}) (pod *corev1.Pod, err error) {
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
	}

	if pod, ok = object.(*corev1.Pod); !ok {
		return nil, fmt.Errorf("failed to covert object to Pod type")
	}

	return pod, nil
}

func GetNodeObject(obj interface{}) (node *corev1.Node, err error) {
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
	}

	if node, ok = object.(*corev1.Node); !ok {
		return nil, fmt.Errorf("failed to covert object to Pod type")
	}

	return node, nil
}

func GetServiceObject(obj interface{}) (service *corev1.Service, err error) {
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
	}

	if service, ok = object.(*corev1.Service); !ok {
		return nil, fmt.Errorf("failed to covert object to Pod type")
	}

	return service, nil
}
