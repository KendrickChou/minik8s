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

type GetPodResponse struct {
	Key  string `json:"key"`
	Pod  v1.Pod `json:"value"`
	Type string `json:"type"`
}

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
		time.Sleep(3 * time.Second)

		resp = apiclient.Rest(pod.UID, "", apiclient.OBJ_POD, apiclient.OP_GET)

		var getPodResp GetPodResponse
		err = json.Unmarshal(resp, &getPodResp)

		if err != nil {
			klog.Errorf("Unmarshal GetPodRequest Error")
			return nil, err
		}
		*pod = getPodResp.Pod

		if pod.Status.PodIP != "" {
			klog.Infof("Pod %s is Ready to go", pod.ObjectMeta.Name)
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
