package rs

import (
	"k8s.io/klog"
	v1 "minik8s.com/minik8s/pkg/api/v1"
	"minik8s.com/minik8s/pkg/controller/component"
	"minik8s.com/minik8s/utils/random"
)

type ReplicaSetController struct {
	podInformer *component.Informer
	rsInformer  *component.Informer
	queue       component.WorkQueue
}

func NewReplicaSetController(podInfo *component.Informer, rsInfo *component.Informer) *ReplicaSetController {
	return &ReplicaSetController{
		podInformer: podInfo,
		rsInformer:  rsInfo,
	}
}

// Run begins watching and syncing.
func (rsc *ReplicaSetController) Run() {
	rsc.queue.Init()

	rsc.rsInformer.AddEventHandler(component.EventHandler{
		OnAdd:    rsc.addRS,
		OnDelete: rsc.deleteRS,
		OnUpdate: rsc.updateRS,
	})

	rsc.podInformer.AddEventHandler(component.EventHandler{
		OnAdd:    rsc.addPod,
		OnDelete: rsc.deletePod,
		OnUpdate: rsc.updatePod,
	})

	rsc.syncAll()
	go rsc.worker()
}

func (rsc *ReplicaSetController) syncAll() {
	rss := rsc.rsInformer.List()
	for _, item := range rss {
		rs := item.(v1.ReplicaSet)
		rsc.enqueueRS(&rs)
	}
}

func (rsc *ReplicaSetController) worker() {
	for rsc.processNextWorkItem() {
	}
}

func (rsc *ReplicaSetController) processNextWorkItem() bool {
	key := rsc.queue.Fetch().(string)
	if !rsc.queue.Process(key) {
		klog.Infof("ReplicaSet %s is being processed", key)
		return true
	}

	err := rsc.syncReplicaSet(key)
	rsc.queue.Done(key)
	if err != nil {
		klog.Error("syncReplicaSet error\n")
	}
	return true
}

func (rsc *ReplicaSetController) syncReplicaSet(key string) error {
	pods := rsc.podInformer.List()
	rsItem := rsc.rsInformer.GetItem(key)
	if rsItem == nil {
		klog.Error("Can't find replicaSet with uid ", key)
		return nil
	}
	rs := rsItem.(v1.ReplicaSet)
	if rs.Status.Replicas == -1 {
		klog.Infof("ReplicaSet %s has been taken control by HPA", rs.Name)
		return nil
	}

	replicaNum := 0
	matchedNotOwnedPods := make([]v1.Pod, 0)
	ownedPods := make([]v1.Pod, 0)
	for _, item := range pods {
		pod := item.(v1.Pod)
		ownerRS := v1.GetOwnerReplicaSet(&pod)
		if ownerRS == rs.UID {
			ownedPods = append(ownedPods, pod)
			replicaNum++
		} else if ownerRS == "" && v1.MatchSelector(rs.Spec.Selector, pod.Labels) {
			matchedNotOwnedPods = append(matchedNotOwnedPods, pod)
		}
	}

	// ---debug---
	//fmt.Printf("ownedPods: ")
	//for _, pod := range ownedPods {
	//	fmt.Printf("%v ", pod.UID)
	//}
	//fmt.Printf("\nmatchedNotOwnedPods: ")
	//for _, pod := range matchedNotOwnedPods {
	//	fmt.Printf("%v ", pod.UID)
	//}
	//fmt.Printf("\n")

	rs.Status.Replicas = replicaNum
	if replicaNum != rs.Spec.Replicas {
		if replicaNum < rs.Spec.Replicas {
			rsc.increaseReplica(replicaNum, &rs, matchedNotOwnedPods)
		} else {
			rsc.decreaseReplica(replicaNum, &rs, ownedPods)
		}
	}

	return nil
}

func (rsc *ReplicaSetController) decreaseReplica(realReplicaNum int, rs *v1.ReplicaSet, ownedPods []v1.Pod) {
	for i := 0; i < realReplicaNum-rs.Spec.Replicas; i++ {
		pod := ownedPods[i]
		index := v1.CheckOwner(pod.OwnerReferences, rs.UID)
		klog.Infof("remove Pod %s's OwnerReference ReplicaSet %s", pod.UID, rs.UID)
		pod.OwnerReferences = append(pod.OwnerReferences[:index], pod.OwnerReferences[index+1:]...)
		rsc.podInformer.UpdateItem(pod.UID, pod)
		rs.Status.Replicas--
	}

	rsc.rsInformer.UpdateItem(rs.UID, *rs)
}

func (rsc *ReplicaSetController) increaseReplica(realReplicaNum int, rs *v1.ReplicaSet, matchedNotOwnedPods []v1.Pod) {
	podsLen := len(matchedNotOwnedPods)
	expectedIncrease := rs.Spec.Replicas - realReplicaNum
	var length int
	if expectedIncrease > podsLen {
		length = podsLen
	} else {
		length = expectedIncrease
	}

	for i := 0; i < length; i++ {
		pod := matchedNotOwnedPods[i]
		if v1.GetOwnerReplicaSet(&pod) != "" {
			continue
		}

		ref := v1.OwnerReference{
			Name:       rs.Name,
			APIVersion: rs.APIVersion,
			UID:        rs.UID,
			Kind:       rs.Kind,
		}
		pod.OwnerReferences = append(pod.OwnerReferences, ref)

		rsc.podInformer.UpdateItem(pod.UID, pod)
		rs.Status.Replicas++
	}

	for i := 0; i < expectedIncrease-length; i++ {
		pod := v1.Pod{
			ObjectMeta: rs.Spec.Template.ObjectMeta,
			Spec:       rs.Spec.Template.Spec,
		}
		pod.Kind = "Pod"
		pod.APIVersion = rs.APIVersion
		pod.ObjectMeta = rs.Spec.Template.ObjectMeta
		pod.UID = ""
		pod.Name = pod.Name + "-" + random.String(5)

		ref := v1.OwnerReference{
			Name:       rs.Name,
			APIVersion: rs.APIVersion,
			UID:        rs.UID,
			Kind:       rs.Kind,
		}
		pod.OwnerReferences = append(pod.OwnerReferences, ref)

		rsc.podInformer.AddItem(pod)

		rs.Status.Replicas++
	}
	rsc.rsInformer.UpdateItem(rs.UID, *rs)
}

// return replicaSets that matches the pod, while there
// is only one replicaSet actually manage the pod
func (rsc *ReplicaSetController) getPodReplicaSet(pod *v1.Pod) []v1.ReplicaSet {
	rss := rsc.rsInformer.List()
	var result []v1.ReplicaSet

	for _, item := range rss {
		rs := item.(v1.ReplicaSet)

		flag := v1.MatchSelector(rs.Spec.Selector, pod.Labels)

		if flag {
			result = append(result, rs)
		}
	}

	return result
}

func (rsc *ReplicaSetController) getPodOwnerReplicaSet(pod *v1.Pod) *v1.ReplicaSet {
	rsUID := v1.GetOwnerReplicaSet(pod)

	if rsUID != "" {
		rsItem := rsc.rsInformer.GetItem(rsUID)

		if rsItem != nil {
			rs := rsItem.(v1.ReplicaSet)
			return &rs
		}
	}

	return nil
}

func (rsc *ReplicaSetController) enqueueRS(rs *v1.ReplicaSet) {
	key := rs.UID
	rsc.queue.Push(key)
}

func (rsc *ReplicaSetController) addRS(obj any) {
	rs := obj.(v1.ReplicaSet)
	klog.Info("add ReplicaSet ", rs.UID)
	rsc.enqueueRS(&rs)
}

func (rsc *ReplicaSetController) updateRS(newObj any, oldObj any) {
	newRS := newObj.(v1.ReplicaSet)
	oldRS := oldObj.(v1.ReplicaSet)

	// replica num changes
	if newRS.Spec.Replicas != oldRS.Spec.Replicas {
		klog.Info("update ReplicaSet ", newRS.UID)
		rsc.enqueueRS(&newRS)
	}

	// get control back from HPA
	if oldRS.Status.Replicas == -1 && newRS.Status.Replicas >= 0 {
		klog.Info("update ReplicaSet ", newRS.UID)
		rsc.enqueueRS(&newRS)
	}

	// ReplicaSet Selector changes
	if len(newRS.Spec.Selector.MatchLabels) != len(oldRS.Spec.Selector.MatchLabels) ||
		!v1.MatchLabels(newRS.Spec.Selector.MatchLabels, oldRS.Spec.Selector.MatchLabels) {
		klog.Info("update ReplicaSet ", newRS.UID)
		rsc.enqueueRS(&newRS)
	}
}

func (rsc *ReplicaSetController) deleteRS(obj any) {
	rs := obj.(v1.ReplicaSet)
	klog.Info("delete ReplicaSet ", rs.UID)

	pods := rsc.podInformer.List()
	for _, item := range pods {
		pod := item.(v1.Pod)
		ownerIndex := v1.CheckOwner(pod.OwnerReferences, rs.UID)

		if ownerIndex >= 0 {
			pod.OwnerReferences = append(pod.OwnerReferences[:ownerIndex], pod.OwnerReferences[ownerIndex+1:]...)
			rsc.podInformer.UpdateItem(pod.UID, pod)
		}
	}
}

func (rsc *ReplicaSetController) addPod(obj any) {
	pod := obj.(v1.Pod)

	rss := rsc.getPodReplicaSet(&pod)

	for _, rs := range rss {
		klog.Infof("enqueue ReplicaSet %s when add Pod %s", rs.UID, pod.UID)
		rsc.enqueueRS(&rs)
	}
}

func (rsc *ReplicaSetController) updatePod(newObj, oldObj any) {
	newPod := newObj.(v1.Pod)
	oldPod := oldObj.(v1.Pod)

	needAttention := len(newPod.Labels) != len(oldPod.Labels) || !v1.MatchLabels(newPod.Labels, oldPod.Labels) ||
		!v1.ComparePodStatus(&newPod.Status, &oldPod.Status)

	if !needAttention {
		return
	}

	rs := rsc.getPodOwnerReplicaSet(&oldPod)
	if rs == nil {
		rss := rsc.getPodReplicaSet(&newPod)

		for _, tmpRS := range rss {
			klog.Infof("enqueue ReplicaSet %s when update Pod %s", tmpRS.UID, newPod.UID)
			rsc.enqueueRS(&tmpRS)
		}

		return
	} else {
		klog.Infof("enqueue ReplicaSet %s when update Pod %s", rs.UID, newPod.UID)
		rsc.enqueueRS(rs)
	}

}

func (rsc *ReplicaSetController) deletePod(obj any) {
	pod := obj.(v1.Pod)

	rs := rsc.getPodOwnerReplicaSet(&pod)
	if rs != nil {
		klog.Infof("enqueue ReplicaSet %s when delete Pod %s", rs.UID, pod.UID)
		rsc.enqueueRS(rs)
	} else {
		klog.Infof("delete Pod %s, which has no owner ReplicaSet", pod.UID)
	}
}
