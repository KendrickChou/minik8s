package podpoolmanager

import (
	"encoding/json"
	"errors"
	"fmt"
	"time"

	set "github.com/deckarep/golang-set"
	"k8s.io/klog"
	v1 "minik8s.com/minik8s/pkg/api/v1"
	"minik8s.com/minik8s/pkg/apiclient"
	"minik8s.com/minik8s/pkg/aqualake/apis/constants"
	"minik8s.com/minik8s/pkg/kubectl/commands"
)

type PodPoolManager interface {
	GetPod(podEnv string) *v1.Pod 
}

type podPoolManager struct {
	PodPool map[string]set.Set
}

func NewPodPoolManager() PodPoolManager {
	ppm := &podPoolManager{
		PodPool: make(map[string]set.Set),
	}

	for _, env := range constants.SupportedEnvs {
		ppm.PodPool[env] = set.NewSet()

		for i := 0; i < constants.DefaultPoolSetSize; i++ {
			pod, err := newGenericPod(env)

			if err != nil {
				klog.Errorf("Init Pod Pool Error: %s", err)
				continue
			}

			ppm.PodPool[env].Add(pod)
		}
	}

	return ppm
}


func newGenericPod(env string) (*v1.Pod, error) {
	pod := constants.NewPodConfig(env)
	buf, _ := json.Marshal(pod)
	resp := apiclient.Rest("", string(buf), apiclient.OBJ_POD, apiclient.OP_POST)

	var stat commands.StatusResponse
	json.Unmarshal(resp, &stat)

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

func (ppm *podPoolManager) GetPod(podEnv string) *v1.Pod {

}
