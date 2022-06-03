package main

import (
	"k8s.io/klog/v2"
	v1 "minik8s.com/minik8s/pkg/api/v1"
	"minik8s.com/minik8s/pkg/kubelet/apis/constants"
	"minik8s.com/minik8s/pkg/kubelet/pod"
)

func main() {
	podManager := pod.NewPodManager()

	// portPod := &v1.Pod{
	// 	TypeMeta: v1.TypeMeta{
	// 		Kind:       "pod",
	// 		APIVersion: "v1",
	// 	},
	// 	ObjectMeta: v1.ObjectMeta{
	// 		Name:      "cntr-features",
	// 		Namespace: "default",
	// 		UID:       "123456789",
	// 	},
	// 	Spec: v1.PodSpec{
	// 		InitialContainers: map[string]v1.Container{
	// 			constants.InitialPauseContainerKey: constants.InitialPauseContainer,
	// 		},
	// 		Containers: []*v1.Container{
	// 			{
	// 				Name:            "myFirstContainer",
	// 				Namespace:       "example",
	// 				Image:           "nginx:latest",
	// 				ImagePullPolicy: "IfNotPresent",
	// 			},
	// 		},
	// 		ExposedPorts: []string{
	// 			"80/tcp",
	// 		},
	// 		BindPorts: map[string]string{
	// 			"80/tcp": "0.0.0.0:9095",
	// 		},
	// 	},
	// 	Status: v1.PodStatus{},
	// }

	// resourcePod := &v1.Pod{
	// 	TypeMeta: v1.TypeMeta{
	// 		Kind:       "pod",
	// 		APIVersion: "v1",
	// 	},
	// 	ObjectMeta: v1.ObjectMeta{
	// 		Name:      "cntr-features",
	// 		Namespace: "default",
	// 		UID:       "123456789",
	// 	},
	// 	Spec: v1.PodSpec{
	// 		InitialContainers: map[string]v1.Container{
	// 			constants.InitialPauseContainerKey: constants.InitialPauseContainer,
	// 		},
	// 		Containers: []*v1.Container{
	// 			{
	// 				Name:            "myFirstContainer",
	// 				Namespace:       "example",
	// 				Image:           "nginx:latest",
	// 				ImagePullPolicy: "IfNotPresent",
	// 				Resources:       map[string]string{
	// 					"cpu": "2",
	// 					"memory": "128MB",
	// 				},
	// 			},
	// 		},
	// 	},
	// 	Status: v1.PodStatus{},
	// }

	volumePod := &v1.Pod{
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
					Image:           "alpine:latest",
					ImagePullPolicy: "IfNotPresent",
					Entrypoint:      []string{"/bin/sh", "-c", "wget -O /home/mountdir/nginx.html http://i.sjtu.edu.cn"},
					Mounts: []v1.Mount{
						{
							Type: "volume",
							Source: "volume1",
							Target: "/home/mountdir",
						},
					},
				},
				{
					Name:            "mySecondContainer",
					Namespace:       "example",
					Image:           "alpine:latest",
					ImagePullPolicy: "IfNotPresent",
					Entrypoint:      []string{"/bin/sh", "-c", "cat /root/mountdir/nginx.html"},
					Mounts: []v1.Mount{
						{
							Type: "volume",
							Source: "volume1",
							Target: "/root/mountdir",
						},
					},
				},
			},
			Volumes: []string{
				"volume1",
				"volume2",
			},
		},
		Status: v1.PodStatus{},
	}


	podManager.AddPod(volumePod)

	status, _ := podManager.PodStatus(volumePod.UID)

	klog.Info(status)

	podManager.DeletePod(volumePod.UID)
}
