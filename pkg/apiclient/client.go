package apiclient

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"io"
	"k8s.io/klog"
	"minik8s.com/minik8s/config"
	"net/http"
	"strconv"
)

type ObjType int8
type OpType int8

const (
	OBJ_ALL_PODS     ObjType = 0
	OBJ_ALL_SERVICES ObjType = 1
	OBJ_ALL_REPLICAS ObjType = 2
	OBJ_ALL_NODES    ObjType = 3
	OBJ_POD          ObjType = 4
	OBJ_SERVICE      ObjType = 5
	OBJ_REPLICA      ObjType = 6
	OBJ_NODE         ObjType = 7

	OP_GET    OpType = 8
	OP_POST   OpType = 9
	OP_PUT    OpType = 10
	OP_DELETE OpType = 11
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
	case OBJ_POD:
		resp, err = http.Get(baseUrl + config.AC_WatchPod_Path)
	case OBJ_SERVICE:
		resp, err = http.Get(baseUrl + config.AC_WatchService_Path)
	case OBJ_REPLICA:
		resp, err = http.Get(baseUrl + config.AC_WatchReplica_Path)
	default:
		klog.Error("Invalid arguments!\n")
		return
	}
	if err != nil {
		fmt.Printf("error: %v", err)
		return
	}

	for {
		select {
		case <-ctx.Done():
			return
		default:
			reader := bufio.NewReader(resp.Body)
			buf, err := reader.ReadBytes(26)
			if err != nil {
				return
			}

			buf[len(buf)-1] = '\n'
			ch <- buf
		}
	}
}

/*
 这个函数用于发送rest请求，请正确填写参数
 objTy: which kind of object you want to operate on.
 opTy: which operation you want. use OP_XXX.
 key: used in GET, PUT, DELETE operation.
 value: used in PUT, POST operation.
*/
func Rest(id string, value []byte, objTy ObjType, opTy OpType) []byte {
	var resp *http.Response
	var err error
	url := config.AC_ServerAddr + ":" + strconv.Itoa(config.AC_ServerPort)
	switch objTy {
	case OBJ_ALL_PODS:
		url += config.AC_RestPods_Path
	case OBJ_ALL_SERVICES:
		url += config.AC_RestServices_Path
	case OBJ_ALL_REPLICAS:
		url += config.AC_RestReplicas_Path
	case OBJ_ALL_NODES:
		url += config.AC_RestNodes_Path
	case OBJ_POD:
		url += config.AC_RestPod_Path
	case OBJ_SERVICE:
		url += config.AC_RestService_Path
	case OBJ_REPLICA:
		url += config.AC_RestReplica_Path
	case OBJ_NODE:
		url += config.AC_RestNode_Path
	default:
		klog.Error("Invalid arguments!\n")
		return nil
	}
	switch opTy {
	case OP_GET:
		resp, err = http.Get(url)
	case OP_PUT:
		cli := http.Client{}
		req, _ := http.NewRequest(http.MethodPut, url+"/"+id, bytes.NewReader(value))
		resp, err = cli.Do(req)
	case OP_POST:
		cli := http.Client{}
		req, _ := http.NewRequest(http.MethodPost, url+"/"+id, bytes.NewReader(value))
		resp, err = cli.Do(req)
	case OP_DELETE:
		cli := http.Client{}
		req, _ := http.NewRequest(http.MethodDelete, url+"/"+id, bytes.NewReader(value))
		resp, err = cli.Do(req)
	default:
		klog.Error("Invalid arguments!\n")
		return nil
	}
	buf, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil
	}

	return buf
}
