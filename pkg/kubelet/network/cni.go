package network

import (
	"context"
	"encoding/json"

	cni "github.com/containernetworking/cni/libcni"
	cnitypes "github.com/containernetworking/cni/pkg/types"
	"k8s.io/klog"
	v1 "minik8s.com/minik8s/pkg/api/v1"
	"minik8s.com/minik8s/pkg/kubelet/apis/config"
	"minik8s.com/minik8s/pkg/kubelet/apis/constants"
)

func SetUpPodNetwork(pod *v1.Pod) error {
	cniConfig := cni.NewCNIConfig([]string{config.CNIPluginPath}, nil)

	netConfig := &cnitypes.NetConf{
		CNIVersion: config.CNIVersion,
		Name: pod.Name + "-" + pod.UID + "-" + "network",
		Type: config.NetWorkType,
		IPAM: cnitypes.IPAM{Type: config.IPAMType},
	}
	
	bytes, _ := json.Marshal(netConfig)
	networkConfig, err := cni.ConfFromBytes(bytes)

	if err != nil {
		klog.Error(err.Error())
		return err
	}

	runtimeConfig := &cni.RuntimeConf{
		ContainerID: pod.Spec.InitialContainers[constants.InitialPauseContainer.Name].ID,
		NetNS: "/var/run/docker/netns/...",
		IfName: "eth0",
	}
	cniConfig.AddNetwork(context.TODO(), networkConfig, runtimeConfig)
}