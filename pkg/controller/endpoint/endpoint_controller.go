package endpoint

import "minik8s.com/minik8s/pkg/controller/component"

type EndpointController struct {
	podInformer      *component.Informer
	endpointInformer *component.Informer
	serviceInformer  *component.Informer
}
