package main

import (
	v1 "minik8s.com/minik8s/pkg/api/v1"
	"minik8s.com/minik8s/pkg/kubelet"
)

func main() {
	kubelet := kubelet.NewKubelet("name")

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
					Name:            "nginx",
					Namespace:       "example",
					Image:           "nginx:latest",
					ImagePullPolicy: "IfNotPresent",
				},
				{
					Name:            "alpine",
					Namespace:       "example",
					Image:           "alpine:latest",
					ImagePullPolicy: "IfNotPresent",
					Entrypoint:      []string{"/bin/sh", "-c", "wget -O /home/mountdir/nginx.html localhost:80"},
					Mounts: []v1.Mount{
						{
							Type:   v1.TypeBind,
							Source: "/home/kendrick/mountdir",
							Target: "/home/mountdir",
						},
					},
				},
			},
		},
		Status: v1.PodStatus{},
	}

	kubelet.CreatePod(*pod)

	// kubelet.DeletePod(pod.UID)
}
