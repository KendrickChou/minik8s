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
	"reflect"
	"strconv"
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
	currActionName := chain.StartAt
	currActionArg := arg
	for {
		currAction := chain.Chain[currActionName]
		klog.Infof("Invoking Action Chain, currAction is %v", currActionName)
		ret, err := ivk.invokeAction(currAction, currActionArg)
		if err != nil {
			klog.Errorf("Invoke Action Chain error, action returns a internal error: %v", err)
			return err
		}
		if currAction.End {
			return nil
		}
		if currAction.Type == actionchain.ACT_TASK {
			currAction = chain.Chain[currAction.Next]
		} else {
			for _, choice := range currAction.Choices {
				if strings.HasPrefix(choice.Variable, "$.") {
					varName := strings.TrimPrefix(choice.Variable, "$.")
					if reflect.TypeOf(ret).Kind() == reflect.Map {
						ret = ret.(map[string]interface{})[varName]
					} else {
						return errors.New("return value is not a map")
					}
				} else if strings.HasPrefix(choice.Variable, "$") {
					index, _ := strconv.Atoi(strings.TrimPrefix(choice.Variable, "$"))
					if reflect.TypeOf(ret).Kind() == reflect.Array {
						ret = ret.([]interface{})[index]
					} else {
						return errors.New("return value is not a array")
					}
				}
				switch choice.Type {
				case actionchain.VAR_INT:
					i := reflect.ValueOf(ret).Int()
					if choice.NumericEqual == i {
						currActionName = choice.Next
						currActionArg = ret
					}
				case actionchain.VAR_STRING:
					i := reflect.ValueOf(ret).String()
					if choice.StringEqual == i {
						currActionName = choice.Next
					}
				case actionchain.VAR_BOOL:
					i := reflect.ValueOf(ret).Bool()
					if choice.BooleanEqual == i {
						currActionName = choice.Next
					}
				}
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
