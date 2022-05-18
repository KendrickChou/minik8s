package endpoint

import (
	"k8s.io/klog"
	v1 "minik8s.com/minik8s/pkg/api/v1"
	"minik8s.com/minik8s/pkg/apiclient"
	"minik8s.com/minik8s/pkg/controller/component"
)

type EndpointController struct {
	podInformer      *component.Informer
	endpointInformer *component.Informer
	serviceInformer  *component.Informer
	queue            component.WorkQueue
}

func NewEndpointController(podInfo *component.Informer, servInfo *component.Informer,
	epInfo *component.Informer) *EndpointController {
	return &EndpointController{
		podInformer:      podInfo,
		endpointInformer: epInfo,
		serviceInformer:  servInfo,
	}
}

func (epc *EndpointController) Run() {
	epc.queue.Init()

	for !(epc.endpointInformer.HasSynced() && epc.podInformer.HasSynced() && epc.serviceInformer.HasSynced()) {
	}

	epc.podInformer.AddEventHandler(component.EventHandler{
		OnAdd:    epc.addPod,
		OnDelete: epc.deletePod,
		OnUpdate: epc.updatePod,
	})

	epc.serviceInformer.AddEventHandler(component.EventHandler{
		OnAdd:    epc.addService,
		OnDelete: epc.deleteService,
		OnUpdate: epc.updateService,
	})

	epc.syncAll()
	go epc.worker()
}

func (epc *EndpointController) syncAll() {
	services := epc.serviceInformer.List()
	for _, item := range services {
		service := item.(v1.Service)
		epc.queue.Push(service.UID)
	}
}

func (epc *EndpointController) worker() {
	for epc.processNextWorkItem() {
	}
}

func (epc *EndpointController) processNextWorkItem() bool {
	key := epc.queue.Fetch().(string)

	item := epc.serviceInformer.GetItem(key)
	if item == nil {
		klog.Warning("item " + key + " not found\n")
		return true
	}

	service := item.(v1.Service)
	klog.Info("processing Service ", service.Name)
	err := epc.syncEndpoint(service)
	if err != nil {
		klog.Error("syncEndpoint error\n")
		return false
	}
	return true
}

func (epc *EndpointController) syncEndpoint(service v1.Service) error {
	var relatedPods []v1.Pod
	allPods := epc.podInformer.List()
	if service.Spec.Selector != nil {
		for _, podObj := range allPods {
			pod := podObj.(v1.Pod)
			podStatus := component.GetPodStatusObject(&pod)
			klog.Info(podStatus)
			if podStatus == nil || podStatus.(v1.PodStatus).PodIP == "" {
				continue
			} else {
				pod.Status = podStatus.(v1.PodStatus)
			}

			if v1.MatchLabels(service.Spec.Selector, pod.Labels) {
				relatedPods = append(relatedPods, pod)
				if v1.CheckOwner(pod.OwnerReferences, service.UID) == -1 {
					newOwner := v1.OwnerReference{
						Name:       service.Name,
						UID:        service.UID,
						APIVersion: service.APIVersion,
						Kind:       service.Kind,
					}
					pod.OwnerReferences = append(pod.OwnerReferences, newOwner)
					epc.podInformer.UpdateItem(pod.UID, pod)
				}
			}
		}
	}

	if len(relatedPods) == 0 {
		klog.Info("Service ", service.Name, " has no related pods\n")
		return nil
	}

	// check if there are existing endpoints
	eps := epc.endpointInformer.List()
	UID := ""
	for _, obj := range eps {
		ep := obj.(v1.Endpoint)
		if ep.Name == service.Name {
			UID = ep.UID
			epc.endpointInformer.DeleteItem(ep.UID)
			break
		}
	}

	// create new endpoint
	epc.createEndpoint(service, relatedPods, UID)
	return nil
}

func (epc *EndpointController) createEndpoint(service v1.Service, pods []v1.Pod, prevID string) {
	var endpoint v1.Endpoint
	endpoint.Kind = "Endpoint"
	endpoint.APIVersion = service.APIVersion
	endpoint.ObjectMeta.Name = service.ObjectMeta.Name
	endpoint.UID = prevID
	endpoint.ServiceIp = service.Spec.ClusterIP

	endpoint.ServiceIp = "1.2.3.4"

	var subset v1.EndpointSubset

	ipNum := len(pods)
	addresses := make([]v1.EndpointAddress, ipNum)
	for i := 0; i < ipNum; i++ {
		addresses[i].IP = pods[i].Status.PodIP
	}
	subset.Addresses = addresses

	portsNum := len(service.Spec.Ports)
	ports := make([]v1.EndpointPort, portsNum)
	for i := 0; i < portsNum; i++ {
		ports[i].Name = service.Spec.Ports[i].Name
		ports[i].Port = service.Spec.Ports[i].TargetPort
		ports[i].ServicePort = service.Spec.Ports[i].Port
		if service.Spec.Ports[i].Protocol == "" {
			ports[i].Protocol = "tcp"
		} else {
			ports[i].Protocol = service.Spec.Ports[i].Protocol
		}
	}
	subset.Ports = ports

	endpoint.Subset = subset
	if apiclient.PostEndpoint(endpoint) {
		epc.endpointInformer.AddItem(endpoint.UID, endpoint)
	}
}

// get Service by OwnerReferences
func (epc *EndpointController) getPodOwnerService(pod *v1.Pod) []v1.Service {
	var result []v1.Service
	for _, owner := range pod.OwnerReferences {
		service := epc.serviceInformer.GetItem(owner.UID)
		if service != nil {
			result = append(result, service.(v1.Service))
		}
	}
	return result
}

func (epc *EndpointController) getPodMatchService(pod *v1.Pod) []v1.Service {
	services := epc.serviceInformer.List()
	result := make([]v1.Service, 0)

	for _, item := range services {
		service := item.(v1.Service)

		if v1.MatchLabels(service.Labels, pod.Labels) {
			result = append(result, service)
		}
	}

	return result
}

func (epc *EndpointController) enqueueService(service v1.Service) {
	key := service.UID
	epc.queue.Push(key)
}

func (epc *EndpointController) addService(obj any) {
	service := obj.(v1.Service)
	klog.Info("Add Service", service)
	epc.enqueueService(service)
}

func (epc *EndpointController) updateService(newObj any, oldObj any) {
	service := newObj.(v1.Service)
	epc.enqueueService(service)
}

func (epc *EndpointController) deleteService(obj any) {
	service := obj.(v1.Service)
	epc.enqueueService(service)
}

func (epc *EndpointController) addPod(obj any) {
	pod := obj.(v1.Pod)
	klog.Info("Add Pod", pod)
	if pod.Status.PodIP == "" {
		return
	}

	services := epc.getPodMatchService(&pod)
	for _, service := range services {
		epc.enqueueService(service)
	}
}

func (epc *EndpointController) updatePod(newObj any, oldObj any) {
	newPod := newObj.(v1.Pod)
	oldPod := oldObj.(v1.Pod)

	if newPod.Status.PodIP == "" {
		return
	}

	// TODO: this kind of status ain't right
	if newPod.Status.PodIP != oldPod.Status.PodIP {
		services := epc.getPodMatchService(&oldPod)
		for _, service := range services {
			epc.enqueueService(service)
		}
	}
}

func (epc *EndpointController) deletePod(obj any) {
	pod := obj.(v1.Pod)
	if pod.Status.PodIP == "" {
		return
	}

	services := epc.getPodMatchService(&pod)
	for _, service := range services {
		epc.enqueueService(service)
	}
}
