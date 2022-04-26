package kubelet

import (
	"errors"
	"os"

	"github.com/gin-gonic/gin"
	"k8s.io/klog/v2"

	"minik8s.com/minik8s/pkg/api/v1"
	kubeconfig "minik8s.com/minik8s/pkg/kubelet/apis/config"
	"minik8s.com/minik8s/pkg/kubelet/pod"
	"minik8s.com/minik8s/pkg/kubelet/server"
)

type Kubelet struct {
	nodeName string

	podManager pod.PodManager	
	// same as podManager, just for test
	PodManager *pod.PodManager
}

func NewKubelet(apiServerAddr string) Kubelet {
	kubelet := Kubelet{
		nodeName: "kubelet",
		podManager: pod.NewPodManager(),
	}

	kubelet.PodManager = &kubelet.podManager

	return kubelet
}

func (kl *Kubelet) ListenAndServe(kubeCfg *kubeconfig.KubeletConfiguration) {
	address := kubeCfg.Address
	port := kubeCfg.Port

	klog.Infof("Run Kubelet Server in %s:%s", address, port)

	s := gin.Default()

	server.InstallDefaultHandlers(s, kl)

	if err := s.Run(address + ":" + port); err != nil {
		klog.Errorln(err, "Failed to listen and serve")
		os.Exit(1)
	}
}

func (kl *Kubelet) GetPods() ([]v1.Pod, error) {
	return kl.podManager.GetPods(), nil
}

func (kl *Kubelet) GetPodByUID(UID string) (v1.Pod, error) {

	pod, ok := kl.podManager.GetPodByUID(UID)

	if !ok {
		return v1.Pod{}, errors.New("pod " + string(UID) + " is not found")
	}

	return pod, nil
}

func (kl *Kubelet) CreatePod(pod v1.Pod) (v1.Pod, error) {
	err := kl.podManager.AddPod(&pod)

	if err != nil {
		return v1.Pod{}, err
	}
	
	return pod, err
}

func (kl *Kubelet) DeletePod(UID string) error {
	return kl.podManager.DeletePod(UID)
}