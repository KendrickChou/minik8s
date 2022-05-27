package podpoolmanager

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"minik8s.com/minik8s/pkg/aqualake/apis/actionchain"
	"sync"
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

type PodEntry struct {
	PodIP       string
	NeedInstall bool
	uid         string
	cancel      context.CancelFunc
	mtx         sync.Mutex
}

type PodPoolManager struct {
	pp map[string][]*PodEntry
}

var bigLock sync.Mutex

func NewPodPoolManager() *PodPoolManager {
	ppm := &PodPoolManager{
		pp: make(map[string][]*PodEntry),
	}

	return ppm
}

func newGenericPodEntry(env string) (*PodEntry, error) {
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
	for iter := 0; iter < 20; iter++ {
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
			return &PodEntry{PodIP: pod.Status.PodIP, NeedInstall: true, mtx: sync.Mutex{}, uid: pod.UID, cancel: func() {}}, nil
		}
	}

	errInfo := fmt.Sprintf("Pod %s has no response for a long time", pod.ObjectMeta.Name)
	klog.Error(errInfo)
	return nil, errors.New(errInfo)
}

func (ppm *PodPoolManager) GetPod(action actionchain.Action) (*PodEntry, error) {
	for _, pe := range ppm.pp[action.Function] {
		if pe.mtx.TryLock() {
			klog.Infof("reuse podIP %v, reset delete timer..", pe.PodIP)
			pe.cancel()
			newCtx, cancel := context.WithCancel(context.Background())
			go ppm.DeletePodAfter5Minute(newCtx, pe, action)
			pe.cancel = cancel
			return pe, nil
		}
	}

	pe, err := newGenericPodEntry(action.Env)
	if err != nil {
		return nil, err
	}
	bigLock.Lock()
	ppm.pp[action.Function] = append(ppm.pp[action.Function], pe)
	bigLock.Unlock()
	pe.mtx.Lock()
	ctx, cancel := context.WithCancel(context.Background())
	go ppm.DeletePodAfter5Minute(ctx, pe, action)
	pe.cancel = cancel
	return pe, err
}

func (ppm *PodPoolManager) FreePod(pe *PodEntry) {
	if pe != nil {
		pe.mtx.Unlock()
	}
}

func (ppm *PodPoolManager) DeletePodAfter5Minute(ctx context.Context, pe *PodEntry, action actionchain.Action) {
	if pe != nil {
		for {
			select {
			case <-ctx.Done():
				klog.Infof("ctx canceled, delete task..")
				return
			case <-time.After(3 * time.Minute):
				ppm.DeletePod(pe, action)
			}
		}
	}
}
func (ppm *PodPoolManager) CancelDeletePod(pe *PodEntry) {
	pe.cancel()
}

func (ppm *PodPoolManager) DeletePod(pe *PodEntry, action actionchain.Action) {
	bigLock.Lock()
	for i, podEntry := range ppm.pp[action.Function] {
		if podEntry == pe {
			klog.Infof("forget podIP %v for action %v", pe.PodIP, action)
			_ = apiclient.Rest(pe.uid, "", apiclient.OBJ_POD, apiclient.OP_DELETE)
			ppm.pp[action.Function] = append(ppm.pp[action.Function][:i], ppm.pp[action.Function][i+1:]...)
			break
		}
		i++
	}
	bigLock.Unlock()
}
