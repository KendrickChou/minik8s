package main

import (
	"minik8s.com/minik8s/pkg/controller/component"
	"minik8s.com/minik8s/pkg/controller/endpoint"
)

func main() {
	podInformer := component.NewInformer("Pod")
	podStopChan := make(chan bool)
	go podInformer.Run(podStopChan)

	rsInformer := component.NewInformer("ReplicaSet")
	rsStopChan := make(chan bool)
	go rsInformer.Run(rsStopChan)

	serviceInformer := component.NewInformer("Service")
	serviceStopChan := make(chan bool)
	go serviceInformer.Run(serviceStopChan)

	endpointInformer := component.NewInformer("Endpoint")
	endpointStopChan := make(chan bool)
	go endpointInformer.Run(endpointStopChan)

	//rsController := rs.NewReplicaSetController(podInformer, rsInformer)
	//go rsController.Run()
	endpointController := endpoint.NewEndpointController(podInformer, serviceInformer, endpointInformer)
	go endpointController.Run()

	for {
		
	}
}
