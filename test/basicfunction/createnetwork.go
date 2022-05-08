package main

import (
	"k8s.io/klog"
	"minik8s.com/minik8s/pkg/kubelet"
)

func main() {
	_, err := kubelet.NewKubelet("simpleKubelet", "1234", "10.2.1.0/24", "10.0.0.0/24")

	if err != nil {
		klog.Error(err)
		return
	}

}
