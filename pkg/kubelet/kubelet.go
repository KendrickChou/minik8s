package kubelet

import (
	"errors"
	"fmt"
	"os"
	"os/exec"

	"github.com/gin-gonic/gin"
	"k8s.io/klog/v2"

	v1 "minik8s.com/minik8s/pkg/api/v1"
	"minik8s.com/minik8s/pkg/kubelet/apis/config"
	kubeconfig "minik8s.com/minik8s/pkg/kubelet/apis/config"
	"minik8s.com/minik8s/pkg/kubelet/apis/constants"
	"minik8s.com/minik8s/pkg/kubelet/pod"
	"minik8s.com/minik8s/pkg/kubelet/server"
)

type Kubelet struct {
	v1.Node

	podManager pod.PodManager
	// same as podManager, just for test
	PodManager *pod.PodManager

	setNodeStatusFuncs []func(*v1.Node)
}

func NewKubelet(nodeName string, UID string) (Kubelet, error) {
	// err := kubelet.podManager.CreatePodBridgeNetwork(kubelet.Spec.CIDR)
	err := connectWeaveNet()

	if err != nil {
		klog.Errorf("Connect to weave net error: %s", err.Error())
	}

	kubelet := Kubelet{
		Node: v1.Node{
			TypeMeta: v1.TypeMeta{
				Kind:       "Node",
				APIVersion: "v1",
			},
			ObjectMeta: v1.ObjectMeta{
				Name:      nodeName,
				Namespace: "default",
				UID:       UID,
			},
		},
		podManager: pod.NewPodManager(),
	}

	kubelet.PodManager = &kubelet.podManager

	return kubelet, err
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
	installInitialContainers(&pod)

	err := kl.podManager.AddPod(&pod)

	if err != nil {
		return v1.Pod{}, err
	}

	return pod, err
}

func (kl *Kubelet) DeletePod(UID string) error {
	return kl.podManager.DeletePod(UID)
}

func installInitialContainers(pod *v1.Pod) error {
	pod.Spec.InitialContainers = make(map[string]v1.Container)

	pod.Spec.InitialContainers[constants.InitialPauseContainerKey] = constants.InitialPauseContainer

	return nil
}

func connectWeaveNet() error {
	// If there is a firewall between $HOST1 and $HOST2,
	// you must permit traffic to flow through TCP 6783 and UDP 6783/6784,
	// which are Weave’s control and data ports.

	// connect to weave net
	cmd := exec.Command("weave", "connect", config.WeaveServerIP)
	out, err := cmd.CombinedOutput()

	if err != nil {
		errInfo := fmt.Sprintf("Error in Weave Connect: %s", err.Error())
		return errors.New(errInfo)
	}

	klog.Infof("Weave Connect to %s:%s", config.WeaveServerIP, out)

	cmd = exec.Command("bash", "-c","eval $(weave env)")
	_, err = cmd.CombinedOutput()

	if err != nil {
		errInfo := fmt.Sprintf("Error in set Weave env: %s", err.Error())
		return errors.New(errInfo)
	}

	cmd = exec.Command("weave", "expose")
	_, err = cmd.CombinedOutput()

	if err != nil {
		errInfo := fmt.Sprintf("Error in set Weave env: %s", err.Error())
		return errors.New(errInfo)
	}

	klog.Info("Set Weave Env Successfully!")
	return nil
}
