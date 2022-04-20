package kubelet

import (
	"net"
	"net/http"
	"os"

	"k8s.io/klog/v2"
	"minik8s.com/minik8s/pkg/kubelet/pod"

	kubeconfig "minik8s.com/minik8s/pkg/kubelet/apis/config"

	"minik8s.com/minik8s/pkg/kubelet/server"
)

type Kubelet struct {
	nodeName string

	podManager pod.PodManager
}

func (kl *Kubelet) ListenAndServe(kubeCfg *kubeconfig.KubeletConfiguration) {
	address  := kubeCfg.Address
	port := kubeCfg.Port

	handler := server.NewServer()

	s := &http.Server{
		Addr: net.JoinHostPort(address, port),
		Handler: &handler,
	}

	if err := s.ListenAndServe(); err != nil {
		klog.Errorln(err, "Failed to listen and serve")
		os.Exit(1)
	}
}