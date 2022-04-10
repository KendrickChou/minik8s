package kubelet

import (
	clientset "minik8s.com/minik8s/pkg/client"
	"minik8s.com/minik8s/pkg/kubelet/container"
	pod "minik8s.com/minik8s/pkg/kubelet/pod"
	kubetypes "minik8s.com/minik8s/pkg/kubelet/types"
)

type kuberlet struct {
	hostname string

	nodeName kubetypes.NodeName

	kubeClient      clientset.Interface
	heartbeatClient clientset.Interface

	podManager       pod.Manager
	containerManager container.Manager
}
