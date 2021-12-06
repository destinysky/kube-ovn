package controller

import (
	"fmt"
	"reflect"

	kubeovnv1 "github.com/alauda/kube-ovn/pkg/apis/kubeovn/v1"
	"github.com/alauda/kube-ovn/pkg/ovs"
	"github.com/alauda/kube-ovn/pkg/util"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/client-go/tools/cache"
	"k8s.io/klog"
)

func (c *Controller) enqueueAddLpp(obj interface{}) {
	if !c.isLeader() {
		return
	}
	var key string
	var err error
	if key, err = cache.MetaNamespaceKeyFunc(obj); err != nil {
		utilruntime.HandleError(err)
		return
	}
	klog.V(3).Infof("enqueue add lpp %s", key)
	c.addLppQueue.Add(key)
}

func (c *Controller) enqueueDeleteLpp(obj interface{}) {
	if !c.isLeader() {
		return
	}
	var key string
	var err error
	if key, err = cache.MetaNamespaceKeyFunc(obj); err != nil {
		utilruntime.HandleError(err)
		return
	}
	klog.V(3).Infof("enqueue delete lpp %s", key)
	c.deleteLppQueue.Add(key)
}

func (c *Controller) enqueueUpdateLpp(old, new interface{}) {
	if !c.isLeader() {
		return
	}

	oldLpp := old.(*kubeovnv1.LogicalPortPair)
	newLpp := new.(*kubeovnv1.LogicalPortPair)

	if !reflect.DeepEqual(oldLpp.Spec, newLpp.Spec) {
		var key string
		var err error
		if key, err = cache.MetaNamespaceKeyFunc(new); err != nil {
			utilruntime.HandleError(err)
			return
		}
		klog.V(3).Infof("enqueue update lpp %s", key)
		c.updateLppQueue.Add(key)
	}
}

func (c *Controller) runAddLppWorker() {
	for c.processNextAddLppWorkItem() {
	}
}

func (c *Controller) runUpdateLppWorker() {
	for c.processNextUpdateLppWorkItem() {
	}
}

func (c *Controller) runDeleteLppWorker() {
	for c.processNextDeleteLppWorkItem() {
	}
}

func (c *Controller) processNextAddLppWorkItem() bool {
	obj, shutdown := c.addLppQueue.Get()
	if shutdown {
		return false
	}

	err := func(obj interface{}) error {
		defer c.addLppQueue.Done(obj)
		var key string
		var ok bool
		if key, ok = obj.(string); !ok {
			c.addLppQueue.Forget(obj)
			utilruntime.HandleError(fmt.Errorf("expected string in workqueue but got %#v", obj))
			return nil
		}
		if err := c.handleAddLpp(key); err != nil {
			c.addLppQueue.AddRateLimited(key)
			return fmt.Errorf("error syncing '%s': %s, requeuing", key, err.Error())
		}
		c.addLppQueue.Forget(obj)
		return nil
	}(obj)

	if err != nil {
		utilruntime.HandleError(err)
		return true
	}
	return true
}

func (c *Controller) processNextUpdateLppWorkItem() bool {
	obj, shutdown := c.updateLppQueue.Get()

	if shutdown {
		return false
	}

	err := func(obj interface{}) error {
		defer c.updateLppQueue.Done(obj)
		var key string
		var ok bool

		if key, ok = obj.(string); !ok {
			c.updateLppQueue.Forget(obj)
			utilruntime.HandleError(fmt.Errorf("expected string in workqueue but got %#v", obj))
			return nil
		}

		if err := c.handleUpdateLpp(key); err != nil {
			c.updateLppQueue.AddRateLimited(key)
			return fmt.Errorf("error syncing '%s': %s, requeuing", key, err.Error())
		}

		c.updateLppQueue.Forget(obj)
		return nil
	}(obj)

	if err != nil {
		utilruntime.HandleError(err)
		return true
	}

	return true
}

func (c *Controller) processNextDeleteLppWorkItem() bool {
	obj, shutdown := c.deleteLppQueue.Get()

	if shutdown {
		return false
	}

	err := func(obj interface{}) error {
		defer c.deleteLppQueue.Done(obj)
		var key string
		var ok bool

		if key, ok = obj.(string); !ok {
			c.deleteLppQueue.Forget(obj)
			utilruntime.HandleError(fmt.Errorf("expected string in workqueue but got %#v", obj))
			return nil
		}

		if err := c.handleDelLpp(key); err != nil {
			c.deleteLppQueue.AddRateLimited(key)
			return fmt.Errorf("error syncing '%s': %s, requeuing", key, err.Error())
		}

		c.deleteLppQueue.Forget(obj)
		return nil
	}(obj)

	if err != nil {
		utilruntime.HandleError(err)
		return true
	}

	return true
}

func (c *Controller) handleAddLpp(key string) error {
	namespace, name, err := cache.SplitMetaNamespaceKey(key)
	if err != nil {
		utilruntime.HandleError(fmt.Errorf("invalid resource key: %s", key))
		return nil
	}
	lpp, err := c.lppsLister.LogicalPortPairs(namespace).Get(name)
	if err != nil {
		if k8serrors.IsNotFound(err) {
			return nil
		}

		return err
	}

	pod, err := c.podsLister.Pods(namespace).Get(lpp.Spec.PodName)
	if err != nil {
		if k8serrors.IsNotFound(err) {
			return nil
		}
		return err
	}

	lsName, lsExist := pod.Annotations[util.LogicalSwitchAnnotation]
	if !lsExist {
		return nil
	}

	if err := c.ovnClient.CreateLogicalPortPair(lsName, ovs.PodNameToPortName(pod.Name, pod.Namespace), name, lpp.Spec.Weight); err != nil {
		return err
	}
	return nil
}

func (c *Controller) handleUpdateLpp(key string) error {
	namespace, name, err := cache.SplitMetaNamespaceKey(key)
	if err != nil {
		utilruntime.HandleError(fmt.Errorf("invalid resource key: %s", key))
		return nil
	}
	if err := c.ovnClient.DeleteLogicalPortPair(name); err != nil {
		return err
	}
	lpp, err := c.lppsLister.LogicalPortPairs(namespace).Get(name)
	if err != nil {
		if k8serrors.IsNotFound(err) {
			return nil
		}

		return err
	}

	pod, err := c.podsLister.Pods(namespace).Get(lpp.Spec.PodName)
	if err != nil {
		if k8serrors.IsNotFound(err) {
			return nil
		}
		return err
	}

	lsName, lsExist := pod.Annotations[util.LogicalSwitchAnnotation]
	if !lsExist {
		return nil
	}

	if err := c.ovnClient.CreateLogicalPortPair(lsName, ovs.PodNameToPortName(pod.Name, pod.Namespace), name, lpp.Spec.Weight); err != nil {
		return err
	}
	return nil
}

func (c *Controller) handleDelLpp(key string) error {
	_, name, err := cache.SplitMetaNamespaceKey(key)
	if err != nil {
		utilruntime.HandleError(fmt.Errorf("invalid resource key: %s", key))
		return nil
	}
	if err := c.ovnClient.DeleteLogicalPortPair(name); err != nil {
		return err
	}
	return nil
}
