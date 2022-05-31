package main

import (
	"minik8s.com/minik8s/pkg/controller/component"
	"minik8s.com/minik8s/pkg/controller/endpoint"
	"minik8s.com/minik8s/pkg/controller/job"
	"minik8s.com/minik8s/pkg/controller/podautoscaling"
	rs "minik8s.com/minik8s/pkg/controller/replicaset"
	"minik8s.com/minik8s/utils/random"
	"time"
)

func main() {
	random.Init()

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

	hpInformer := component.NewInformer("HorizontalPodAutoscaler")
	hpStopChan := make(chan bool)
	go hpInformer.Run(hpStopChan)

	jobInformer := component.NewInformer("GPUJob")
	jobStopChan := make(chan bool)
	go jobInformer.Run(jobStopChan)

	for !(podInformer.HasSynced() && rsInformer.HasSynced() && serviceInformer.HasSynced() &&
		endpointInformer.HasSynced() && hpInformer.HasSynced() && jobInformer.HasSynced()) {

	}

	rsController := rs.NewReplicaSetController(podInformer, rsInformer)
	go rsController.Run()

	endpointController := endpoint.NewEndpointController(podInformer, serviceInformer, endpointInformer)
	go endpointController.Run()

	hpaController := podautoscaling.NewHorizontalController(hpInformer, podInformer, rsInformer)
	go hpaController.Run()

	jobController := job.NewJobController(podInformer, jobInformer)
	go jobController.Run()

	for {
		time.Sleep(time.Second * 100)
	}
}
