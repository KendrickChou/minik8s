package pod

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/filters"
	"k8s.io/klog/v2"
	v1 "minik8s.com/minik8s/pkg/api/v1"
	"minik8s.com/minik8s/pkg/kubelet/apis/config"
	"minik8s.com/minik8s/pkg/kubelet/apis/constants"
	"minik8s.com/minik8s/pkg/kubelet/container"
)

type PodManager interface {
	GetPods() []v1.Pod

	GetPodByName(podFullName string) (v1.Pod, bool)

	GetPodByUID(UID string) (v1.Pod, bool)

	AddPod(pod *v1.Pod) error

	AddPodWithoutCreate(pod *v1.Pod) error

	UpdatePod(pod *v1.Pod) error

	DeletePod(UID string) error

	PodStatus(UID string) (v1.PodStatus, error)

	CreatePodBridgeNetwork(CIDR string) error

	CheckDuplicate(pod *v1.Pod) bool
}

type podManager struct {
	// Regular pods indexed by UID.
	podByUID map[string]*v1.Pod

	podByName map[string]*v1.Pod

	containerManager container.ContainerManager

	weaveNetwork types.NetworkResource
}

func NewPodManager() PodManager {
	pm := &podManager{}
	pm.podByUID = make(map[string]*v1.Pod)
	pm.podByName = make(map[string]*v1.Pod)

	newContainerManager, err := container.NewContainerManager()

	if err != nil {
		klog.Errorln(err)
	}

	pm.containerManager = newContainerManager

	network, err := pm.containerManager.ListNetwork(context.TODO(),
		types.NetworkListOptions{
			Filters: filters.NewArgs(filters.Arg("name", "weave")),
		})

	if err != nil {
		klog.Errorf("Get Weave Network Error: %s", err.Error())
		os.Exit(0)
	}

	if len(network) == 0 {
		klog.Errorln("Weave Network doesn't exist.")
		os.Exit(0)
	}

	pm.weaveNetwork = network[0]

	klog.Infoln("Weave Network Info: ", pm.weaveNetwork.Name, pm.weaveNetwork.ID, pm.weaveNetwork.IPAM.Config)

	return pm
}

func (pm *podManager) GetPods() []v1.Pod {
	pods := []v1.Pod{}

	for _, value := range pm.podByUID {
		// refresh the pod status
		pm.PodStatus(value.UID)
		pods = append(pods, *value)
	}

	return pods
}

func (pm *podManager) GetPodByName(podFullName string) (v1.Pod, bool) {
	pod, ok := pm.podByName[podFullName]

	if ok {
		pm.PodStatus(pod.UID)
		return *pod, ok
	}

	return v1.Pod{}, ok
}

func (pm *podManager) GetPodByUID(UID string) (v1.Pod, bool) {
	pod, ok := pm.podByUID[UID]

	if ok {
		pm.PodStatus(pod.UID)
		return *pod, ok
	}

	return v1.Pod{}, ok
}

// there is only one namespace named "default"
func (pm *podManager) AddPod(pod *v1.Pod) error {
	klog.Infof("Add Pod %s", pod.Name)

	if pod.UID == "" {
		err := "pod UID is empty"

		klog.Errorln(err)
		return errors.New(err)
	}

	pm.podByUID[pod.UID] = pod
	pm.podByName[pod.Name] = pod

	// start some pod initial containers
	for k := range pod.Spec.InitialContainers {
		container := pod.Spec.InitialContainers[k]

		container.Name = pod.Name + "-" + container.Name

		container.ExposedPorts = pod.Spec.ExposedPorts

		container.BindPorts = pod.Spec.BindPorts

		id, err := pm.containerManager.CreateContainer(context.TODO(), &container)

		if err != nil {
			klog.Errorf("Create pod %s Initial container %s failed: %s", pod.Name, container.Name, err.Error())
			return err
		}

		container.ID = id

		err = pm.containerManager.StartContainer(context.TODO(), &container)

		if err != nil {
			klog.Errorf("Start pod %s Initial container %s failed: %s", pod.Name, container.Name, err.Error())
			klog.Errorln(err)
			return err
		}

		pod.Spec.InitialContainers[k] = container

		timeoutctx, cancel := context.WithTimeout(context.TODO(), time.Second*5)
		err = pm.containerManager.ConnectNetwork(timeoutctx, pm.weaveNetwork.ID, container.ID)
		cancel()

		if err != nil {
			klog.Errorf("Pod %s Connect to Weave Network Failed", pod.Name)
			return err
		}
	}

	// create related volumes
	for _, volume := range pod.Spec.Volumes {
		klog.Infof("Create Volume %s For Pod %s", pod.Name, volume)
		cmd := exec.Command("docker", "volume", "create", pod.Name+"-"+volume)
		cmd.CombinedOutput()
	}

	// start user spec pods
	for _, container := range pod.Spec.Containers {
		container.Name = pod.Name + "-" + container.Name
		container.NetworkMode = constants.NetworkIDPrefix + pod.Spec.InitialContainers[constants.InitialPauseContainerKey].Name

		id, err := pm.containerManager.CreateContainer(context.TODO(), container)

		if err != nil {
			klog.Errorln(err)
			continue
		}

		container.ID = id

		err = pm.containerManager.StartContainer(context.TODO(), container)

		if err != nil {
			klog.Errorln(err)
		}
	}

	// refresh modified pod spec to apiserver.(for restart)
	body, _ := json.Marshal(pod)
	req, _ := http.NewRequest(http.MethodPut, config.ApiServerAddress+constants.RefreshPodRequest(constants.Node.UID, pod.UID), bytes.NewReader(body))
	resp, err := http.DefaultClient.Do(req)

	if err != nil {
		klog.Errorf("Error When refresh Pod Status: %s", err.Error())
		return err
	}

	resp.Body.Close()

	return nil
}

func (pm *podManager) AddPodWithoutCreate(pod *v1.Pod) error {
	klog.Infof("Add Pod %s without create", pod.Name)

	if pod.UID == "" {
		err := "pod UID is empty"

		klog.Errorln(err)
		return errors.New(err)
	}

	pm.podByUID[pod.UID] = pod
	pm.podByName[pod.Name] = pod

	return nil
}

func (pm *podManager) UpdatePod(pod *v1.Pod) error {
	if pod.UID == "" {
		err := "pod UID is empty"
		klog.Errorln(err)
		return errors.New(err)
	}

	oldPod, ok := pm.podByUID[pod.UID]

	if !ok {
		err := "Pod doesn't exist, UID: " + pod.UID
		klog.Errorln(err)
		return errors.New(err)
	}

	delete(pm.podByName, oldPod.Name)
	pm.podByName[pod.Name] = pod
	pm.podByUID[pod.UID] = pod

	errs := []string{}

	for _, container := range oldPod.Spec.Containers {
		err := pm.containerManager.RemoveContainer(context.TODO(), container)

		if err != nil {
			klog.Errorln(err)
			errs = append(errs, err.Error())
		}
	}

	for _, container := range pod.Spec.Containers {
		id, err := pm.containerManager.CreateContainer(context.TODO(), container)

		if err != nil {
			klog.Errorln(err)
			errs = append(errs, err.Error())
			continue
		}

		container.ID = id

		err = pm.containerManager.StartContainer(context.TODO(), container)

		if err != nil {
			klog.Errorln(err)
			errs = append(errs, err.Error())
		}
	}

	if len(errs) == 0 {
		return nil
	}

	var allErrs string
	for _, e := range errs {
		allErrs = fmt.Sprintf("%s\n%s", allErrs, e)
	}

	return errors.New(allErrs)
}

func (pm *podManager) DeletePod(UID string) error {
	if UID == "" {
		err := "pod UID is empty"
		klog.Errorln(err)
		return errors.New(err)
	}

	pod := pm.podByUID[UID]

	klog.Infof("Delete Pod %s", pod.Name)

	var errs []string

	for k := range pod.Spec.InitialContainers {
		container := pod.Spec.InitialContainers[k]

		err := pm.containerManager.StopContainer(context.TODO(), &container)

		if err != nil {
			klog.Errorf("Remove Container %s: %s", container.Name, err)
			errs = append(errs, err.Error())
		}

		err = pm.containerManager.RemoveContainer(context.TODO(), &container)

		if err != nil {
			klog.Errorf("Remove Container %s: %s", container.Name, err)
			errs = append(errs, err.Error())
		}
	}

	for _, volume := range pod.Spec.Volumes {
		klog.Infof("Delete Volume %s For Pod %s", pod.Name, volume)
		cmd := exec.Command("docker", "volume", "rm", pod.Name+"-"+volume)
		cmd.CombinedOutput()
	}

	for _, container := range pod.Spec.Containers {
		err := pm.containerManager.StopContainer(context.TODO(), container)

		if err != nil {
			klog.Errorf("Remove Container %s: %s", container.Name, err)
			errs = append(errs, err.Error())
		}

		err = pm.containerManager.RemoveContainer(context.TODO(), container)

		if err != nil {
			klog.Errorf("Remove Container %s: %s", container.Name, err)
			errs = append(errs, err.Error())
		}
	}

	delete(pm.podByName, pod.Name)
	delete(pm.podByUID, pod.UID)

	if len(errs) == 0 {
		return nil
	}

	var allErrs string
	for _, e := range errs {
		allErrs = fmt.Sprintf("%s\n%s", allErrs, e)
	}

	return errors.New(allErrs)
}

func (pm *podManager) PodStatus(UID string) (v1.PodStatus, error) {
	if UID == "" {
		err := "pod UID is empty"
		klog.Errorln(err)
		return v1.PodStatus{}, errors.New(err)
	}

	pod := pm.podByUID[UID]

	klog.Infof("Inspect Pod %s, UID: %s status", pod.Name, pod.UID)

	pod.Status.ContainerStatuses = []v1.ContainerStatus{}

	// get pause container statuses
	for k := range pod.Spec.InitialContainers {
		cntr := pod.Spec.InitialContainers[k]
		stats, err := pm.containerManager.ContainerStatus(context.TODO(), cntr.ID)

		dynamicStats, err1 := pm.containerManager.ContainerStats(context.TODO(), cntr.ID)

		if err != nil || err1 != nil || len(dynamicStats) != 2 {
			klog.Error(err, err1)
			pod.Status.ContainerStatuses = append(pod.Status.ContainerStatuses,
				v1.ContainerStatus{Name: cntr.Name})
			continue
		}

		var cpuPerc, memPerc string
		if cpu, ok := cntr.Resources["cpu"]; ok {
			cpuPerc = dynamicStats[0] + "/" + cpu
		} else {
			klog.Infof("Container %s doesn't specify cpu limit", cntr.Name)
			cpuPerc = dynamicStats[0] + "/"
		}

		if mem, ok := cntr.Resources["memory"]; ok {
			memPerc = dynamicStats[1] + "/" + mem
		} else {
			klog.Infof("Container %s doesn't specify mem limit", cntr.Name)
			memPerc = dynamicStats[1] + "/"
		}

		var containerState v1.ContainerState = v1.ContainerState{
			Status:     stats.State.Status,
			ExitCode:   stats.State.ExitCode,
			Error:      stats.State.Error,
			StartedAt:  stats.State.StartedAt,
			FinishedAt: stats.State.FinishedAt,
			CPUPerc:    cpuPerc,
			MemPerc:    memPerc,
		}

		pod.Status.ContainerStatuses = append(pod.Status.ContainerStatuses,
			v1.ContainerStatus{
				Name:  cntr.Name,
				State: containerState,
			})

		if cntr.Name == pod.Name+"-"+constants.InitialPauseContainer.Name {
			pod.Status.PodIP = stats.NetworkSettings.Networks[constants.WeaveNetworkName].IPAddress
		}
	}

	// get pod running statuses
	pod.Status.Phase = v1.PodRunning

	var pendingNum, runningNum, succeedNum, failedNum int = 0, 0, 0, 0

	for _, cntr := range pod.Spec.Containers {
		stats, err := pm.containerManager.ContainerStatus(context.TODO(), cntr.ID)

		dynamicStats, err1 := pm.containerManager.ContainerStats(context.TODO(), cntr.ID)

		if err != nil || err1 != nil || len(dynamicStats) != 2 {
			klog.Error(err, err1)
			pod.Status.ContainerStatuses = append(pod.Status.ContainerStatuses,
				v1.ContainerStatus{Name: cntr.Name})
			continue
		}

		var cpuPerc, memPerc string
		if cpu, ok := cntr.Resources["cpu"]; ok {
			cpuPerc = dynamicStats[0] + "/" + cpu
		} else {
			klog.Infof("Container %s doesn't specify cpu limit", cntr.Name)
			cpuPerc = dynamicStats[0] + "/"
		}

		if mem, ok := cntr.Resources["memory"]; ok {
			memPerc = dynamicStats[1] + "/" + mem
		} else {
			klog.Infof("Container %s doesn't specify mem limit", cntr.Name)
			memPerc = dynamicStats[1] + "/"
		}

		var containerState v1.ContainerState = v1.ContainerState{
			Status:     stats.State.Status,
			ExitCode:   stats.State.ExitCode,
			Error:      stats.State.Error,
			StartedAt:  stats.State.StartedAt,
			FinishedAt: stats.State.FinishedAt,
			CPUPerc:    cpuPerc,
			MemPerc:    memPerc,
		}

		pod.Status.ContainerStatuses = append(pod.Status.ContainerStatuses,
			v1.ContainerStatus{
				Name:  cntr.Name,
				State: containerState,
			})

		switch containerState.Status {
		case "created":
			pendingNum++
		case "running":
		case "paused":
		case "restarting":
			runningNum++
		case "exited":
			if containerState.ExitCode == 0 {
				succeedNum++
			} else {
				failedNum++
			}
		case "removing":
			succeedNum++
		case "dead":
			failedNum++
		default:
			//do nothing
			klog.Errorln("Unknown Container Status %s", containerState.Status)
		}
	}

	switch {
	case runningNum != 0:
		pod.Status.Phase = v1.PodRunning
	case failedNum != 0:
		pod.Status.Phase = v1.PodFailed
	case succeedNum == len(pod.Spec.Containers):
		pod.Status.Phase = v1.PodSucceeded
	default:
		pod.Status.Phase = v1.PodUnknown
	}

	// not need to update pm.podByName & pod.podByUID

	return pod.Status, nil
}

func (pm *podManager) CreatePodBridgeNetwork(CIDR string) error {
	_, err := pm.containerManager.CreateNetwork(context.TODO(), constants.WeaveNetworkName, CIDR)

	if err != nil {
		klog.Fatalf("Create Bridge Network %s Failed: %s", constants.WeaveNetworkName, err.Error())
		return err
	}

	return nil
}

func (pm *podManager) CheckDuplicate(pod *v1.Pod) bool {
	if existedPod, ok := pm.podByUID[pod.UID]; ok {
		// check if it't refreshed by myself
		existedPauseID := existedPod.Spec.InitialContainers[constants.InitialPauseContainerKey].ID
		if pausePod, ok := pod.Spec.InitialContainers[constants.InitialPauseContainerKey]; ok && pausePod.ID == existedPauseID {
			klog.Infof("Pod is refreshed by myself: Name: %s, UID: %s", pod.Name, pod.UID)
		} else {
			klog.Errorf("Duplicated pod: name %s, UID %s ", pod.Name, pod.UID)
		}
		return true
	}

	if _, ok := pm.podByName[pod.Name]; ok {
		klog.Errorf("Duplicated pod: name %s, UID %s ", pod.Name, pod.UID)
		return true
	}

	return false
}
