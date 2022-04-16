package pod

import v1 "minik8s.com/minik8s/pkg/api/v1"

type Manager interface {
	GetPods() []*v1.Pod

	GetPodByName(podFullName string) (*v1.Pod, bool)

	GetPodByUID(UID v1.UID) (*v1.Pod, bool)

	AddPod(pod *v1.Pod)

	UpdatePod(pod *v1.Pod)

	DeletePod(pod *v1.Pod)

}

type podManager struct {
	// Regular pods indexed by UID.
	podByUID map[v1.UID]*v1.Pod

	podByName map[string]*v1.Pod

}

func NewPodManager() Manager {
	pm := &podManager{}
	pm.podByUID = make(map[v1.UID]*v1.Pod)
	pm.podByName = make(map[string]*v1.Pod)

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

func (pm *podManager) GetPodByUID(UID v1.UID) (*v1.Pod, bool) {
	pod, ok := pm.podByUID[UID]

	return pod, ok
}

func (pm *podManager) AddPod(pod *v1.Pod) {
	
}

func (pm *podManager) UpdatePod(pod *v1.Pod) {

}

func (pm *podManager) DeletePod(pod *v1.Pod) {

}

