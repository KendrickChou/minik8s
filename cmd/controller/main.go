package main

import (
	"minik8s.com/minik8s/pkg/controller/component"
	rs "minik8s.com/minik8s/pkg/controller/replicaset"
)

func main() {
	podInformer := component.NewInformer("Pod")
	podStopChan := make(chan bool)
	go podInformer.Run(podStopChan)

	rsInformer := component.NewInformer("ReplicaSet")
	rsStopChan := make(chan bool)
	go rsInformer.Run(rsStopChan)

	rsController := rs.NewReplicaSetController(podInformer, rsInformer)
	go rsController.Run()

	for {

	}
}
