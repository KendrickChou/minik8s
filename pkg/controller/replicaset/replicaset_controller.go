package rs

import (
	"k8s.io/klog"
	v1 "minik8s.com/minik8s/pkg/api/v1"
	"minik8s.com/minik8s/pkg/apiclient"
	"minik8s.com/minik8s/pkg/controller/component"
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
	for !(rsc.rsInformer.HasSynced() && rsc.podInformer.HasSynced()) {
	}

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

	err := rsc.syncReplicaSet(key)
	if err != nil {
		klog.Error("syncReplicaSet error\n")
	}
	return true
}

func (rsc *ReplicaSetController) syncReplicaSet(key string) error {
	pods := rsc.podInformer.List()
	rsItem := rsc.rsInformer.GetItem(key)
	if rsItem == nil {
		return nil
	}
	rs := rsItem.(v1.ReplicaSet)
	replicaNum := 0

	matchedNotOwnedPods := make([]v1.Pod, 0)
	ownedPods := make([]v1.Pod, 0)
	for _, item := range pods {
		pod := item.(v1.Pod)

		if v1.CheckOwner(pod.OwnerReferences, rs.UID) >= 0 {
			ownedPods = append(ownedPods, pod)
			replicaNum++
		} else if v1.MatchSelector(rs.Spec.Selector, pod.Labels) {
			matchedNotOwnedPods = append(matchedNotOwnedPods, pod)
		}
	}

	rs.Status.Replicas = replicaNum

	if replicaNum == rs.Spec.Replicas {
		return nil
	} else {
		if replicaNum < rs.Spec.Replicas {
			rsc.increaseReplica(replicaNum, &rs, matchedNotOwnedPods)
		} else {
			rsc.decreaseReplica(replicaNum, &rs, ownedPods)
		}
	}

	return nil
}

func (rsc *ReplicaSetController) decreaseReplica(realReplicaNum int, rs *v1.ReplicaSet, ownedPods []v1.Pod) {
	for i := 0; i < rs.Spec.Replicas-realReplicaNum; i++ {
		pod := ownedPods[i]
		index := v1.CheckOwner(pod.OwnerReferences, rs.UID)
		pod.OwnerReferences = append(pod.OwnerReferences[:index], pod.OwnerReferences[index+1:]...)
		if apiclient.UpdatePod(&pod) {
			rsc.podInformer.UpdateItem(pod.UID, pod)
			rs.Status.Replicas--
		} else {
			klog.Error("Update Pod Error")
		}
	}

	if apiclient.UpdateReplicaSet(rs) {
		rsc.rsInformer.UpdateItem(rs.UID, *rs)
	}
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

		if apiclient.UpdatePod(&pod) {
			rsc.podInformer.UpdateItem(pod.UID, pod)
			rs.Status.Replicas++
		} else {
			klog.Error("Update Pod Error")
		}
	}

	for i := 0; i < expectedIncrease-length; i++ {
		pod := v1.Pod{
			ObjectMeta: rs.Spec.Template.ObjectMeta,
			Spec:       rs.Spec.Template.Spec,
		}
		pod.Kind = "Pod"
		pod.APIVersion = rs.APIVersion

		ref := v1.OwnerReference{
			Name:       rs.Name,
			APIVersion: rs.APIVersion,
			UID:        rs.UID,
			Kind:       rs.Kind,
		}
		pod.OwnerReferences = append(pod.OwnerReferences, ref)

		uid := apiclient.PostPod(&pod)
		if uid != "" {
			pod.UID = uid
			rsc.podInformer.AddItem(uid, pod)
			rs.Status.Replicas++
		} else {
			klog.Error("Post Pod Error")
		}
	}

	if apiclient.UpdateReplicaSet(rs) {
		rsc.rsInformer.UpdateItem(rs.UID, *rs)
	}
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
	rss := rsc.rsInformer.List()

	for _, item := range rss {
		rs := item.(v1.ReplicaSet)
		if v1.CheckOwner(pod.OwnerReferences, rs.UID) >= 0 {
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
	rsc.enqueueRS(&rs)
}

func (rsc *ReplicaSetController) updateRS(newObj any, oldObj any) {
	newRS := newObj.(v1.ReplicaSet)
	oldRS := oldObj.(v1.ReplicaSet)

	// replica num changes
	if newRS.Spec.Replicas != oldRS.Spec.Replicas {
		rsc.enqueueRS(&newRS)
	}

	// ReplicaSet Selector changes
	if len(newRS.Spec.Selector.MatchLabels) != len(oldRS.Spec.Selector.MatchLabels) ||
		!v1.MatchLabels(newRS.Spec.Selector.MatchLabels, oldRS.Spec.Selector.MatchLabels) {
		rsc.enqueueRS(&newRS)
	}
}

func (rsc *ReplicaSetController) deleteRS(obj any) {
	rs := obj.(v1.ReplicaSet)

	pods := rsc.podInformer.List()
	for _, item := range pods {
		pod := item.(v1.Pod)
		ownerIndex := v1.CheckOwner(pod.OwnerReferences, rs.UID)

		if ownerIndex >= 0 {
			pod.OwnerReferences = append(pod.OwnerReferences[:ownerIndex], pod.OwnerReferences[ownerIndex+1:]...)
		}
	}
}

func (rsc *ReplicaSetController) addPod(obj any) {
	pod := obj.(v1.Pod)

	rss := rsc.getPodReplicaSet(&pod)

	for _, rs := range rss {
		rsc.enqueueRS(&rs)
	}
}

func (rsc *ReplicaSetController) updatePod(newObj, oldObj any) {
	newPod := newObj.(v1.Pod)
	oldPod := oldObj.(v1.Pod)

	rs := rsc.getPodOwnerReplicaSet(&oldPod)
	if rs == nil {
		rss := rsc.getPodReplicaSet(&newPod)

		for _, tmpRS := range rss {
			rsc.enqueueRS(&tmpRS)
		}

		return
	}

	if len(newPod.Labels) != len(oldPod.Labels) ||
		!v1.MatchLabels(newPod.Labels, oldPod.Labels) {
		rsc.enqueueRS(rs)
	}

	newPodStatusObj := component.GetPodStatusObject(&newPod)
	oldPodStatusObj := component.GetPodStatusObject(&oldPod)
	if newPodStatusObj == nil || oldPodStatusObj == nil {
		klog.Error("Can't get status of Pod")
	}

	newPodStatus := newPodStatusObj.(v1.PodStatus)
	oldPodStatus := newPodStatusObj.(v1.PodStatus)
	if !v1.ComparePodStatus(&newPodStatus, &oldPodStatus) {
		rsc.enqueueRS(rs)
	}
}

func (rsc *ReplicaSetController) deletePod(obj any) {
	pod := obj.(v1.Pod)

	rs := rsc.getPodOwnerReplicaSet(&pod)
	if rs != nil {
		rsc.enqueueRS(rs)
	}
}
