package apiclient

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"io"
	"k8s.io/klog"
	"minik8s.com/minik8s/config"
	v1 "minik8s.com/minik8s/pkg/api/v1"
	"net/http"
	"strconv"
	"strings"
)

type ObjType int8
type OpType int8

type HttpResponse struct {
	ID     string `json:"id,omitempty"`
	Status string `json:"status,omitempty"`
	Error  string `json:"error,omitempty"`
}

const (
	OBJ_ALL_PODS      ObjType = 0
	OBJ_ALL_SERVICES  ObjType = 1
	OBJ_ALL_REPLICAS  ObjType = 2
	OBJ_ALL_ENDPOINTS ObjType = 3
	OBJ_ALL_NODES     ObjType = 4
	OBJ_ALL_DNSS      ObjType = 10
	OBJ_ALL_GPUS      ObjType = 12
	OBJ_ALL_HPAS      ObjType = 15

	OBJ_POD      ObjType = 5
	OBJ_SERVICE  ObjType = 6
	OBJ_REPLICAS ObjType = 7
	OBJ_ENDPOINT ObjType = 8
	OBJ_NODE     ObjType = 9
	OBJ_DNS      ObjType = 11
	OBJ_GPU      ObjType = 13
	OBJ_HPA      ObjType = 14

	OP_GET    OpType = 60
	OP_POST   OpType = 70
	OP_PUT    OpType = 80
	OP_DELETE OpType = 90
)

/*
	这个函数用于发送watch请求，请正确填写参数
	ctx: context for cancel watch task. watch task will be destroyed if ctx is done.
	ch: the channel where can you get response. use "for str := range <- ch" to get results.
	ty: which kind of object you want to watch. use OBJ_XXX.
*/
func Watch(ctx context.Context, ch chan []byte, ty ObjType) {
	var resp *http.Response
	var err error
	baseUrl := config.AC_ServerAddr + ":" + strconv.Itoa(config.AC_ServerPort)
	switch ty {
	case OBJ_ALL_PODS:
		resp, err = http.Get(baseUrl + config.AC_WatchPods_Path)
	case OBJ_ALL_NODES:
		resp, err = http.Get(baseUrl + config.AC_WatchNodes_Path)
	case OBJ_ALL_SERVICES:
		resp, err = http.Get(baseUrl + config.AC_WatchServices_Path)
	case OBJ_ALL_REPLICAS:
		resp, err = http.Get(baseUrl + config.AC_WatchReplicas_Path)
	case OBJ_ALL_HPAS:
		resp, err = http.Get(baseUrl + config.AC_WatchHPAs_Path)
	case OBJ_ALL_ENDPOINTS:
		resp, err = http.Get(baseUrl + config.AC_WatchEndpoints_Path)
	case OBJ_ALL_DNSS:
		resp, err = http.Get(baseUrl + config.AC_WatchDnss_Path)
	case OBJ_ALL_GPUS:
		resp, err = http.Get(baseUrl + config.AC_WatchGpus_Path)
	case OBJ_POD:
		resp, err = http.Get(baseUrl + config.AC_WatchPod_Path)
	case OBJ_SERVICE:
		resp, err = http.Get(baseUrl + config.AC_WatchService_Path)
	case OBJ_REPLICAS:
		resp, err = http.Get(baseUrl + config.AC_WatchReplica_Path)
	case OBJ_HPA:
		resp, err = http.Get(baseUrl + config.AC_WatchHPA_Path)
	case OBJ_ENDPOINT:
		resp, err = http.Get(baseUrl + config.AC_WatchEndpoint_Path)
	default:
		klog.Error("Invalid arguments!\n")
		return
	}
	if err != nil {
		klog.Errorf("error: %v", err)
		klog.Errorf("Rewatch...\n")
		go Watch(ctx, ch, ty)
		return
	}

	reader := bufio.NewReader(resp.Body)
	for {
		select {
		case <-ctx.Done():
			return
		default:
			buf, err := reader.ReadBytes(26)
			//var buf []byte
			//b := make([]byte, 1)
			//for {
			//	_, _ = resp.Body.Read(b)
			//	buf = append(buf, b[0])
			//	if b[0] == 26 {
			//		break
			//	}
			//}
			if err != nil {
				klog.Errorf("error: %v", err)
				klog.Errorf("Rewatch...\n")
				go Watch(ctx, ch, ty)
				return
			}

			buf[len(buf)-1] = '\n'
			//klog.Infof("buf: %s\n", buf)
			ch <- buf
		}
	}
}

func GetAll(objType ObjType) []byte {
	var resp *http.Response
	var err error
	url := config.AC_ServerAddr + ":" + strconv.Itoa(config.AC_ServerPort)
	switch objType {
	case OBJ_ALL_PODS:
		url += config.AC_RestPods_Path
	case OBJ_ALL_SERVICES:
		url += config.AC_RestServices_Path
	case OBJ_ALL_REPLICAS:
		url += config.AC_RestReplicas_Path
	case OBJ_ALL_ENDPOINTS:
		url += config.AC_RestEndpoints_Path
	case OBJ_ALL_HPAS:
		url += config.AC_RestHPAs_Path
	case OBJ_POD:
		url += config.AC_RestPod_Path
	case OBJ_SERVICE:
		url += config.AC_RestService_Path
	case OBJ_REPLICAS:
		url += config.AC_RestReplica_Path
	case OBJ_ENDPOINT:
		url += config.AC_RestEndpoint_Path
	case OBJ_HPA:
		url += config.AC_RestHPA_Path
	default:
		klog.Error("Invalid arguments!\n")
		return nil
	}
	resp, err = http.Get(url)
	buf, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil
	}
	return buf
}

/*
	这个函数用于发送rest请求，请正确填写参数
	objTy: which kind of object you want to operate on.
	opTy: which operation you want. use OP_XXX.
	key: used in GET, PUT, DELETE operation.
	value: used in PUT, POST operation.
*/
func Rest(id string, value string, objTy ObjType, opTy OpType) []byte {
	var resp *http.Response
	var err error
	url := config.AC_ServerAddr + ":" + strconv.Itoa(config.AC_ServerPort)
	switch objTy {
	case OBJ_ALL_PODS:
		url += config.AC_RestPods_Path
	case OBJ_ALL_NODES:
		url += config.AC_RestNodes_Path
	case OBJ_ALL_SERVICES:
		url += config.AC_RestServices_Path
	case OBJ_ALL_REPLICAS:
		url += config.AC_RestReplicas_Path
	case OBJ_ALL_HPAS:
		url += config.AC_RestHPAs_Path
	case OBJ_ALL_ENDPOINTS:
		url += config.AC_RestEndpoints_Path
	case OBJ_ALL_DNSS:
		url += config.AC_RestDnss_Path
	case OBJ_ALL_GPUS:
		url += config.AC_RestGpus_Path
	case OBJ_POD:
		url += config.AC_RestPod_Path
	case OBJ_NODE:
		url += config.AC_RestNode_Path
	case OBJ_SERVICE:
		url += config.AC_RestService_Path
	case OBJ_REPLICAS:
		url += config.AC_RestReplica_Path
	case OBJ_HPA:
		url += config.AC_RestHPA_Path
	case OBJ_ENDPOINT:
		url += config.AC_RestEndpoint_Path
	case OBJ_DNS:
		url += config.AC_RestDns_Path
	case OBJ_GPU:
		url += config.AC_RestGpu_Path
	default:
		klog.Error("Invalid arguments!\n")
		return nil
	}
	switch opTy {
	case OP_GET:
		resp, err = http.Get(url + "/" + id)
	case OP_PUT:
		cli := http.Client{}
		req, _ := http.NewRequest(http.MethodPut, url+"/"+id, strings.NewReader(value))
		resp, err = cli.Do(req)
	case OP_POST:
		cli := http.Client{}
		req, _ := http.NewRequest(http.MethodPost, url+"/"+id, strings.NewReader(value))
		resp, err = cli.Do(req)
	case OP_DELETE:
		cli := http.Client{}
		req, _ := http.NewRequest(http.MethodDelete, url+"/"+id, strings.NewReader(value))
		resp, err = cli.Do(req)
	default:
		klog.Error("Invalid arguments!\n")
		return nil
	}

	buf, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil
	}
	resp.Body.Close()

	return buf
}

// GetPodStatusHttp deprecated
func GetPodStatusHttp(pod *v1.Pod) []byte {
	var resp *http.Response
	var err error
	url := config.AC_ServerAddr + ":" + strconv.Itoa(config.AC_ServerPort)
	url += "/innode/" + pod.Spec.NodeName + "/podstatus/" + pod.UID

	resp, err = http.Get(url)
	buf, err := io.ReadAll(resp.Body)
	resp.Body.Close()
	if err != nil {
		klog.Error("http get error\n")
		return nil
	}

	return buf
}

func PostEndpoint(endpoint *v1.Endpoint) string {
	var resp *http.Response
	var err error
	url := config.AC_ServerAddr + ":" + strconv.Itoa(config.AC_ServerPort)

	cli := http.Client{}
	epBytes, err := json.Marshal(endpoint)
	if err != nil {
		klog.Error("Json marshall error\n")
		return ""
	}

	req, _ := http.NewRequest(http.MethodPost, url+config.AC_RestEndpoint_Path+"/"+endpoint.UID, bytes.NewReader(epBytes))
	resp, err = cli.Do(req)

	buf, err := io.ReadAll(resp.Body)
	resp.Body.Close()
	if err != nil {
		klog.Error("Read response error\n")
		return ""
	}

	var response HttpResponse
	err = json.Unmarshal(buf, &response)
	if err != nil {
		klog.Error("Json unmarshal error\n")
		return ""
	}

	if response.Status == "OK" {
		return response.ID
	} else {
		return ""
	}
}

func UpdateEndpoint(ep *v1.Endpoint) bool {
	epByte, err := json.Marshal(ep)
	if err != nil {
		return false
	}

	responseBytes := Rest(ep.UID, string(epByte), OBJ_ENDPOINT, OP_PUT)

	var responseBody HttpResponse
	err = json.Unmarshal(responseBytes, &responseBody)
	if err != nil {
		klog.Error("Json unmarshal error\n")
		return false
	}

	if responseBody.Status == "OK" {
		return true
	} else {
		return false
	}
}

// PostPod returns the UID of the pod
func PostPod(pod *v1.Pod) string {
	podByte, err := json.Marshal(pod)
	if err != nil {
		return ""
	}

	responseBytes := Rest(pod.UID, string(podByte), OBJ_POD, OP_POST)

	var responseBody HttpResponse
	err = json.Unmarshal(responseBytes, &responseBody)
	if err != nil {
		klog.Error("Json unmarshal error\n")
		return ""
	}

	if responseBody.Status == "OK" {
		return responseBody.ID
	} else {
		return ""
	}
}

func UpdatePod(pod *v1.Pod) bool {
	podByte, err := json.Marshal(pod)
	if err != nil {
		return false
	}

	responseBytes := Rest(pod.UID, string(podByte), OBJ_POD, OP_PUT)

	var responseBody HttpResponse
	err = json.Unmarshal(responseBytes, &responseBody)
	if err != nil {
		klog.Error("Json unmarshal error\n")
		return false
	}

	if responseBody.Status == "OK" {
		return true
	} else {
		return false
	}
}

func UpdateReplicaSet(rs *v1.ReplicaSet) bool {
	rsByte, err := json.Marshal(rs)
	if err != nil {
		return false
	}

	responseBytes := Rest(rs.UID, string(rsByte), OBJ_REPLICAS, OP_PUT)

	var responseBody HttpResponse
	err = json.Unmarshal(responseBytes, &responseBody)
	if err != nil {
		klog.Error("Json unmarshal error\n")
		return false
	}

	if responseBody.Status == "OK" {
		return true
	} else {
		return false
	}
}

func UpdateHorizontalPodAutoscaler(hpa *v1.HorizontalPodAutoscaler) bool {
	rsByte, err := json.Marshal(hpa)
	if err != nil {
		return false
	}

	responseBytes := Rest(hpa.UID, string(rsByte), OBJ_HPA, OP_PUT)

	var responseBody HttpResponse
	err = json.Unmarshal(responseBytes, &responseBody)
	if err != nil {
		klog.Error("Json unmarshal error\n")
		return false
	}

	if responseBody.Status == "OK" {
		return true
	} else {
		return false
	}
}

func DeleteEndpoint(epID string) bool {
	responseBytes := Rest(epID, "", OBJ_ENDPOINT, OP_DELETE)

	var responseBody HttpResponse
	err := json.Unmarshal(responseBytes, &responseBody)
	if err != nil {
		klog.Error("Json unmarshal error\n")
		return false
	}

	if responseBody.Status == "OK" {
		return true
	} else {
		return false
	}
}

func DeletePod(podID string) bool {
	responseBytes := Rest(podID, "", OBJ_POD, OP_DELETE)

	var responseBody HttpResponse
	err := json.Unmarshal(responseBytes, &responseBody)
	if err != nil {
		klog.Error("Json unmarshal error\n")
		return false
	}

	if responseBody.Status == "OK" {
		return true
	} else {
		return false
	}
}
