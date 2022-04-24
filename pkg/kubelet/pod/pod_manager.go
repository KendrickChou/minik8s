package pod

import (
	"context"
	"errors"

	"k8s.io/klog/v2"
	v1 "minik8s.com/minik8s/pkg/api/v1"
	"minik8s.com/minik8s/pkg/kubelet/container"
)

type PodManager interface {
	GetPods() []*v1.Pod

	GetPodByName(podFullName string) (*v1.Pod, bool)

	GetPodByUID(UID v1.UID) (v1.Pod, bool)

	AddPod(pod *v1.Pod) error

	UpdatePod(pod *v1.Pod)

	DeletePod(pod *v1.Pod)

}

type podManager struct {
	// Regular pods indexed by UID.
	podByUID map[v1.UID]*v1.Pod

	podByName map[string]*v1.Pod

	containerManager container.ContainerManager

}

func NewPodManager() PodManager {
	pm := &podManager{}
	pm.podByUID = make(map[v1.UID]*v1.Pod)
	pm.podByName = make(map[string]*v1.Pod)

	newContainerManager,err := container.NewContainerManager()

	if err != nil {
		klog.Errorln(err)
	}

	pm.containerManager = newContainerManager
	return pm
}

func (pm *podManager) GetPods() []*v1.Pod {
	pods := make([]*v1.Pod, len(pm.podByUID))

	for _, value := range pm.podByUID {
		pods = append(pods, value)
	}

	return pods
}

func (pm *podManager) GetPodByName(podFullName string) (*v1.Pod, bool) {
	pod, ok := pm.podByName[podFullName]

	return pod, ok
}

func (pm *podManager) GetPodByUID(UID v1.UID) (v1.Pod, bool) {
	pod, ok := pm.podByUID[UID]

	return *pod, ok
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
		err := "duplicated pod name: " + pod.Name + " in namespace: " +  pod.Namespace

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

func (pm *podManager) UpdatePod(pod *v1.Pod) {
	if pod.UID == "" {
		klog.Errorln("pod UID is empty")
		return
	}

	oldPod, ok := pm.podByUID[pod.UID]

	if !ok {
		klog.Errorln("Pod doesn't exist, UID: %s", pod.UID)
		return
	}

	delete(pm.podByName, oldPod.Name)
	pm.podByName[pod.Name] = pod
	pm.podByUID[pod.UID] = pod

	for _, container := range oldPod.Spec.Containers {
		pm.containerManager.RemoveContainer(context.TODO(), container)
	}

	for _, container := range pod.Spec.Containers {
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
}

func (pm *podManager) DeletePod(pod *v1.Pod) {
	klog.Infof("Delete Pod %s", pod.Name)

	if pod.UID == "" {
		klog.Errorln("pod UID is empty")
		return
	}

	for _, container := range pod.Spec.Containers {
		err := pm.containerManager.StopContainer(context.TODO(), container)

		if err != nil {
			klog.Errorf("Remove Container %s: %s", container.Name, err)
		}

		err = pm.containerManager.RemoveContainer(context.TODO(), container)

		if err != nil {
			klog.Errorf("Remove Container %s: %s", container.Name, err)
		}
	}

	delete(pm.podByName, pod.Name)
	delete(pm.podByUID, pod.UID)
}

