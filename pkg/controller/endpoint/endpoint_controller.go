package endpoint

import (
	"k8s.io/klog"
	v1 "minik8s.com/minik8s/pkg/api/v1"
	"minik8s.com/minik8s/pkg/controller/component"
	"strconv"
	"strings"
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

	if !epc.queue.Process(key) {
		return true
	}
	service := item.(v1.Service)
	err := epc.syncEndpoint(service)
	epc.queue.Done(key)
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
			// check podStatus

			if pod.Status.PodIP == "" {
				klog.Infof("Pod %s has no ip", pod.Name)
				continue
			}

			ownerServiceID := v1.GetOwnerService(pod.OwnerReferences)
			labelMatchFlag := v1.MatchLabels(service.Spec.Selector, pod.Labels)

			if ownerServiceID == "" {
				if labelMatchFlag {
					newOwner := v1.OwnerReference{
						Name:       service.Name,
						UID:        service.UID,
						APIVersion: service.APIVersion,
						Kind:       service.Kind,
					}
					klog.Infof("Pod %s is owned by Service %s now", pod.UID, service.UID)
					pod.OwnerReferences = append(pod.OwnerReferences, newOwner)

					epc.podInformer.UpdateItem(pod.UID, pod)

					relatedPods = append(relatedPods, pod)
				}
			} else if ownerServiceID == service.UID {
				if labelMatchFlag {
					relatedPods = append(relatedPods, pod)
				} else {
					// the pod is owned by the service but the labels don't match the selector
					index := v1.CheckOwner(pod.OwnerReferences, service.UID)
					klog.Infof("Pod %s is not owned by Service %s now", pod.UID, service.UID)
					pod.OwnerReferences = append(pod.OwnerReferences[:index], pod.OwnerReferences[index+1:]...)

					epc.podInformer.UpdateItem(pod.UID, pod)
				}
			}
		}
	}

	noRelated := len(relatedPods) == 0
	if noRelated {
		klog.Info("Service ", service.Name, " has no related pods\n")
	}

	// check if there are existing endpoints
	eps := epc.endpointInformer.List()
	UID := ""
	for _, obj := range eps {
		ep := obj.(v1.Endpoint)
		if v1.GetOwnerService(ep.OwnerReferences) == service.UID {
			UID = ep.UID
			break
		}
	}

	if !noRelated {
		epc.createEndpoint(service, relatedPods, UID)
	} else if UID != "" {
		epc.endpointInformer.DeleteItem(UID)
	}

	return nil
}

func (epc *EndpointController) createEndpoint(service v1.Service, pods []v1.Pod, prevID string) {
	var endpoint v1.Endpoint
	endpoint.Kind = "Endpoint"
	endpoint.APIVersion = service.APIVersion
	endpoint.ObjectMeta.Name = service.ObjectMeta.Name + "-endpoint"
	endpoint.UID = prevID
	endpoint.ServiceIp = service.Spec.ClusterIP

	owner := v1.OwnerReference{
		Name:       service.Name,
		Kind:       service.Kind,
		APIVersion: service.APIVersion,
		UID:        service.UID,
	}
	endpoint.OwnerReferences = make([]v1.OwnerReference, 1)
	endpoint.OwnerReferences[0] = owner
	endpoint.ServiceIp = service.Spec.ClusterIP

	subsets := make([]v1.EndpointSubset, len(pods))

	for _, pod := range pods {
		newSubset := v1.EndpointSubset{}
		addresses := make([]v1.EndpointAddress, 1)
		addresses[0].IP = pod.Status.PodIP
		newSubset.Addresses = addresses

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
		newSubset.Ports = ports

		subsets = append(subsets, newSubset)
	}

	endpoint.Subset = subsets

	if prevID == "" {
		epc.endpointInformer.AddItem(endpoint)
	} else {
		epc.endpointInformer.UpdateItem(endpoint.UID, endpoint)
	}
}

func (epc *EndpointController) portsMatch(pod *v1.Pod, ports []v1.EndpointPort) bool {
	if len(pod.Spec.ExposedPorts) == len(ports) {
		portSet := map[int]struct{}{}
		for _, port := range ports {
			portSet[int(port.Port)] = struct{}{}
		}

		for _, exposedPort := range pod.Spec.ExposedPorts {
			portAndProtocol := strings.Split(exposedPort, "/")
			port, err := strconv.Atoi(portAndProtocol[0])
			if err != nil {
				klog.Error("convert string error")
			}

			if _, exists := portSet[port]; !exists {
				return false
			}
		}

		return true
	}

	return false
}

// get Service by OwnerReferences
func (epc *EndpointController) getPodOwnerService(pod *v1.Pod) *v1.Service {
	serviceUID := v1.GetOwnerService(pod.OwnerReferences)
	if serviceUID != "" {
		item := epc.serviceInformer.GetItem(serviceUID)

		if item != nil {
			service := item.(v1.Service)
			return &service
		}
	}

	return nil
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

func (epc *EndpointController) enqueueService(service *v1.Service) {
	key := service.UID
	epc.queue.Push(key)
}

func (epc *EndpointController) addService(obj any) {
	service := obj.(v1.Service)
	klog.Info("add Service ", service.UID)
	epc.enqueueService(&service)
}

func (epc *EndpointController) updateService(newObj any, oldObj any) {
	newService := newObj.(v1.Service)
	oldService := oldObj.(v1.Service)

	needCheck := false

	if newService.Spec.ClusterIP != oldService.Spec.ClusterIP {
		needCheck = true
	} else if len(newService.Spec.Selector) != len(oldService.Spec.Selector) ||
		!v1.MatchLabels(newService.Spec.Selector, oldService.Spec.Selector) {
		needCheck = true
	} else {
		newPorts := newService.Spec.Ports
		oldPorts := oldService.Spec.Ports
		if len(newPorts) != len(oldPorts) {
			needCheck = true
		} else {
			for i := 0; i < len(newPorts); i++ {
				if !v1.CompareServicePort(newPorts[i], oldPorts[i]) {
					needCheck = true
					break
				}
			}
		}
	}

	if needCheck {
		klog.Info("update Service ", newService.UID)
		epc.enqueueService(&newService)
	}
}

func (epc *EndpointController) deleteService(obj any) {
	service := obj.(v1.Service)
	klog.Info("delete Service ", service.UID)

	eps := epc.endpointInformer.List()
	for _, item := range eps {
		ep := item.(v1.Endpoint)
		if v1.GetOwnerService(ep.OwnerReferences) == service.UID {
			epc.endpointInformer.DeleteItem(ep.UID)
			break
		}
	}

	pods := epc.podInformer.List()
	for _, item := range pods {
		pod := item.(v1.Pod)
		index := v1.CheckOwner(pod.OwnerReferences, service.UID)
		if index >= 0 {
			pod.OwnerReferences = append(pod.OwnerReferences[:index], pod.OwnerReferences[index+1:]...)
			epc.podInformer.UpdateItem(pod.UID, pod)
		}
	}
}

func (epc *EndpointController) addPod(obj any) {
	pod := obj.(v1.Pod)
	if pod.Status.PodIP == "" {
		return
	}

	services := epc.getPodMatchService(&pod)
	for _, service := range services {
		klog.Infof("enqueue Service %s when add Pod %s", service.UID, pod.UID)
		epc.enqueueService(&service)
	}
}

func (epc *EndpointController) updatePod(newObj any, oldObj any) {
	newPod := newObj.(v1.Pod)
	oldPod := oldObj.(v1.Pod)

	if !v1.MatchLabels(newPod.Labels, oldPod.Labels) || (newPod.Status.PodIP != "" && newPod.Status.PodIP != oldPod.Status.PodIP) {
		service := epc.getPodOwnerService(&oldPod)
		if service != nil {
			klog.Infof("enqueue Service %s when update Pod %s", service.UID, newPod.UID)
			epc.enqueueService(service)
			return
		}

		services := epc.getPodMatchService(&newPod)
		for _, s := range services {
			klog.Infof("enqueue Service %s when update Pod %s", s.UID, newPod.UID)
			epc.enqueueService(&s)
		}
	}
}

func (epc *EndpointController) deletePod(obj any) {
	pod := obj.(v1.Pod)

	service := epc.getPodOwnerService(&pod)
	if service != nil {
		klog.Infof("enqueue Service %s when delete Pod %s", service.UID, pod.UID)
		epc.enqueueService(service)
	} else {
		klog.Infof("delete Pod %s, which has no owner Service", pod.UID)
	}
}
