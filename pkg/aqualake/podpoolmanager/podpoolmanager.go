package podpoolmanager

import (
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"k8s.io/klog"
	v1 "minik8s.com/minik8s/pkg/api/v1"
	"minik8s.com/minik8s/pkg/apiclient"
	"minik8s.com/minik8s/pkg/aqualake/apis/constants"
	"minik8s.com/minik8s/pkg/kubectl/commands"
)

type PodPoolManager struct {
	pp map[string][]*v1.Pod
}

func NewPodPoolManager() *PodPoolManager {
	ppm := &PodPoolManager{
		pp: make(map[string][]*v1.Pod),
	}

	return ppm
}

func newGenericPod(env string) (*v1.Pod, error) {
	pod := constants.NewPodConfig(env)
	buf, _ := json.Marshal(pod)
	resp := apiclient.Rest("", string(buf), apiclient.OBJ_POD, apiclient.OP_POST)

	var stat commands.StatusResponse
	err := json.Unmarshal(resp, &stat)
	if err != nil {
		return nil, err
	}

	if stat.Status != "OK" {
		errInfo := fmt.Sprintf("Init Pod %s Error: %s", pod.ObjectMeta.Name, stat.Error)
		klog.Errorf(errInfo)
		return nil, errors.New(errInfo)
	} else {
		pod.ObjectMeta.UID = stat.Id
	}

	// wait pod IP ready
	for iter := 0; iter < 10; iter++ {
		time.Sleep(3)

		resp = apiclient.GetPodStatus(pod)

		if resp == nil {
			errInfo := "GetPodStatus Response nil"
			klog.Error(errInfo)
			return nil, errors.New(errInfo)
		}

		var podStatus v1.PodStatus
		err := json.Unmarshal(resp, &podStatus)

		if err != nil {
			klog.Errorf("Unmarshal GetPodStatus Request Error")
			return nil, err
		}

		if podStatus.PodIP != "" {
			klog.Infof("Pod %s is Ready to go", pod.ObjectMeta.Name)
			pod.Status = podStatus
			return pod, nil
		}
	}

	errInfo := fmt.Sprintf("Pod %s has no response for a long time", pod.ObjectMeta.Name)
	klog.Error(errInfo)
	return nil, errors.New(errInfo)
}

func (ppm *PodPoolManager) GetPod(podEnv string) (*v1.Pod, error) {
	pod, err := newGenericPod(podEnv)
	if err != nil {
		return nil, err
	}
	ppm.pp[podEnv] = append(ppm.pp[podEnv], pod)
	return pod, err
}
