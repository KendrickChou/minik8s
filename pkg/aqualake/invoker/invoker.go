package invoker

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"k8s.io/klog"
	"minik8s.com/minik8s/pkg/aqualake/apis/actionchain"
	"minik8s.com/minik8s/pkg/aqualake/apis/podserver"
	"minik8s.com/minik8s/pkg/aqualake/podpoolmanager"
	"net/http"
	"strings"
)

type Invoker struct {
	ppm *podpoolmanager.PodPoolManager
}

func NewInvoker() *Invoker {
	ivk := &Invoker{
		ppm: podpoolmanager.NewPodPoolManager(),
	}

	return ivk
}

func (ivk *Invoker) InvokeActionChain(chain actionchain.ActionChain, arg interface{}) error {
	currAction := chain.Chain[chain.StartAt]
	for {
		/*ret*/ _, err := ivk.invokeAction(currAction, arg)
		if err != nil {
			return err
		}
		if currAction.End {
			return nil
		}
		if currAction.Type == actionchain.ACT_TASK {
			currAction = chain.Chain[currAction.Next]
		} else {
			for _, _ = range currAction.Choices {

			}
		}
	}

}

func (ivk *Invoker) invokeAction(action actionchain.Action, arg interface{}) (interface{}, error) {
	env := action.Env
	pod, err := ivk.ppm.GetPod(env)
	if err != nil {
		return nil, err
	}
	resp, err := http.Post(pod.Status.PodIP+"/installfunction",
		"application/json; charset=utf-8",
		strings.NewReader(action.Function))
	if err != nil {
		klog.Errorf("install function err: %v", err)
		return nil, err
	}
	buf, _ := io.ReadAll(resp.Body)
	klog.Infof("install function response: %s", buf)

	var triggerReq podserver.TriggerReq
	triggerReq.Args = arg
	triggerBuf, _ := json.Marshal(triggerReq)
	resp, err = http.Post(pod.Status.PodIP+"/trigger",
		"application/json; charset=utf-8",
		bytes.NewReader(triggerBuf))
	if err != nil {
		klog.Errorf("trigger function err: %v", err)
		return nil, err
	}
	buf, _ = io.ReadAll(resp.Body)
	klog.Infof("trigger function response: %s", buf)

	var triggerResp podserver.TriggerResp
	err = json.Unmarshal(buf, &triggerResp)

	if err != nil {
		klog.Errorf("trigger function err: %v", err)
		return nil, err
	}
	if len(triggerResp.Err) > 0 {
		return nil, errors.New(triggerResp.Err)
	}
	return triggerResp.Ret, nil
}
