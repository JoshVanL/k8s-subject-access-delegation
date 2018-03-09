package utils

import (
	"errors"

	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/tools/cache"
)

func GetPodObject(obj interface{}) (*corev1.Pod, error) {
	var object metav1.Object
	var ok bool
	if object, ok = obj.(metav1.Object); !ok {
		tombstone, ok := obj.(cache.DeletedFinalStateUnknown)
		if !ok {
			return nil, errors.New("error decoding object, invalid type")
		}
		object, ok = tombstone.Obj.(metav1.Object)
		if !ok {
			return nil, errors.New("error decoding object tombstone, invalid type")
		}
	}

	pod, ok := object.(*corev1.Pod)
	if !ok {
		return nil, errors.New("failed to covert object to Pod type")
	}

	return pod, nil
}

func GetNodeObject(obj interface{}) (*corev1.Node, error) {
	var object metav1.Object
	var ok bool
	if object, ok = obj.(metav1.Object); !ok {
		tombstone, ok := obj.(cache.DeletedFinalStateUnknown)
		if !ok {
			return nil, errors.New("error decoding object, invalid type")
		}
		object, ok = tombstone.Obj.(metav1.Object)
		if !ok {
			return nil, errors.New("error decoding object tombstone, invalid type")
		}
	}

	node, ok := object.(*corev1.Node)
	if !ok {
		return nil, errors.New("failed to covert object to Pod type")
	}

	return node, nil
}

func GetServiceObject(obj interface{}) (*corev1.Service, error) {
	var object metav1.Object
	var ok bool
	if object, ok = obj.(metav1.Object); !ok {
		tombstone, ok := obj.(cache.DeletedFinalStateUnknown)
		if !ok {
			return nil, errors.New("error decoding object, invalid type")
		}
		object, ok = tombstone.Obj.(metav1.Object)
		if !ok {
			return nil, errors.New("error decoding object tombstone, invalid type")
		}
	}

	service, ok := object.(*corev1.Service)
	if !ok {
		return nil, errors.New("failed to covert object to ServiceAccount type")
	}

	return service, nil
}

func GetRoleBindingObject(obj interface{}) (*rbacv1.RoleBinding, error) {
	var object metav1.Object
	var ok bool
	if object, ok = obj.(metav1.Object); !ok {
		tombstone, ok := obj.(cache.DeletedFinalStateUnknown)
		if !ok {
			return nil, errors.New("error decoding object, invalid type")
		}
		object, ok = tombstone.Obj.(metav1.Object)
		if !ok {
			return nil, errors.New("error decoding object tombstone, invalid type")
		}
	}

	binding, ok := object.(*rbacv1.RoleBinding)
	if !ok {
		return nil, errors.New("failed to covert object to RoleBinding type")
	}

	return binding, nil
}

func GetClusterRoleBindingObject(obj interface{}) (*rbacv1.ClusterRoleBinding, error) {
	var object metav1.Object
	var ok bool
	if object, ok = obj.(metav1.Object); !ok {
		tombstone, ok := obj.(cache.DeletedFinalStateUnknown)
		if !ok {
			return nil, errors.New("error decoding object, invalid type")
		}
		object, ok = tombstone.Obj.(metav1.Object)
		if !ok {
			return nil, errors.New("error decoding object tombstone, invalid type")
		}
	}

	binding, ok := object.(*rbacv1.ClusterRoleBinding)
	if !ok {
		return nil, errors.New("failed to covert object to Cluster RoleBinding type")
	}

	return binding, nil
}
