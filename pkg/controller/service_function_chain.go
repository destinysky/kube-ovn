package controller

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"

	kubeovnv1 "github.com/alauda/kube-ovn/pkg/apis/kubeovn/v1"
	"github.com/alauda/kube-ovn/pkg/ovs"
	"github.com/alauda/kube-ovn/pkg/util"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/client-go/tools/cache"
	"k8s.io/klog"
)

func (c *Controller) enqueueAddSfc(obj interface{}) {
	if !c.isLeader() {
		return
	}
	var key string
	var err error
	if key, err = cache.MetaNamespaceKeyFunc(obj); err != nil {
		utilruntime.HandleError(err)
		return
	}
	klog.V(3).Infof("enqueue add sfc %s", key)
	c.addSfcQueue.Add(key)
}

type DeleteQueueItem struct {
	Status        *kubeovnv1.ServiceFunctionChainStatus
	NamespaceName string
}

func (c *Controller) enqueueDeleteSfc(obj interface{}) {
	if !c.isLeader() {
		return
	}
	delSfc := obj.(*kubeovnv1.ServiceFunctionChain)
	var key string
	var err error
	if key, err = cache.MetaNamespaceKeyFunc(obj); err != nil {
		utilruntime.HandleError(err)
		return
	}
	klog.V(3).Infof("enqueue delete sfc %s", key)
	c.deleteSfcQueue.Add(DeleteQueueItem{
		NamespaceName: key,
		Status:        delSfc.Status.DeepCopy(),
	})
}

type UpdateQueueItem struct {
	OldSpec       *kubeovnv1.ServiceFunctionChainSpec
	NewSpec       *kubeovnv1.ServiceFunctionChainSpec
	NamespaceName string
}

func (c *Controller) enqueueUpdateSfc(old, new interface{}) {
	if !c.isLeader() {
		return
	}

	oldSfc := old.(*kubeovnv1.ServiceFunctionChain)
	newSfc := new.(*kubeovnv1.ServiceFunctionChain)
	if !reflect.DeepEqual(oldSfc.Spec, newSfc.Spec) {
		var key string
		var err error
		if key, err = cache.MetaNamespaceKeyFunc(new); err != nil {
			utilruntime.HandleError(err)
			return
		}
		klog.V(3).Infof("enqueue update sfc %s", key)
		c.updateSfcQueue.Add(UpdateQueueItem{
			OldSpec:       oldSfc.Spec.DeepCopy(),
			NewSpec:       newSfc.Spec.DeepCopy(),
			NamespaceName: key,
		})
	}
}

func (c *Controller) runAddSfcWorker() {
	for c.processNextAddSfcWorkItem() {
	}
}

func (c *Controller) runUpdateSfcWorker() {
	for c.processNextUpdateSfcWorkItem() {
	}
}

func (c *Controller) runDeleteSfcWorker() {
	for c.processNextDeleteSfcWorkItem() {
	}
}

func (c *Controller) processNextAddSfcWorkItem() bool {
	obj, shutdown := c.addSfcQueue.Get()
	if shutdown {
		return false
	}

	err := func(obj interface{}) error {
		defer c.addSfcQueue.Done(obj)
		var key string
		var ok bool
		if key, ok = obj.(string); !ok {
			c.addSfcQueue.Forget(obj)
			utilruntime.HandleError(fmt.Errorf("expected string in workqueue but got %#v", obj))
			return nil
		}
		if err := c.handleAddSfc(key); err != nil {
			c.addSfcQueue.AddRateLimited(key)
			return fmt.Errorf("error syncing '%s': %s, requeuing", key, err.Error())
		}
		c.addSfcQueue.Forget(obj)
		return nil
	}(obj)

	if err != nil {
		utilruntime.HandleError(err)
		return true
	}
	return true
}

func (c *Controller) processNextUpdateSfcWorkItem() bool {
	obj, shutdown := c.updateSfcQueue.Get()

	if shutdown {
		return false
	}

	err := func(obj interface{}) error {
		defer c.updateSfcQueue.Done(obj)
		var key UpdateQueueItem
		var ok bool

		if key, ok = obj.(UpdateQueueItem); !ok {
			c.updateSfcQueue.Forget(obj)
			utilruntime.HandleError(fmt.Errorf("expected struct UpdateQueueItem in workqueue but got %#v", obj))
			return nil
		}

		if err := c.handleUpdateSfc(key); err != nil {
			c.updateSfcQueue.AddRateLimited(key)
			return fmt.Errorf("error syncing '%s': %s, requeuing", key.NamespaceName, err.Error())
		}

		c.updateSfcQueue.Forget(obj)
		return nil
	}(obj)

	if err != nil {
		utilruntime.HandleError(err)
		return true
	}

	return true
}

func (c *Controller) processNextDeleteSfcWorkItem() bool {
	obj, shutdown := c.deleteSfcQueue.Get()

	if shutdown {
		return false
	}

	err := func(obj interface{}) error {
		defer c.deleteSfcQueue.Done(obj)
		var key DeleteQueueItem
		var ok bool
		if key, ok = obj.(DeleteQueueItem); !ok {
			c.deleteSfcQueue.Forget(obj)
			utilruntime.HandleError(fmt.Errorf("expected string in workqueue but got %#v", obj))
			return nil
		}
		if err := c.handleDelSfc(key); err != nil {
			c.deleteSfcQueue.AddRateLimited(key)
			return fmt.Errorf("error syncing '%s': %s, requeuing", key.NamespaceName, err.Error())
		}

		c.deleteSfcQueue.Forget(obj)
		return nil
	}(obj)

	if err != nil {
		utilruntime.HandleError(err)
		return true
	}

	return true
}

func (c *Controller) reverseChanges(del_classifier, del_chain string, del_lpp, del_lppg []string) error {

	if del_classifier != "" {
		if err := c.ovnClient.DeleteChainClassifier(del_classifier); err != nil {
			return err
		}
	}

	if del_chain != "" {
		if err := c.ovnClient.DeleteLogicalPortChain(del_chain); err != nil {
			return err
		}
	}

	if len(del_lppg) != 0 {
		for _, delPPG := range del_lppg {
			if err := c.ovnClient.DeleteLogicalPortPairGroup(delPPG); err != nil {
				return err
			}
		}
	}

	if len(del_lpp) != 0 {
		for _, delPP := range del_lpp {
			if err := c.ovnClient.DeleteLogicalPortPair(delPP); err != nil {
				return err
			}
		}
	}

	return nil

}

func (c *Controller) handleAddSfc(key string) error {
	namespace, name, err := cache.SplitMetaNamespaceKey(key)
	if err != nil {
		utilruntime.HandleError(fmt.Errorf("invalid resource key: %s", key))
		return nil
	}

	sfc, err := c.sfcsLister.ServiceFunctionChains(namespace).Get(name)
	if err != nil {
		if k8serrors.IsNotFound(err) {
			return nil
		}

		return err
	}

	if len(sfc.Spec.Chain) < 1 {
		return fmt.Errorf("The length of chain %s is zero.", key)
	}

	if len(sfc.Spec.Chain[0].PortPairs) < 1 {
		return fmt.Errorf("The number of portPairs in the first port group of chain %s is zero.", key)
	}

	pod, err := c.podsLister.Pods(namespace).Get(sfc.Spec.Chain[0].PortPairs[0].PodName)

	if err != nil {
		if k8serrors.IsNotFound(err) {
			return fmt.Errorf("Do not found Pod with name %s in chain %s", sfc.Spec.Chain[0].PortPairs[0].PodName, key)
		}
		return err
	}

	lsName_main, lsExist := pod.Annotations[util.LogicalSwitchAnnotation]
	if !lsExist {
		return fmt.Errorf("Pod %s in chain %s does not have switch name in annotation.", sfc.Spec.Chain[0].PortPairs[0].PodName, key)
	}

	if err := c.ovnClient.CreateLogicalPortChain(lsName_main, namespace+"_"+name+"_chain"); err != nil {
		return err
	}
	created_lpc := namespace + "_" + name + "_chain"

	var created_lpp []string
	var created_lppg []string

	for i, lppg := range sfc.Spec.Chain {
		if len(lppg.PortPairs) < 1 {
			if err2 := c.reverseChanges("", created_lpc, created_lpp, created_lppg); err2 != nil {
				return fmt.Errorf("The number of portPairs in the No.%d port group of chain %s is zero. Error in rollback configuration: %v.", i, key, err2)
			}
			return fmt.Errorf("The number of portPairs in the No.%d port group of chain %s is zero.", i, key)
		}

		if err := c.ovnClient.AddToLogicalPortChain(created_lpc, namespace+"_"+name+"_lppg_"+strconv.Itoa(i)); err != nil {
			if err2 := c.reverseChanges("", created_lpc, created_lpp, created_lppg); err2 != nil {
				return fmt.Errorf("%v Error in rollback configuration: %v.", err, err2)
			}
			return err
		}
		created_lppg = append(created_lppg, namespace+"_"+name+"_lppg_"+strconv.Itoa(i))

		for j, lpp := range lppg.PortPairs {
			pod, err := c.podsLister.Pods(namespace).Get(lpp.PodName)

			if err != nil {
				if k8serrors.IsNotFound(err) {
					if err2 := c.reverseChanges("", created_lpc, created_lpp, created_lppg); err2 != nil {
						return fmt.Errorf("Do not found Pod with name %s in chain %s. Error in rollback configuration: %v.", lpp.PodName, key, err2)
					}
					return fmt.Errorf("Do not found Pod with name %s in chain %s.", lpp.PodName, key)
				}
				if err2 := c.reverseChanges("", created_lpc, created_lpp, created_lppg); err2 != nil {
					return fmt.Errorf("%v Error in rollback configuration: %v.", err, err2)
				}
				return err
			}

			lsName, lsExist := pod.Annotations[util.LogicalSwitchAnnotation]
			if !lsExist {
				if err2 := c.reverseChanges("", created_lpc, created_lpp, created_lppg); err2 != nil {
					return fmt.Errorf("Pod %s in chain %s does not have switch name in annotation. Error in rollback configuration: %v.", lpp.PodName, key, err2)
				}
				return fmt.Errorf("Pod %s in chain %s does not have switch name in annotation.", lpp.PodName, key)
			}
			if lsName_main == "" {
				lsName_main = lsName
			} else if lsName_main != lsName {
				if err2 := c.reverseChanges("", created_lpc, created_lpp, created_lppg); err2 != nil {
					return fmt.Errorf("Pods in different switches while adding Pod %s to chain %s. Error in rollback configuration: %v.", lpp.PodName, key, err2)
				}
				return fmt.Errorf("Pods in different switches while adding Pod %s to chain %s.", lpp.PodName, key)
			}

			if err := c.ovnClient.CreateLogicalPortPair(lsName, ovs.PodNameToPortName(pod.Name, pod.Namespace), namespace+"_"+name+"_lpp_"+lpp.PodName+"_"+strconv.Itoa(j), lpp.Weight); err != nil {
				if err2 := c.reverseChanges("", created_lpc, created_lpp, created_lppg); err2 != nil {
					return fmt.Errorf("%v Error in rollback configuration: %v.", err, err2)
				}
				return err
			}
			created_lpp = append(created_lpp, namespace+"_"+name+"_lpp_"+lpp.PodName+"_"+strconv.Itoa(j))

			if err := c.ovnClient.AddToLogicalPortPairGroup(namespace+"_"+name+"_lppg_"+strconv.Itoa(i), namespace+"_"+name+"_lpp_"+lpp.PodName+"_"+strconv.Itoa(j)); err != nil {
				if err2 := c.reverseChanges("", created_lpc, created_lpp, created_lppg); err2 != nil {
					return fmt.Errorf("%v Error in rollback configuration: %v.", err, err2)
				}
				return err
			}
		}
	}

	priority := sfc.Spec.Match.Priority
	if priority > 32767 || priority < 0 {
		if err2 := c.reverseChanges("", created_lpc, created_lpp, created_lppg); err2 != nil {
			return fmt.Errorf("Priority of chain %s is larger than 32767 or less than 0. Error in rollback configuration: %v.", key, err2)
		}
		return fmt.Errorf("Priority of chain %s is larger than 32767 or less than 0.", key)
	}

	var sourcePodPort string
	var destPodPort string
	if sfc.Spec.Match.SourcePod != "" {
		sourcePod, err := c.podsLister.Pods(namespace).Get(sfc.Spec.Match.SourcePod)
		if err != nil {
			if k8serrors.IsNotFound(err) {
				if err2 := c.reverseChanges("", created_lpc, created_lpp, created_lppg); err2 != nil {
					return fmt.Errorf("Do not found Pod with name %s in the matching field of chain %s. Error in rollback configuration: %v.", sfc.Spec.Match.SourcePod, key, err2)
				}
				return fmt.Errorf("Do not found Pod with name %s in the matching field of chain %s", sfc.Spec.Match.SourcePod, key)
			}
			if err2 := c.reverseChanges("", created_lpc, created_lpp, created_lppg); err2 != nil {
				return fmt.Errorf("%v Error in rollback configuration: %v.", err, err2)
			}
			return err
		}
		sourcePodSwitch, lsExist := sourcePod.Annotations[util.LogicalSwitchAnnotation]
		if !lsExist {
			if err2 := c.reverseChanges("", created_lpc, created_lpp, created_lppg); err2 != nil {
				return fmt.Errorf("Pod %s in the matching field of chain %s does not have switch name in annotation. Error in rollback configuration: %v.", sfc.Spec.Match.SourcePod, key, err2)
			}
			return fmt.Errorf("Pod %s in the matching field of chain %s does not have switch name in annotation.", sfc.Spec.Match.SourcePod, key)
		}
		if sourcePodSwitch != lsName_main {
			if err2 := c.reverseChanges("", created_lpc, created_lpp, created_lppg); err2 != nil {
				return fmt.Errorf("Pod %s in the matching field of chain %s does not have the same switch name in chain. Error in rollback configuration: %v.", sfc.Spec.Match.SourcePod, key, err2)
			}
			return fmt.Errorf("Pod %s in the matching field of chain %s does not have the same switch name in chain.", sfc.Spec.Match.SourcePod, key)
		}
		sourcePodPort = ovs.PodNameToPortName(sourcePod.Name, sourcePod.Namespace)
	}
	if sfc.Spec.Match.DestinationPod != "" {
		destPod, err := c.podsLister.Pods(namespace).Get(sfc.Spec.Match.DestinationPod)
		if err != nil {
			if k8serrors.IsNotFound(err) {
				if err2 := c.reverseChanges("", created_lpc, created_lpp, created_lppg); err2 != nil {
					return fmt.Errorf("Do not found Pod with name %s in the matching field of chain %s. Error in rollback configuration: %v.", sfc.Spec.Match.DestinationPod, key, err2)
				}
				return fmt.Errorf("Do not found Pod with name %s in the matching field of chain %s", sfc.Spec.Match.DestinationPod, key)
			}
			if err2 := c.reverseChanges("", created_lpc, created_lpp, created_lppg); err2 != nil {
				return fmt.Errorf("%v Error in rollback configuration: %v.", err, err2)
			}
			return err
		}

		destPodSwitch, lsExist := destPod.Annotations[util.LogicalSwitchAnnotation]
		if !lsExist {
			if err2 := c.reverseChanges("", created_lpc, created_lpp, created_lppg); err2 != nil {
				return fmt.Errorf("Pod %s in the matching field of chain %s does not have switch name in annotation. Error in rollback configuration: %v.", sfc.Spec.Match.DestinationPod, key, err2)
			}
			return fmt.Errorf("Pod %s in the matching field of chain %s does not have switch name in annotation.", sfc.Spec.Match.DestinationPod, key)
		}
		if destPodSwitch != lsName_main {
			if err2 := c.reverseChanges("", created_lpc, created_lpp, created_lppg); err2 != nil {
				return fmt.Errorf("Pod %s in the matching field of chain %s does not have the same switch name in chain. Error in rollback configuration: %v.", sfc.Spec.Match.DestinationPod, key, err2)
			}
			return fmt.Errorf("Pod %s in the matching field of chain %s does not have the same switch name in chain.", sfc.Spec.Match.DestinationPod, key)
		}
		destPodPort = ovs.PodNameToPortName(destPod.Name, destPod.Namespace)

	}

	match_str := ""

	if sfc.Spec.Match.SourceIP != "" {
		match_str += "ip4.src == " + sfc.Spec.Match.SourceIP
	}
	if sfc.Spec.Match.DestinationIP != "" {
		if match_str != "" {
			match_str += " && "
		}
		match_str += "ip4.dst == " + sfc.Spec.Match.DestinationIP
	}
	if sfc.Spec.Match.Protocol != "" {
		proForCmp := strings.Replace(strings.ToLower(sfc.Spec.Match.Protocol), " ", "", -1)
		if proForCmp == "tcp" {
			if match_str != "" {
				match_str += " && "
			}
			match_str += "tcp"
			if sfc.Spec.Match.SourcePort != 0 {
				match_str += " && tcp.src == " + strconv.FormatUint(uint64(sfc.Spec.Match.SourcePort), 10)
			}
			if sfc.Spec.Match.DestinationPort != 0 {
				match_str += " && tcp.dst == " + strconv.FormatUint(uint64(sfc.Spec.Match.DestinationPort), 10)
			}

		} else if proForCmp == "udp" {
			if match_str != "" {
				match_str += " && "
			}
			match_str += "udp"
			if sfc.Spec.Match.SourcePort != 0 {
				match_str += " && udp.src == " + strconv.FormatUint(uint64(sfc.Spec.Match.SourcePort), 10)
			}
			if sfc.Spec.Match.DestinationPort != 0 {
				match_str += " && udp.dst == " + strconv.FormatUint(uint64(sfc.Spec.Match.DestinationPort), 10)
			}

		} else if proForCmp == "sctp" {
			if match_str != "" {
				match_str += " && "
			}
			match_str += "sctp"
			if sfc.Spec.Match.SourcePort != 0 {
				match_str += " && sctp.src == " + strconv.FormatUint(uint64(sfc.Spec.Match.SourcePort), 10)
			}
			if sfc.Spec.Match.DestinationPort != 0 {
				match_str += " && sctp.dst == " + strconv.FormatUint(uint64(sfc.Spec.Match.DestinationPort), 10)
			}
		}
	}

	if sfc.Spec.Match.SourceMAC != "" {
		if match_str != "" {
			match_str += " && "
		}
		match_str += "eth.src == " + sfc.Spec.Match.SourceMAC
	}
	if sfc.Spec.Match.DestinationMAC != "" {
		if match_str != "" {
			match_str += " && "
		}
		match_str += "eth.dst == " + sfc.Spec.Match.DestinationMAC
	}

	if sfc.Spec.Match.Others != "" {
		if match_str != "" {
			match_str += " && "
		}
		match_str += sfc.Spec.Match.Others
	}

	if sourcePodPort == "" && destPodPort == "" && match_str == "" {
		if err2 := c.reverseChanges("", created_lpc, created_lpp, created_lppg); err2 != nil {
			return fmt.Errorf("Match field of sfc %s is empty. Error in rollback configuration: %v.", key, err2)
		}
		return fmt.Errorf("Match field of sfc %s is empty.", key)
	}

	// add classifer
	if err := c.ovnClient.AddChainClassifier(lsName_main, namespace+"_"+name+"_chain", match_str, sourcePodPort, destPodPort, namespace+"_"+name, strconv.FormatInt(priority, 10)); err != nil {
		if err2 := c.reverseChanges("", created_lpc, created_lpp, created_lppg); err2 != nil {
			return fmt.Errorf("%v Error in rollback configuration: %v.", err, err2)
		}
		return err
	}

	// update status
	status := kubeovnv1.ServiceFunctionChainStatus{
		SwitchName:                lsName_main,
		ChainName:                 created_lpc,
		LogicalPortPairGroupNames: created_lppg,
		LogicalPortPairNames:      created_lpp,
	}
	if !reflect.DeepEqual(sfc.Status, status) {
		sfc.Status = status
		_, err := c.config.KubeOvnClient.KubeovnV1().ServiceFunctionChains(namespace).UpdateStatus(sfc)
		if err != nil {
			if err2 := c.reverseChanges(namespace+"_"+name, created_lpc, created_lpp, created_lppg); err2 != nil {
				return fmt.Errorf("%v Error in rollback configuration: %v.", err, err2)
			}
			return err
		}
	}
	return nil
}

func (c *Controller) handleUpdateSfc(key UpdateQueueItem) error {
	namespace, name, err := cache.SplitMetaNamespaceKey(key.NamespaceName)
	if err != nil {
		utilruntime.HandleError(fmt.Errorf("invalid resource key: %s", key.NamespaceName))
		return nil
	}
	sfc, err := c.sfcsLister.ServiceFunctionChains(namespace).Get(name)
	if err != nil {
		if k8serrors.IsNotFound(err) {
			return nil
		}
		return err
	}

	var matchDiff bool
	var chainDiff bool

	if !reflect.DeepEqual(key.NewSpec.Match, key.OldSpec.Match) {
		matchDiff = true
	} else {
		matchDiff = false
	}

	if !reflect.DeepEqual(key.NewSpec.Chain, key.OldSpec.Chain) {
		chainDiff = true
	} else {
		chainDiff = false
	}

	if matchDiff && chainDiff {
		if err2 := c.reverseChanges(namespace+"_"+name, sfc.Status.ChainName, sfc.Status.LogicalPortPairNames, sfc.Status.LogicalPortPairGroupNames); err2 != nil {
			return fmt.Errorf("Error in deleting sfc %s while updating: %v.", key.NamespaceName, err2)
		}
		c.addSfcQueue.Add(key.NamespaceName)
	} else if matchDiff {

		lsName_main := sfc.Status.SwitchName

		priority := sfc.Spec.Match.Priority
		if priority > 32767 || priority < 0 {
			return fmt.Errorf("Priority of chain %s is larger than 32767 or less than 0.", key.NamespaceName)
		}

		var sourcePodPort string
		var destPodPort string
		if sfc.Spec.Match.SourcePod != "" {
			sourcePod, err := c.podsLister.Pods(namespace).Get(sfc.Spec.Match.SourcePod)
			if err != nil {
				if k8serrors.IsNotFound(err) {
					return fmt.Errorf("Do not found Pod with name %s in the matching field of chain %s", sfc.Spec.Match.SourcePod, key.NamespaceName)
				}
				return err
			}
			sourcePodSwitch, lsExist := sourcePod.Annotations[util.LogicalSwitchAnnotation]
			if !lsExist {
				return fmt.Errorf("Pod %s in the matching field of chain %s does not have switch name in annotation.", sfc.Spec.Match.SourcePod, key.NamespaceName)
			}
			if sourcePodSwitch != lsName_main {
				return fmt.Errorf("Pod %s in the matching field of chain %s does not have the same switch name in chain.", sfc.Spec.Match.SourcePod, key.NamespaceName)
			}
			sourcePodPort = ovs.PodNameToPortName(sourcePod.Name, sourcePod.Namespace)
		}
		if sfc.Spec.Match.DestinationPod != "" {
			destPod, err := c.podsLister.Pods(namespace).Get(sfc.Spec.Match.DestinationPod)
			if err != nil {
				if k8serrors.IsNotFound(err) {
					return fmt.Errorf("Do not found Pod with name %s in the matching field of chain %s", sfc.Spec.Match.DestinationPod, key.NamespaceName)
				}
				return err
			}

			destPodSwitch, lsExist := destPod.Annotations[util.LogicalSwitchAnnotation]
			if !lsExist {
				return fmt.Errorf("Pod %s in the matching field of chain %s does not have switch name in annotation.", sfc.Spec.Match.DestinationPod, key.NamespaceName)
			}
			if destPodSwitch != lsName_main {
				return fmt.Errorf("Pod %s in the matching field of chain %s does not have the same switch name in chain.", sfc.Spec.Match.DestinationPod, key.NamespaceName)
			}
			destPodPort = ovs.PodNameToPortName(destPod.Name, destPod.Namespace)

		}

		match_str := ""
		if sfc.Spec.Match.SourceIP != "" {
			match_str += "ip4.src == " + sfc.Spec.Match.SourceIP
		}
		if sfc.Spec.Match.DestinationIP != "" {
			if match_str != "" {
				match_str += " && "
			}
			match_str += "ip4.dst == " + sfc.Spec.Match.DestinationIP
		}
		if sfc.Spec.Match.Protocol != "" {
			proForCmp := strings.Replace(strings.ToLower(sfc.Spec.Match.Protocol), " ", "", -1)
			if proForCmp == "tcp" {
				if match_str != "" {
					match_str += " && "
				}
				match_str += "tcp"
				if sfc.Spec.Match.SourcePort != 0 {
					match_str += " && tcp.src == " + strconv.FormatUint(uint64(sfc.Spec.Match.SourcePort), 10)
				}
				if sfc.Spec.Match.DestinationPort != 0 {
					match_str += " && tcp.dst == " + strconv.FormatUint(uint64(sfc.Spec.Match.DestinationPort), 10)
				}

			} else if proForCmp == "udp" {
				if match_str != "" {
					match_str += " && "
				}
				match_str += "udp"
				if sfc.Spec.Match.SourcePort != 0 {
					match_str += " && udp.src == " + strconv.FormatUint(uint64(sfc.Spec.Match.SourcePort), 10)
				}
				if sfc.Spec.Match.DestinationPort != 0 {
					match_str += " && udp.dst == " + strconv.FormatUint(uint64(sfc.Spec.Match.DestinationPort), 10)
				}

			} else if proForCmp == "sctp" {
				if match_str != "" {
					match_str += " && "
				}
				match_str += "sctp"
				if sfc.Spec.Match.SourcePort != 0 {
					match_str += " && sctp.src == " + strconv.FormatUint(uint64(sfc.Spec.Match.SourcePort), 10)
				}
				if sfc.Spec.Match.DestinationPort != 0 {
					match_str += " && sctp.dst == " + strconv.FormatUint(uint64(sfc.Spec.Match.DestinationPort), 10)
				}
			}
		}

		if sfc.Spec.Match.SourceMAC != "" {
			if match_str != "" {
				match_str += " && "
			}
			match_str += "eth.src == " + sfc.Spec.Match.SourceMAC
		}
		if sfc.Spec.Match.DestinationMAC != "" {
			if match_str != "" {
				match_str += " && "
			}
			match_str += "eth.dst == " + sfc.Spec.Match.DestinationMAC
		}

		if sfc.Spec.Match.Others != "" {
			if match_str != "" {
				match_str += " && "
			}
			match_str += sfc.Spec.Match.Others
		}

		if sourcePodPort == "" && destPodPort == "" && match_str == "" {
			return fmt.Errorf("Match field of sfc %s is empty.", key.NamespaceName)
		}

		// del and add classifer
		if err := c.ovnClient.DeleteChainClassifier(namespace + "_" + name); err != nil {
			return err
		}
		if err := c.ovnClient.AddChainClassifier(lsName_main, namespace+"_"+name+"_chain", match_str, sourcePodPort, destPodPort, namespace+"_"+name, strconv.FormatInt(priority, 10)); err != nil {
			c.deleteSfcQueue.Add(key.NamespaceName)
			return err
		}

	} else if chainDiff {
		if err2 := c.reverseChanges(namespace+"_"+name, sfc.Status.ChainName, sfc.Status.LogicalPortPairNames, sfc.Status.LogicalPortPairGroupNames); err2 != nil {
			return fmt.Errorf("Error in deleting sfc %s while updating: %v.", key.NamespaceName, err2)
		}
		c.addSfcQueue.Add(key.NamespaceName)
	}

	return nil
}

func (c *Controller) handleDelSfc(key DeleteQueueItem) error {
	namespace, name, err := cache.SplitMetaNamespaceKey(key.NamespaceName)
	if err != nil {
		utilruntime.HandleError(fmt.Errorf("invalid resource key: %s", key.NamespaceName))
		return nil
	}
	if err2 := c.reverseChanges(namespace+"_"+name, key.Status.ChainName, key.Status.LogicalPortPairNames, key.Status.LogicalPortPairGroupNames); err2 != nil {
		return fmt.Errorf("Error in deleting sfc %s: %v.", key.NamespaceName, err2)
	}

	return nil
}
