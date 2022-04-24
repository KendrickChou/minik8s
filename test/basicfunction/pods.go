package main

import (
	"k8s.io/klog/v2"
	v1 "minik8s.com/minik8s/pkg/api/v1"
	"minik8s.com/minik8s/pkg/kubelet/pod"
)

func main() {
	podManager := pod.NewPodManager()

	klog.Infoln("create pod manager ", podManager)

	pod := &v1.Pod{
		TypeMeta: v1.TypeMeta{
			Kind:       "pod",
			APIVersion: "v1",
		},
		ObjectMeta: v1.ObjectMeta{
			Name:      "myFirstPod",
			Namespace: "default",
			UID:       "123456789",
		},
		Spec: v1.PodSpec{
			Containers: []*v1.Container{
				{
					Name:      "myFirstContainer",
					Namespace: "example",
					ID:        "123456789",
					Image:     "docker.io/library/hello-world:latest",
				},
			},
		},
		Status: v1.PodStauts{},
	}

	podManager.AddPod(pod)
}
