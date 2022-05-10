package rs

import (
	"fmt"
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

	rsc.worker()
}

func (rsc *ReplicaSetController) worker() {
	for rsc.processNextWorkItem() {
	}
}

func (rsc *ReplicaSetController) processNextWorkItem() bool {
	key := rsc.queue.Get().(string)
	defer rsc.queue.Done(key)

	err := rsc.syncReplicaSet(key)
	if err != nil {
		klog.Error("syncReplicaSet error\n")
	}
	return true
}

func (rsc *ReplicaSetController) syncReplicaSet(key string) error {
	pods := rsc.podInformer.List()
	rs := rsc.rsInformer.GetItem(key).(v1.ReplicaSet)

	// TODO: sync logic
	for _, item := range pods {
		pod := item.(v1.Pod)
		var own, labelMatch bool

		// check ownerReferences
		for _, ownerRef := range pod.OwnerReferences {
			own = ownerRef.UID == rs.UID
		}

		labelMatch = v1.MatchSelector(rs.Spec.Selector, pod.Labels)

		if own && !labelMatch {
			// TODO: change pod ownerReference
		}

		// TODO: handle cases when labels match but ownerReference not set
	}

	// following codes are just for template use
	fmt.Println(pods)

	var requestedPod v1.Pod
	requestedPod.Kind = "Pod"
	requestedPod.APIVersion = rs.APIVersion
	requestedPod.Spec = rs.Spec.Template.Spec
	requestedPod.ObjectMeta = rs.Spec.Template.ObjectMeta

	for i := 0; i < rs.Spec.Replicas; i++ {
		apiclient.TemplateArrangePodToNode(requestedPod)
	}

	return nil
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

func (rsc *ReplicaSetController) enqueueRS(rs v1.ReplicaSet) {
	key := rs.UID
	rsc.queue.Add(key)
}

/*
addRS 在有 ReplicaSet 被创建（Informer 发现从前未出现过的 ReplicaSet）时被调用。
*/
func (rsc *ReplicaSetController) addRS(obj any) {
	rs := obj.(v1.ReplicaSet)
	rsc.enqueueRS(rs)
}

func (rsc *ReplicaSetController) updateRS(newObj any, oldObj any) {
	newRS := newObj.(v1.ReplicaSet)

	rsc.enqueueRS(newRS)
}

func (rsc *ReplicaSetController) deleteRS(obj any) {
	rs := obj.(v1.ReplicaSet)
	rsc.enqueueRS(rs)
}

func (rsc *ReplicaSetController) addPod(obj any) {
	pod := obj.(v1.Pod)

	rss := rsc.getPodReplicaSet(&pod)

	for _, rs := range rss {
		rsc.enqueueRS(rs)
	}
}

func (rsc *ReplicaSetController) updatePod(newObj, oldObj any) {
	newPod := newObj.(v1.Pod)

	rss := rsc.getPodReplicaSet(&newPod)

	for _, rs := range rss {
		rsc.enqueueRS(rs)
	}
}

func (rsc *ReplicaSetController) deletePod(obj any) {
	pod := obj.(v1.Pod)

	rss := rsc.getPodReplicaSet(&pod)

	for _, rs := range rss {
		rsc.enqueueRS(rs)
	}
}
