package main

import (
	"k8s.io/klog/v2"
	v1 "minik8s.com/minik8s/pkg/api/v1"
	"minik8s.com/minik8s/pkg/kubelet/apis/constants"
	"minik8s.com/minik8s/pkg/kubelet/pod"
)

func main() {
	podManager := pod.NewPodManager()

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
			InitialContainers: map[string]v1.Container{
				constants.InitialPauseContainerKey: constants.InitialPauseContainer,
			},
			Containers: []*v1.Container{
				{
					Name:            "myFirstContainer",
					Namespace:       "example",
					Image:           "alpine:latest",
					ImagePullPolicy: "IfNotPresent",
					Entrypoint:      []string{"/bin/sh", "-c", "wget -o /home/mountdir/nginx.html 10.40.0.0:8080"},
					Mounts: []v1.Mount{
						{
							Type:   v1.TypeBind,
							Source: "/home/kendrick/mountdir",
							Target: "/home/mountdir",
						},
					},
				},
				// {
				// 	Name:            "mySecondContainer",
				// 	Namespace:       "example",
				// 	Image:           "alpine:latest",
				// 	ImagePullPolicy: "IfNotPresent",
				// 	Entrypoint:      []string{"echo", "hello world 2"},
				// },
			},
		},
		Status: v1.PodStatus{},
	}

	podManager.AddPod(pod)

	status, _ := podManager.PodStatus(pod.UID)

	klog.Info(status)

	podManager.DeletePod(pod.UID)
}
