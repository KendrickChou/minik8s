package main

import (
	"context"

	"k8s.io/klog"
	v1 "minik8s.com/minik8s/pkg/api/v1"
	kubeproxy "minik8s.com/minik8s/pkg/kubelet/kube-proxy"
)

func main() {
	kp, err := kubeproxy.NewKubeProxy()

	if err != nil {
		klog.Fatalf(err.Error())
		return
	}

	ep := v1.Endpoint{
		TypeMeta: v1.TypeMeta{
			Kind:       "endpoint",
			APIVersion: "v1",
		},
		ObjectMeta: v1.ObjectMeta{
			Name: "test-endpoint",
		},
		ServiceIp: "10.8.8.8",
		Subset: []v1.EndpointSubset{
			{
				Addresses: []v1.EndpointAddress{
					{IP: "10.40.0.0"},
					{IP: "10.40.0.3"},
				},
				Ports: []v1.EndpointPort{
					{Name: "port1", Port: 80, ServicePort: 81, Protocol: "tcp"},
					{Name: "port2", Port: 80, ServicePort: 82, Protocol: "tcp"},
				},
			},
		},
	}

	kp.AddEndpoint(context.TODO(), "TEST", ep)
	kp.RemoveEndpoint(context.TODO(), "TEST")
}
