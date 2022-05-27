package podautoscaling

import (
	"errors"
	"k8s.io/klog"
	"math"
	v1 "minik8s.com/minik8s/pkg/api/v1"
	"minik8s.com/minik8s/pkg/controller/component"
	"strconv"
	"time"
)

type HorizontalController struct {
	hpaInformer          *component.Informer
	podInformer          *component.Informer
	rsInformer           *component.Informer
	queue                component.WorkQueue
	defaultScaleDownRule *v1.HPAScalingRules
	defaultScaleUpRule   *v1.HPAScalingRules
}

func NewHorizontalController(hpaInf *component.Informer, podInf *component.Informer, rsInf *component.Informer) *HorizontalController {
	upPolicies := make([]v1.HPAScalingPolicy, 2)
	upPolicies[0] = v1.HPAScalingPolicy{
		Type:          v1.PercentScalingPolicy,
		Value:         100,
		PeriodSeconds: 15,
	}
	upPolicies[1] = v1.HPAScalingPolicy{
		Type:          v1.PodsScalingPolicy,
		Value:         4,
		PeriodSeconds: 15,
	}

	downPolicies := make([]v1.HPAScalingPolicy, 1)
	downPolicies[0] = v1.HPAScalingPolicy{
		Type:          v1.PercentScalingPolicy,
		Value:         100,
		PeriodSeconds: 15,
	}

	return &HorizontalController{
		hpaInformer: hpaInf,
		podInformer: podInf,
		rsInformer:  rsInf,
		defaultScaleUpRule: &v1.HPAScalingRules{
			StabilizationWindowSeconds: 0,
			SelectPolicy:               v1.MaxChangePolicySelect,
			Policies:                   upPolicies,
		},
		defaultScaleDownRule: &v1.HPAScalingRules{
			StabilizationWindowSeconds: 300,
			SelectPolicy:               v1.MinChangePolicySelect,
			Policies:                   downPolicies,
		},
	}
}

func (hpaC *HorizontalController) Run() {
	hpaC.queue.Init()

	for !(hpaC.rsInformer.HasSynced() && hpaC.podInformer.HasSynced() && hpaC.hpaInformer.HasSynced()) {
	}

	hpaC.hpaInformer.AddEventHandler(component.EventHandler{
		OnAdd:    hpaC.addHPA,
		OnUpdate: hpaC.updateHPA,
		OnDelete: hpaC.deleteHPA,
	})

	go hpaC.worker()
	go hpaC.periodicallyScaleAll()
}

func (hpaC *HorizontalController) periodicallyScaleAll() {
	for {
		time.Sleep(time.Second * 15)
		klog.Info("periodically scale")
		hpas := hpaC.hpaInformer.List()
		for _, item := range hpas {
			hpa := item.(v1.HorizontalPodAutoscaler)
			hpaC.enqueueHPA(&hpa)
		}
	}
}

func (hpaC *HorizontalController) worker() {
	for hpaC.processNextWorkItem() {

	}
}

func (hpaC *HorizontalController) processNextWorkItem() bool {
	key := hpaC.queue.Fetch().(string)
	if !hpaC.queue.Process(key) {
		klog.Infof("HPA %s is being processed", key)
		return true
	}

	err := hpaC.reconcileAutoScaler(key)
	if err != nil {
		klog.Error(err.Error())
		return false
	}
	return true
}

func (hpaC *HorizontalController) getTargetReplicaSet(hpa *v1.HorizontalPodAutoscaler) *v1.ReplicaSet {
	rss := hpaC.rsInformer.List()
	for _, item := range rss {
		rs := item.(v1.ReplicaSet)
		if rs.Name == hpa.Spec.ScaleTargetRef.Name && rs.APIVersion == hpa.Spec.ScaleTargetRef.APIVersion {
			return &rs
		}
	}

	return nil
}

func (hpaC *HorizontalController) getRSOwnedPods(rs *v1.ReplicaSet) []v1.Pod {
	relatedPods := make([]v1.Pod, 0)
	pods := hpaC.podInformer.List()
	for _, item := range pods {
		pod := item.(v1.Pod)
		if v1.CheckOwner(pod.OwnerReferences, rs.UID) != -1 {
			relatedPods = append(relatedPods, pod)
		}
	}

	return relatedPods
}

func (hpaC *HorizontalController) reconcileAutoScaler(key string) error {
	hpaItem := hpaC.hpaInformer.GetItem(key)
	if hpaItem == nil {
		return errors.New("can't find horizontalPodAutoScaler")
	}
	hpa := hpaItem.(v1.HorizontalPodAutoscaler)

	// get autoscaling target, should be "Replicaset"
	var targetRS *v1.ReplicaSet = nil

	if hpa.Spec.ScaleTargetRef.Kind == "ReplicaSet" {
		targetRS = hpaC.getTargetReplicaSet(&hpa)
		if targetRS == nil {
			return errors.New("can't find target replicaset")
		}

		if targetRS.Status.Replicas != -1 {
			targetRS.Status.Replicas = -1
			owner := v1.OwnerReference{
				Name:       hpa.Name,
				APIVersion: hpa.APIVersion,
				UID:        hpa.UID,
				Kind:       hpa.Kind,
			}
			targetRS.OwnerReferences = append(targetRS.OwnerReferences, owner)
			hpaC.rsInformer.UpdateItem(targetRS.UID, *targetRS)
		}

		relatedPods := hpaC.getRSOwnedPods(targetRS)
		hpa.Status.CurrentReplicas = len(relatedPods)
		return hpaC.autoScaleReplicaSet(&hpa, targetRS, relatedPods)
	} else {
		return errors.New("horizontalPodAutoScaler target at unsupported object")
	}
}

func (hpaC *HorizontalController) autoScaleReplicaSet(hpa *v1.HorizontalPodAutoscaler, rs *v1.ReplicaSet, pods []v1.Pod) error {
	// calculate the expectation
	metrics := hpa.Spec.Metrics

	var expectReplica = 0
	for _, metric := range metrics {
		if metric.Type != v1.ResourceMetricSourceType {
			return errors.New("unsupported metric type")
		}

		totalUtilization := 0.0
		avgUtil := 0.0
		if metric.Resource.Name == "cpu" {
			for _, pod := range pods {
				totalUtilization = totalUtilization + hpaC.calcPodCpuUtilization(&pod)
			}
			avgUtil = totalUtilization
		} else if metric.Resource.Name == "memory" {
			for _, pod := range pods {
				totalUtilization = totalUtilization + hpaC.calcPodMemoryUtilization(&pod)
			}
			avgUtil = totalUtilization / float64(len(pods))
		} else {
			klog.Info("unsupported resource type")
			continue
		}

		if metric.Resource.Target.AverageUtilization == 0 {
			klog.Warning("resource utilization cannot be 0")
			continue
		}

		proportion := avgUtil / float64(metric.Resource.Target.AverageUtilization)
		expectation := int(math.Ceil(float64(hpa.Status.CurrentReplicas) * proportion))
		if expectation > expectReplica {
			expectReplica = expectation
		}
	}

	if expectReplica > hpa.Spec.MaxReplicas {
		expectReplica = hpa.Spec.MaxReplicas
	}

	if expectReplica < hpa.Spec.MinReplicas {
		expectReplica = hpa.Spec.MinReplicas
	}

	hpa.Status.DesiredReplicas = expectReplica

	var rule *v1.HPAScalingRules
	if hpa.Status.CurrentReplicas == expectReplica {
		klog.Infof("%s don't need scaling", rs.Name)
		hpaC.queue.Done(hpa.UID)
		return nil
	} else if hpa.Status.CurrentReplicas > expectReplica {
		// scale down
		if hpa.Spec.Behavior != nil && hpa.Spec.Behavior.ScaleDown != nil {
			rule = hpa.Spec.Behavior.ScaleDown
		} else {
			rule = hpaC.defaultScaleDownRule
		}
	} else {
		// scale up
		if hpa.Spec.Behavior != nil && hpa.Spec.Behavior.ScaleUp != nil {
			rule = hpa.Spec.Behavior.ScaleUp
		} else {
			rule = hpaC.defaultScaleUpRule
		}
	}

	klog.Infof("rule: %v", rule)
	return hpaC.scale(hpa, rule, pods, rs)
}

func (hpaC *HorizontalController) scale(hpa *v1.HorizontalPodAutoscaler, scalingRule *v1.HPAScalingRules, pods []v1.Pod, rs *v1.ReplicaSet) error {
	duration := time.Since(hpa.Status.LastScaleTime)
	if int(duration.Seconds()) < scalingRule.StabilizationWindowSeconds {
		klog.Info("scaling canceled because of stabilization window")
		hpaC.queue.Done(hpa.UID)
		return nil
	}

	if scalingRule.SelectPolicy == v1.DisabledPolicySelect {
		klog.Info("autoscaling disabled")
		hpaC.queue.Done(hpa.UID)
		return nil
	}

	// choose the matched policy
	deltaNumPerMinute := 0.0
	var chosenPolicy v1.HPAScalingPolicy
	for i, policy := range scalingRule.Policies {
		var tmpDeltaNum float64
		if policy.Type == v1.PodsScalingPolicy {
			tmpDeltaNum = float64(policy.Value) * 60.0 / float64(policy.PeriodSeconds)
		} else {
			tmpDeltaNum = float64(policy.Value/100*hpa.Status.CurrentReplicas) * 60.0 / float64(policy.PeriodSeconds)
		}

		if i == 0 {
			deltaNumPerMinute = tmpDeltaNum
			chosenPolicy = policy
		} else if scalingRule.SelectPolicy == v1.MinChangePolicySelect {
			if tmpDeltaNum < deltaNumPerMinute {
				deltaNumPerMinute = tmpDeltaNum
				chosenPolicy = policy
			}
		} else {
			if tmpDeltaNum > deltaNumPerMinute {
				deltaNumPerMinute = tmpDeltaNum
				chosenPolicy = policy
			}
		}
	}

	podUIDs := make([]string, len(pods))
	for i, pod := range pods {
		podUIDs[i] = pod.UID
	}
	go hpaC.periodicallyScale(hpa, &chosenPolicy, podUIDs, rs)

	return nil
}

func (hpaC *HorizontalController) periodicallyScale(hpa *v1.HorizontalPodAutoscaler, policy *v1.HPAScalingPolicy, podUIDs []string, rs *v1.ReplicaSet) {
	max := 0
	if policy.Type == v1.PodsScalingPolicy {
		max = policy.Value
	} else {
		max = policy.Value / 100 * hpa.Status.CurrentReplicas
	}

	klog.Infof("HPA %s, desired: %v, current: %v", hpa.UID, hpa.Status.DesiredReplicas, hpa.Status.CurrentReplicas)
	if hpa.Status.DesiredReplicas < hpa.Status.CurrentReplicas {
		// delete pods
		delta := hpa.Status.CurrentReplicas - hpa.Status.DesiredReplicas
		leftToDelete := delta

		for i := 0; i < delta; {
			numInPeriod := max
			endFlag := false
			if max > leftToDelete {
				numInPeriod = leftToDelete
				endFlag = true
			}

			for j := 0; j < numInPeriod; j++ {
				pod := hpaC.podInformer.GetItem(podUIDs[i])
				if pod == nil {
					continue
				}

				klog.Infof("delete Pod %s", podUIDs[i])
				hpaC.podInformer.DeleteItem(podUIDs[i])
				hpa.Status.CurrentReplicas--
				hpa.Status.LastScaleTime = time.Now()
				i++
			}

			if !endFlag {
				leftToDelete -= max
				time.Sleep(time.Duration(policy.PeriodSeconds) * time.Second)
			}
		}
	} else {
		hpaC.masterCurrentPods(max, policy, hpa, rs)
		// create pods
		hpaC.createPods(max, policy, hpa, rs)
	}

	hpaC.hpaInformer.UpdateItem(hpa.UID, *hpa)
	hpaC.queue.Done(hpa.UID)
}

func (hpaC *HorizontalController) masterCurrentPods(max int, policy *v1.HPAScalingPolicy, hpa *v1.HorizontalPodAutoscaler, rs *v1.ReplicaSet) {
	pods := hpaC.podInformer.List()
	readyPods := make([]v1.Pod, 0)
	for _, item := range pods {
		pod := item.(v1.Pod)
		if v1.GetOwnerReplicaSet(&pod) == "" && v1.MatchSelector(rs.Spec.Selector, pod.Labels) {
			readyPods = append(readyPods, pod)
		}
	}

	totalReady := len(readyPods)
	expectNum := hpa.Status.DesiredReplicas - hpa.Status.CurrentReplicas
	var delta int
	if totalReady < expectNum {
		delta = totalReady
	} else {
		delta = expectNum
	}

	leftToAdd := delta

	for i := 0; i < delta; {
		numInPeriod := max
		endFlag := false
		if max > leftToAdd {
			numInPeriod = leftToAdd
			endFlag = true
		}

		for j := 0; j < numInPeriod; j++ {
			pod := readyPods[i]
			ref := v1.OwnerReference{
				Name:       rs.Name,
				APIVersion: rs.APIVersion,
				UID:        rs.UID,
				Kind:       rs.Kind,
			}
			pod.OwnerReferences = append(pod.OwnerReferences, ref)
			hpaC.podInformer.UpdateItem(pod.UID, pod)

			hpa.Status.CurrentReplicas++
			hpa.Status.LastScaleTime = time.Now()
			i++
		}

		if !endFlag {
			leftToAdd -= max
			time.Sleep(time.Duration(policy.PeriodSeconds) * time.Second)
		}
	}
}

func (hpaC *HorizontalController) createPods(max int, policy *v1.HPAScalingPolicy, hpa *v1.HorizontalPodAutoscaler, rs *v1.ReplicaSet) {
	podTemplate := v1.Pod{
		ObjectMeta: rs.Spec.Template.ObjectMeta,
		Spec:       rs.Spec.Template.Spec,
	}
	podTemplate.Kind = "Pod"
	podTemplate.APIVersion = rs.APIVersion
	podTemplate.ObjectMeta = rs.Spec.Template.ObjectMeta
	podTemplate.UID = ""
	podTemplate.Name = podTemplate.Name + "-"

	ref := v1.OwnerReference{
		Name:       rs.Name,
		APIVersion: rs.APIVersion,
		UID:        rs.UID,
		Kind:       rs.Kind,
	}
	podTemplate.OwnerReferences = append(podTemplate.OwnerReferences, ref)

	delta := hpa.Status.DesiredReplicas - hpa.Status.CurrentReplicas
	leftToAdd := delta

	for i := 0; i < delta; {
		numInPeriod := max
		endFlag := false
		if max > leftToAdd {
			numInPeriod = leftToAdd
			endFlag = true
		}

		for j := 0; j < numInPeriod; j++ {
			podTemplate.Name += strconv.Itoa(i)
			hpaC.podInformer.AddItem(podTemplate)
			hpa.Status.CurrentReplicas++
			hpa.Status.LastScaleTime = time.Now()
			i++
		}

		if !endFlag {
			leftToAdd -= max
			time.Sleep(time.Duration(policy.PeriodSeconds) * time.Second)
		}
	}
}

func (hpaC *HorizontalController) calcPodCpuUtilization(pod *v1.Pod) float64 {
	var utilization float64 = 0
	for _, cs := range pod.Status.ContainerStatuses {
		utilStr := cs.State.CPUPerc
		utilStr = utilStr[0 : len(utilStr)-1]
		cpuUtil, err := strconv.ParseFloat(utilStr, 64)
		if err != nil {
			klog.Warningf("convert string error %s", utilStr)
		}

		utilization = utilization + cpuUtil
	}

	return utilization
}

func (hpaC *HorizontalController) calcPodMemoryUtilization(pod *v1.Pod) float64 {

	var utilization float64 = 0
	for _, cs := range pod.Status.ContainerStatuses {
		utilStr := cs.State.MemPerc
		utilStr = utilStr[0 : len(utilStr)-1]
		memoryUtil, err := strconv.ParseFloat(utilStr, 64)
		if err != nil {
			klog.Warningf("convert string error %s", utilStr)
		}

		utilization = utilization + memoryUtil
	}

	return utilization / float64(len(pod.Status.ContainerStatuses))
}

func (hpaC *HorizontalController) enqueueHPA(hpa *v1.HorizontalPodAutoscaler) {
	key := hpa.UID
	hpaC.queue.Push(key)
}

func (hpaC *HorizontalController) updateHPA(newObj any, oldObj any) {
	newHPA := newObj.(v1.HorizontalPodAutoscaler)
	oldHPA := oldObj.(v1.HorizontalPodAutoscaler)

	if newHPA.Status.LastScaleTime == oldHPA.Status.LastScaleTime &&
		newHPA.Status.CurrentReplicas == oldHPA.Status.CurrentReplicas &&
		newHPA.Status.DesiredReplicas == oldHPA.Status.DesiredReplicas {
		return
	}

	klog.Infof("update HPA %s", newHPA.UID)
	hpaC.enqueueHPA(&newHPA)
}

func (hpaC *HorizontalController) addHPA(obj any) {
	hpa := obj.(v1.HorizontalPodAutoscaler)
	klog.Infof("add HPA %s", hpa.UID)
	hpaC.enqueueHPA(&hpa)
}

func (hpaC *HorizontalController) deleteHPA(obj any) {
	hpa := obj.(v1.HorizontalPodAutoscaler)
	rs := hpaC.getTargetReplicaSet(&hpa)

	rs.Status.Replicas = hpa.Status.CurrentReplicas
	index := v1.CheckOwner(rs.OwnerReferences, hpa.UID)
	rs.OwnerReferences = append(rs.OwnerReferences[index:], rs.OwnerReferences[index+1:]...)
	hpaC.rsInformer.UpdateItem(rs.UID, *rs)
}
