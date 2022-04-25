package main

import (
	v1 "minik8s.com/minik8s/pkg/api/v1"
	"minik8s.com/minik8s/pkg/kubelet"
	kubeconfig "minik8s.com/minik8s/pkg/kubelet/apis/config"
)

func createAPod(kl *kubelet.Kubelet) {
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
					Name:            "myFirstContainer",
					Namespace:       "example",
					Image:           "alpine:latest",
					ImagePullPolicy: "IfNotPresent",
					Entrypoint:      []string{"/bin/sh", "-c", "cat /home/mountdir/hello.txt"},
					Mounts: []v1.Mount{
						{
							Type:   v1.TypeBind,
							Source: "/home/kendrick/mountdir",
							Target: "/home/mountdir",
						},
					},
				},
				{
					Name:            "mySecondContainer",
					Namespace:       "example",
					Image:           "alpine:latest",
					ImagePullPolicy: "IfNotPresent",
					Entrypoint:      []string{"echo", "hello world 2"},
				},
			},
		},
		Status: v1.PodStatus{},
	}

	(*kl.PodManager).AddPod(pod)
}

func main() {
	kubelet :=kubelet.NewKubelet()

	// createAPod(&kubelet)

	kubelet.ListenAndServe(&kubeconfig.KubeletConfiguration{
		TypeMeta: v1.TypeMeta{Kind: "kubelet", APIVersion: "v1"},
		Address:  "localhost",
		Port:     "8080"})

	return
}
