package apiclient

import (
	"context"
	"fmt"
	"k8s.io/klog"
	"minik8s.com/minik8s/config"
	"net/http"
	"strconv"
	"strings"
	"time"
)

type objType int8
type opType int8

const (
	OBJ_ALL_PODS     objType = 0
	OBJ_ALL_SERVICES objType = 1
	OBJ_ALL_REPLICAS objType = 2
	OBJ_POD          objType = 3
	OBJ_SERVICE      objType = 4
	OBJ_REPLICAS     objType = 5

	OP_GET    opType = 6
	OP_POST   opType = 7
	OP_PUT    opType = 8
	OP_DELETE opType = 9
)

func ExampleWatch() {
	ctx, cl := context.WithCancel(context.Background())
	watchChan := make(chan string)
	go watch(ctx, watchChan, OBJ_ALL_PODS)
	for {
		select {
		case <-time.After(time.Second * 5):
			cl()
			return
		case result := <-watchChan:
			klog.Infof("watch result: %v", result)
		}
	}
}
func ExampleRestOperate() {
	respStr := rest("", "test pod", OBJ_POD, OP_POST)

	//TODO:
	//get id from respStr ...
	fmt.Println(respStr)

	//replace xxxxx:
	rest("xxxxx", "", OBJ_POD, OP_GET)
	rest("xxxxx", "test pod2", OBJ_POD, OP_PUT)
	rest("xxxxx", "", OBJ_POD, OP_DELETE)
}

/*
	这个函数用于发送watch请求，请正确填写参数
	ctx: context for cancel watch task. watch task will be destroyed if ctx is done.
	ch: the channel where can you get response. use "for str := range <- ch" to get results.
	ty: which kind of object you want to watch. use OBJ_XXX.
*/
func watch(ctx context.Context, ch chan string, ty objType) {
	var resp *http.Response
	var err error
	baseUrl := config.AC_ServerAddr + ":" + strconv.Itoa(config.AC_ServerPort)
	switch ty {
	case OBJ_ALL_PODS:
		resp, err = http.Get(baseUrl + config.AC_WatchPods_Path)
	case OBJ_ALL_SERVICES:
		resp, err = http.Get(baseUrl + config.AC_WatchServices_Path)
	case OBJ_ALL_REPLICAS:
		resp, err = http.Get(baseUrl + config.AC_WatchReplicas_Path)
	case OBJ_POD:
		resp, err = http.Get(baseUrl + config.AC_WatchPod_Path)
	case OBJ_SERVICE:
		resp, err = http.Get(baseUrl + config.AC_WatchService_Path)
	case OBJ_REPLICAS:
		resp, err = http.Get(baseUrl + config.AC_WatchReplica_Path)
	default:
		klog.Error("Invalid arguments!\n")
		return
	}
	if err != nil {
		fmt.Printf("error: %v", err)
		return
	}

	buf := make([]byte, 128)
	podString := ""
	for {
		select {
		case <-ctx.Done():
			return
		default:
			readN, err := resp.Body.Read(buf)
			if err != nil {
				return
			}
			if readN < 128 {
				klog.Infof("read from server: %v\n", string(buf))
				ch <- podString
				podString = ""
			} else {
				podString += string(buf)
			}
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
func rest(id string, value string, objTy objType, opTy opType) string {
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
	case OBJ_POD:
		url += config.AC_RestPod_Path
	case OBJ_SERVICE:
		url += config.AC_RestService_Path
	case OBJ_REPLICAS:
		url += config.AC_RestReplica_Path
	default:
		klog.Error("Invalid arguments!\n")
		return ""
	}
	switch opTy {
	case OP_GET:
		resp, err = http.Get(url)
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
		return ""
	}
	buf := make([]byte, resp.ContentLength)
	_, err = resp.Body.Read(buf)
	if err != nil {
		return ""
	}
	return string(buf)
}
