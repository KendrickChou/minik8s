package pod

import (
	"context"
	"errors"
	"fmt"

	"k8s.io/klog/v2"
	v1 "minik8s.com/minik8s/pkg/api/v1"
	"minik8s.com/minik8s/pkg/kubelet/container"
)

type PodManager interface {
	GetPods() []v1.Pod

	GetPodByName(podFullName string) (v1.Pod, bool)

	GetPodByUID(UID string) (v1.Pod, bool)

	AddPod(pod *v1.Pod) error

	UpdatePod(pod *v1.Pod) error

	DeletePod(UID string) error

	PodStatus(UID string) (v1.PodStatus, error)
}

type podManager struct {
	// Regular pods indexed by UID.
	podByUID map[string]*v1.Pod

	podByName map[string]*v1.Pod

	containerManager container.ContainerManager
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
	return pm
}

func (pm *podManager) GetPods() []v1.Pod {
	pods := make([]v1.Pod, len(pm.podByUID))

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

	if _, ok := pm.podByUID[pod.UID]; ok {
		err := "duplicated pod UID: " + string(pod.UID)

		klog.Errorln(err)
		return errors.New(err)
	}

	if dupPod, ok := pm.podByName[pod.Name]; ok && dupPod.Namespace == pod.Namespace {
		err := "duplicated pod name: " + pod.Name + " in namespace: " + pod.Namespace

		klog.Errorln(err)
		return errors.New(err)
	}

	pm.podByUID[pod.UID] = pod
	pm.podByName[pod.Name] = pod

	for _, container := range pod.Spec.Containers {
		container.Name = pod.Name + "-" + container.Name
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

	klog.Infof("Inspect Pod %s status", pod.Name)

	pod.Status.ContainerStatuses = []v1.ContainerStatus{}
	pod.Status.Phase = v1.PodRunning

	var pendingNum, runningNum, succeedNum, failedNum int = 0, 0, 0, 0

	for _, cntr := range pod.Spec.Containers {
		stats, err := pm.containerManager.ContainerStatus(context.TODO(), cntr.ID)

		if err != nil {
			klog.Errorln(err)
			pod.Status.ContainerStatuses = append(pod.Status.ContainerStatuses,
				v1.ContainerStatus{Name: cntr.Name})
			continue
		}

		var containerState v1.ContainerState = v1.ContainerState{
			Status:     stats.Status,
			ExitCode:   stats.ExitCode,
			Error:      stats.Error,
			StartedAt:  stats.StartedAt,
			FinishedAt: stats.FinishedAt,
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
	case pendingNum+runningNum+succeedNum+failedNum < len(pod.Spec.Containers):
		pod.Status.Phase = v1.PodUnknown
	case pendingNum != 0:
		pod.Status.Phase = v1.PodPending
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
