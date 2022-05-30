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

var ApiServerIP string = "10.119.11.209"

var WeaveServerIP string = "10.119.11.209"

var ApiServerAddress string = "http://" + ApiServerIP + ":8080"

var GatewayAddress string = "addr"

var NodeName string = "node"
