package invoker

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"k8s.io/klog"
	"minik8s.com/minik8s/pkg/aqualake/apis/actionchain"
	"minik8s.com/minik8s/pkg/aqualake/apis/constants"
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

func (ivk *Invoker) InvokeActionChain(chain actionchain.ActionChain, arg interface{}) (interface{}, error) {
	currActionName := chain.StartAt
	currActionArg := arg
	for {
		currAction := chain.Chain[currActionName]
		klog.Infof("Invoking Action Chain, currAction is %v", currAction)
		ret, err := ivk.invokeAction(currAction, currActionArg)
		if err != nil {
			klog.Errorf("Invoke Action Chain error, action returns a internal error: %v", err)
			return nil, err
		}
		if currAction.End {
			return ret, nil
		}
		if currAction.Type == actionchain.ACT_TASK {
			currActionName = currAction.Next
			currActionArg = append([]interface{}{}, ret)
		} else {
			f := false
			for _, choice := range currAction.Choices {
				klog.Infof("Scanning choice: %v", choice)
				if strings.HasPrefix(choice.Variable, "$.") {
					varName := strings.TrimPrefix(choice.Variable, "$.")
					if reflect.TypeOf(ret).Kind() == reflect.Map {
						ret = ret.(map[string]interface{})[varName]
					} else {
						return nil, errors.New("return value is not a map")
					}
				} else if strings.HasPrefix(choice.Variable, "$") {
					index, _ := strconv.Atoi(strings.TrimPrefix(choice.Variable, "$"))
					if reflect.TypeOf(ret).Kind() == reflect.Array {
						ret = ret.([]interface{})[index]
					} else {
						return nil, errors.New("return value is not a array")
					}
				}
				switch choice.Type {
				case actionchain.VAR_FLOAT:
					i := reflect.ValueOf(ret).Float()
					if choice.NumericEqual == i {
						currActionName = choice.Next
						currActionArg = append([]interface{}{}, i)
						f = true
					}
				case actionchain.VAR_STRING:
					i := reflect.ValueOf(ret).String()
					if choice.StringEqual == i {
						currActionName = choice.Next
						currActionArg = append([]interface{}{}, i)
						f = true
					}
				case actionchain.VAR_BOOL:
					i := reflect.ValueOf(ret).Bool()
					if choice.BooleanEqual == i {
						currActionName = choice.Next
						currActionArg = append([]interface{}{}, i)
						f = true
					}
				}
			}
			if !f {
				return nil, errors.New("no choice")
			}
		}
	}

}

func (ivk *Invoker) invokeAction(action actionchain.Action, arg interface{}) (interface{}, error) {
	podEntry, err := ivk.ppm.GetPod(action)
	defer ivk.ppm.FreePod(podEntry)
	if err != nil {
		return nil, err
	}

	if podEntry.NeedInstall{
		var installReq podserver.InstallFuncReq
		installReq.Name = action.Function
		installReq.Url = constants.CouchGetFileRequest(constants.FunctionDBId, action.Function, action.Function)
		installBuf, _ := json.Marshal(installReq)
		klog.Infof("install function request: %s", installBuf)
		resp, err := http.Post("http://"+podEntry.PodIP+":8698/installfunction",
			"application/json; charset=utf-8",
			bytes.NewReader(installBuf))
		if err != nil {
			klog.Errorf("install function err: %v", err)
			return nil, err
		}
		buf, _ := io.ReadAll(resp.Body)
		klog.Infof("install function response: %s", buf)
		podEntry.NeedInstall = false
	}

	var triggerReq podserver.TriggerReq
	triggerReq.Args = arg
	triggerBuf, _ := json.Marshal(triggerReq)
	klog.Infof("trigger function request: %s", triggerBuf)
	resp, err := http.Post("http://"+podEntry.PodIP+":8698/trigger",
		"application/json; charset=utf-8",
		bytes.NewReader(triggerBuf))
	if err != nil {
		klog.Errorf("trigger function err: %v", err)
		return nil, err
	}
	buf, _ := io.ReadAll(resp.Body)
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
