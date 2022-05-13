package endpoint

import (
	"k8s.io/klog"
	v1 "minik8s.com/minik8s/pkg/api/v1"
	"minik8s.com/minik8s/pkg/controller/component"
)

type EndpointController struct {
	podInformer      *component.Informer
	endpointInformer *component.Informer
	serviceInformer  *component.Informer
	queue            component.WorkQueue
}

func NewReplicaSetController(podInfo *component.Informer, servInfo *component.Informer,
	epInfo *component.Informer) *EndpointController {
	return &EndpointController{
		podInformer:      podInfo,
		endpointInformer: epInfo,
		serviceInformer:  servInfo,
	}
}

func (epc *EndpointController) Run() {
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

	epc.worker()
}

func (epc *EndpointController) worker() {
	for epc.processNextWorkItem() {
	}
}

func (epc *EndpointController) processNextWorkItem() bool {
	key := epc.queue.Get().(string)
	defer epc.queue.Done(key)

	service := epc.serviceInformer.GetItem(key).(v1.Service)
	err := epc.syncEndpoint(service)
	if err != nil {
		klog.Error("syncEndpoint error\n")
		return false
	}
	return true
}

func (epc *EndpointController) syncEndpoint(service v1.Service) error {
	// check if there existing endpoint
	eps := epc.endpointInformer.List()
	for _, obj := range eps {
		ep := obj.(v1.Endpoint)
		if ep.Name == service.Name {
			// get related pods
			pods := epc.podInformer.List()
			for _, podObj := range pods {
				pod := podObj.(v1.Pod)

			}

			return nil
		}
	}

	// create new endpoint
	epc.createEndpoint(service)
	return nil
}

func (epc *EndpointController) createEndpoint(service v1.Service) {

}

// get Service by OwnerReferences
func (epc *EndpointController) getPodService(pod *v1.Pod) []v1.Service {
	var result []v1.Service
	for _, owner := range pod.OwnerReferences {
		service := epc.serviceInformer.GetItem(owner.UID)
		if service != nil {
			result = append(result, service.(v1.Service))
		}
	}
	return result
}

func (epc *EndpointController) enqueueService(service v1.Service) {
	key := service.UID
	epc.queue.Add(key)
}

func (epc *EndpointController) addService(obj any) {
	service := obj.(v1.Service)
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
	// do nothing
}

func (epc *EndpointController) updatePod(newObj any, oldObj any) {
	newPod := newObj.(v1.Pod)
	oldPod := oldObj.(v1.Pod)

	if newPod.Status.PodIP != oldPod.Status.PodIP {
		services := epc.getPodService(&oldPod)
		for _, service := range services {
			epc.enqueueService(service)
		}
	}
}

func (epc *EndpointController) deletePod(obj any) {
	pod := obj.(v1.Pod)
	services := epc.getPodService(&pod)
	for _, service := range services {
		epc.enqueueService(service)
	}
}
