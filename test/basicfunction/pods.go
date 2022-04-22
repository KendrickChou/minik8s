package pods_test

import (
	"k8s.io/klog/v2"
	v1 "minik8s.com/minik8s/pkg/api/v1"
	"minik8s.com/minik8s/pkg/kubelet/pod"
)

func main() {
	podManager := pod.NewPodManager()
	
	pod := &v1.Pod{
		v1.TypeMeta{
			Kind: "pod",
			APIVersion: "v1",
		},
		v1.ObjectMeta{
			Name: "myFirstPod",
			Namespace: "default",
			UID: "123456789",
		},
		v1.PodSpec{
			Containers: []*v1.Container{
				&v1.Container{
					Name: "myFirstContainer",
					ID: "123456789",
					Image: "docker.io/library/hello-world",
				},
			},
		},
		v1.PodStauts{

		},
	}

	podManager.AddPod(pod)
}