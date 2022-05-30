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
	pp           map[string][]*PodEntry
	scale        map[string]int
	scaleSigChan map[string]chan int
	cancels      map[string][]context.CancelFunc
	readyPodChan map[string]chan *PodEntry
	bigLock      sync.Mutex
}

func NewPodPoolManager() *PodPoolManager {
	ppm := &PodPoolManager{
		pp:           make(map[string][]*PodEntry),
		scale:        make(map[string]int),
		scaleSigChan: make(map[string]chan int),
		cancels:      make(map[string][]context.CancelFunc),
		readyPodChan: make(map[string]chan *PodEntry),
	}
	ppm.scale["python"] = 3
	ppm.readyPodChan["python"] = make(chan *PodEntry)
	for i := 0; i < ppm.scale["python"]; i++ {
		ctx, cancel := context.WithCancel(context.Background())
		go ppm.provider(ctx, "python")
		ppm.cancels["python"] = append(ppm.cancels["python"], cancel)
	}
	ppm.scaleSigChan["python"] = make(chan int)
	go ppm.scaler("python")
	return ppm
}

func (ppm *PodPoolManager) provider(ctx context.Context, env string) {
	for {
		select {
		case <-ctx.Done():
			klog.Infof("Provider: exit...")
			return
		default:
			pod, err := newGenericPodEntry(env)
			if err != nil {
				klog.Errorf("Cannot make more pods on %v", env)
				break
			}
			ppm.readyPodChan[env] <- pod
		}
	}
}

func (ppm *PodPoolManager) scaler(env string) {
	for {
		select {
		case <-time.After(time.Minute * 2):
			ppm.bigLock.Lock()
			s := (ppm.scale[env] + 1) / 2
			if s > 0 {
				i := 0
				for i = 0; i < s; i++ {
					ppm.scale[env]--
					klog.Infof("Scaling: reduce provider for env: %v, current scale: %d", env, ppm.scale[env])
					ppm.cancels[env][i]()
				}
				ppm.cancels[env] = append([]context.CancelFunc{}, ppm.cancels[env][i:]...)
			} else {
				clear := false
				for !clear {
					select {
					case podEntry := <-ppm.readyPodChan[env]:
						deregisterPod(podEntry)
					default:
						clear = true
					}
				}
			}
			ppm.bigLock.Unlock()
		case <-ppm.scaleSigChan[env]:
			ppm.bigLock.Lock()
			ppm.scale[env]++
			klog.Infof("Scaling: add provider for env: %v, current scale: %d", env, ppm.scale[env])
			ctx, cancel := context.WithCancel(context.Background())
			go ppm.provider(ctx, env)
			ppm.cancels[env] = append(ppm.cancels[env], cancel)
			ppm.bigLock.Unlock()
		}
	}
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
	var pe *PodEntry
	for _, pe = range ppm.pp[action.Function] {
		if pe.mtx.TryLock() {
			klog.Infof("reuse podIP %v, reset delete timer..", pe.PodIP)
			pe.cancel()
			newCtx, cancel := context.WithCancel(context.Background())
			go ppm.DeletePodAfter5Minute(newCtx, pe, action)
			pe.cancel = cancel
			return pe, nil
		}
	}
	needMorePod := true
	for {
		select {
		case pe = <-ppm.readyPodChan[action.Env]:
			ppm.bigLock.Lock()
			ppm.pp[action.Function] = append(ppm.pp[action.Function], pe)
			ppm.bigLock.Unlock()
			pe.mtx.Lock()
			ctx, cancel := context.WithCancel(context.Background())
			go ppm.DeletePodAfter5Minute(ctx, pe, action)
			pe.cancel = cancel
			return pe, nil
		default:
			klog.Infof("no pod is ready for env: %v", action.Env)
			if needMorePod {
				needMorePod = false
				ppm.scaleSigChan[action.Env] <- 1
			}
			time.Sleep(10 * time.Second)
		}
	}
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
	ppm.bigLock.Lock()
	for i, podEntry := range ppm.pp[action.Function] {
		if podEntry == pe {
			klog.Infof("forget podIP %v for action %v", pe.PodIP, action)
			deregisterPod(pe)
			ppm.pp[action.Function] = append(ppm.pp[action.Function][:i], ppm.pp[action.Function][i+1:]...)
			break
		}
		i++
	}
	ppm.bigLock.Unlock()
}

func deregisterPod(pe *PodEntry) {
	_ = apiclient.Rest(pe.uid, "", apiclient.OBJ_POD, apiclient.OP_DELETE)
}
