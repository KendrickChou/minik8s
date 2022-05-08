package config

import (
	v1 "minik8s.com/minik8s/pkg/api/v1"
)

type KubeletConfiguration struct {
	v1.TypeMeta

	Address string
	Port    string

	HeartbeatInterval uint64
}

var	InternalPodBridgeNetworkName string = "kubelet"

var ApiServerAddress string = "addr"

var GatewayAddress string = "addr"

var NodeName string = "node"
