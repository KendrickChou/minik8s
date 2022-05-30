package main

import (
	"k8s.io/klog/v2"
	v1 "minik8s.com/minik8s/pkg/api/v1"
	"minik8s.com/minik8s/pkg/kubelet/apis/constants"
	"minik8s.com/minik8s/pkg/kubelet/pod"
)

func main() {
	podManager := pod.NewPodManager()

	portPod := &v1.Pod{
		TypeMeta: v1.TypeMeta{
			Kind:       "pod",
			APIVersion: "v1",
		},
		ObjectMeta: v1.ObjectMeta{
			Name:      "cntr-features",
			Namespace: "default",
			UID:       "123456789",
		},
		Spec: v1.PodSpec{
			InitialContainers: map[string]v1.Container{
				constants.InitialPauseContainerKey: constants.InitialPauseContainer,
			},
			Containers: []*v1.Container{
				{
					Name:            "myFirstContainer",
					Namespace:       "example",
					Image:           "nginx:latest",
					ImagePullPolicy: "IfNotPresent",
				},
			},
			ExposedPorts: []string{
				"80/tcp",
			},
			BindPorts: map[string]string{
				"80/tcp": "127.0.0.1:9095",
			},
		},
		Status: v1.PodStatus{},
	}

	podManager.AddPod(portPod)

	status, _ := podManager.PodStatus(portPod.UID)

	klog.Info(status)

	// podManager.DeletePod(portPod.UID)
}
