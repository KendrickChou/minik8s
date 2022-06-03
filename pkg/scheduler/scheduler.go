package scheduler

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"strconv"
	"sync"
	"time"

	"k8s.io/klog/v2"
	"minik8s.com/minik8s/config"
	v1 "minik8s.com/minik8s/pkg/api/v1"
	"minik8s.com/minik8s/pkg/apiclient"
)

type PodRequest struct {
	Key string `json:"key"`

	Pod v1.Pod `json:"value"`

	Type string `json:"type"`
}

type NodeRequest struct {
	Key string `json:"key"`

	Node v1.Node `json:"value"`

	Type string `json:"type"`
}

var podMap map[string]v1.Pod

var pendingPod []string

var nodeMap map[string]v1.Node
var cancelMap map[string]context.CancelFunc

var mtx sync.Mutex

var shed func(pod v1.Pod) bool

var currNum int

func Init() {
	podMap = make(map[string]v1.Pod)
	nodeMap = make(map[string]v1.Node)
	cancelMap = make(map[string]context.CancelFunc)
	currNum = 0
	switch config.SCHED_STRATEGY {
	case "SIMPLE":
		shed = shed_simple
	case "RR":
		shed = shed_rr
	}
}

func Run() {
	ctx, cancle := context.WithCancel(context.Background())
	defer cancle()

	podChan := make(chan []byte)
	nodeChan := make(chan []byte)
	//watch pod and node
	go apiclient.Watch(ctx, podChan, apiclient.OBJ_ALL_PODS)
	go apiclient.Watch(ctx, nodeChan, apiclient.OBJ_ALL_NODES)

	//add existed nodes
	nodes_raw := apiclient.Rest("", "", apiclient.OBJ_ALL_NODES, apiclient.OP_GET)
	var nodes []NodeRequest
	err := json.Unmarshal(nodes_raw, &nodes)
	if err != nil {
		klog.Error("Unmarshal Pods Failed: %v", err)
	} else {
		for _, nodeReq := range nodes {
			nodeMap[nodeReq.Key] = nodeReq.Node
			newCtx, cancel := context.WithCancel(context.Background())
			go deleteNodeAfter30s(newCtx, nodeReq.Key)
			cancelMap[nodeReq.Key] = cancel
		}
		klog.Infof("Current node num: %v", len(nodeMap))
	}
	//add existed pods
	pods_raw := apiclient.Rest("", "", apiclient.OBJ_ALL_PODS, apiclient.OP_GET)
	var pods []PodRequest
	err = json.Unmarshal(pods_raw, &pods)
	if err != nil {
		klog.Error("Unmarshal Pods Failed: %v", err)
	} else {
		for _, podReq := range pods {
			podMap[podReq.Key] = podReq.Pod
			shed(podReq.Pod)
		}
		klog.Infof("Current pod num: %v", len(podMap))
	}

	//handle watch results
	for {
		select {
		case rawBytes := <-podChan:
			req := &PodRequest{}
			err := json.Unmarshal(rawBytes, req)

			if err != nil {

				klog.Error("Unmarshal Pod Change Req Failed: %v", err)
			} else {
				go handlePodChanRequest(req)
			}
		case rawBytes := <-nodeChan:
			req := &NodeRequest{}
			err := json.Unmarshal(rawBytes, req)

			if err != nil {
				klog.Error("Unmarshal Node Change Req Failed: %v", err)
			} else {
				go handleNodeChanRequest(req)
			}
		}
	}
}

func handlePodChanRequest(req *PodRequest) {
	switch req.Type {
	case "PUT":
		mtx.Lock()
		if _, exist := podMap[req.Key]; exist {
			podMap[req.Key] = req.Pod
			klog.Infof("Pod Changed: Key[%v] Value[...]", req.Key)
		} else {
			podMap[req.Key] = req.Pod
			klog.Infof("New Pod Added: Key[%v] Value[...]", req.Key)
			klog.Infof("Current pod num: %v", len(podMap))
			ok := shed(req.Pod)
			if !ok {
				pendingPod = append(pendingPod, req.Key)
			}
		}
		mtx.Unlock()

	case "DELETE":
		mtx.Lock()
		pod := podMap[req.Key]
		mtx.Unlock()
		podUID := pod.UID
		nodeUID := pod.Spec.NodeName

		cli := http.Client{}
		url := config.AC_ServerAddr + ":" + strconv.Itoa(config.AC_ServerPort)
		buf, _ := json.Marshal(req.Pod)
		http_req, _ := http.NewRequest(http.MethodDelete, url+"/innode/"+nodeUID+"/pod/"+podUID, bytes.NewReader(buf))
		resp, _ := cli.Do(http_req)
		resp.Body.Close()
		klog.Infof("Delete Pod[%v] in Node[%v]", podUID, nodeUID)

		mtx.Lock()
		delete(podMap, req.Key)
		for i := 0; i < len(pendingPod); i++ {
			if pendingPod[i] == req.Key {
				pendingPod = append(pendingPod[:i], pendingPod[i+1:]...)
				break
			}
		}
		mtx.Unlock()
		klog.Infof("Current pod num: %v", len(podMap))
	}
}

func handleNodeChanRequest(req *NodeRequest) {
	switch req.Type {
	case "PUT":
		mtx.Lock()
		if _, exist := nodeMap[req.Key]; exist {
			nodeMap[req.Key] = req.Node
			klog.Infof("Node Changed: Key[%v] Value[...]", req.Key)
			cancelMap[req.Key]()
			newCtx, cancel := context.WithCancel(context.Background())
			go deleteNodeAfter30s(newCtx, req.Key)
			cancelMap[req.Key] = cancel
		} else {
			nodeMap[req.Key] = req.Node
			klog.Infof("New Node Register: Key[%v] Value[...]", req.Key)
			klog.Infof("Current node num: %v", len(nodeMap))
			newCtx, cancel := context.WithCancel(context.Background())
			go deleteNodeAfter30s(newCtx, req.Key)
			cancelMap[req.Key] = cancel
			for _, podKey := range pendingPod {
				shed(podMap[podKey])
			}
		}
		mtx.Unlock()

	case "DELETE":
		mtx.Lock()
		delete(nodeMap, req.Key)
		mtx.Unlock()
		klog.Infof("Node Deregister: Key[%v]", req.Key)
		klog.Infof("Current node num: %v", len(nodeMap))
	}
}

func shed_simple(pod v1.Pod) bool {
	klog.Infof("\nScheduling Pod: UID[%v] NodeName[%v]", pod.UID, pod.Spec.NodeName)
	if pod.Spec.NodeName == "" {
		min := -1
		for _, node := range nodeMap {
			if node.Status.Phase == "Unknown" {
				continue
			}
			resp, _ := http.Get(config.AC_ServerAddr + ":" + strconv.Itoa(config.AC_ServerPort) + "/innode/" + node.UID + "/pods")

			var pods []PodRequest
			buf, _ := io.ReadAll(resp.Body)
			resp.Body.Close()
			_ = json.Unmarshal(buf, &pods)
			if len(pods) < min || min == -1 {
				min = len(pods)
				pod.Spec.NodeName = node.UID
			}
		}
		klog.Infof("Sched handle pod UID[%v]: set appointed Node[%v]", pod.UID, pod.Spec.NodeName)
	}
	if pod.Spec.NodeName == "" {
		klog.Errorf("Sched error: Cannot Assign Pod[%v]: no suitable node", pod.UID)
		return false
	}

	cli := http.Client{}
	url := config.AC_ServerAddr + ":" + strconv.Itoa(config.AC_ServerPort)
	buf, _ := json.Marshal(pod)
	req, _ := http.NewRequest(http.MethodPut, url+"/innode/"+pod.Spec.NodeName+"/pod/"+pod.UID, bytes.NewReader(buf))
	resp2, _ := cli.Do(req)

	if resp2.StatusCode != http.StatusOK {
		klog.Errorf("Sched error: Cannot Assign Pod[%v] to Node[%v]", pod.UID, pod.Spec.NodeName)
		return false
	}
	resp2.Body.Close()

	buf, _ = json.Marshal(pod)
	req, _ = http.NewRequest(http.MethodPut, url+"/pod/"+pod.UID, bytes.NewReader(buf))
	resp3, _ := cli.Do(req)
	resp3.Body.Close()
	klog.Infof("Sched ok with pod UID[%v] to Node UID[%v]", pod.UID, pod.Spec.NodeName)
	return true
}

func shed_rr(pod v1.Pod) bool {
	klog.Infof("\nScheduling Pod: UID[%v] NodeName[%v]", pod.UID, pod.Spec.NodeName)
	if pod.Spec.NodeName == "" && len(nodeMap) > 0 {
		nodeNum := len(nodeMap)
		currNum = (currNum + 1) % nodeNum

		i := 0
		for _, node := range nodeMap {
			if i == currNum {
				pod.Spec.NodeName = node.UID
				break
			}
			i++
		}
		klog.Infof("Sched handle pod UID[%v]: set appointed Node[%v]", pod.UID, pod.Spec.NodeName)
	}
	if pod.Spec.NodeName == "" {
		klog.Errorf("Sched error: Cannot Assign Pod[%v]: no suitable node", pod.UID)
		return false
	}

	cli := http.Client{}
	url := config.AC_ServerAddr + ":" + strconv.Itoa(config.AC_ServerPort)
	buf, _ := json.Marshal(pod)
	req, _ := http.NewRequest(http.MethodPut, url+"/innode/"+pod.Spec.NodeName+"/pod/"+pod.UID, bytes.NewReader(buf))
	resp2, _ := cli.Do(req)
	if resp2.StatusCode != http.StatusOK {
		klog.Errorf("Sched error: Cannot Assign Pod[%v] to Node[%v]", pod.UID, pod.Spec.NodeName)
		return false
	}
	resp2.Body.Close()
	buf, _ = json.Marshal(pod)
	req, _ = http.NewRequest(http.MethodPut, url+"/pod/"+pod.UID, bytes.NewReader(buf))
	resp3, _ := cli.Do(req)
	klog.Infof("Sched ok with pod UID[%v] to Node UID[%v]", pod.UID, pod.Spec.NodeName)
	resp3.Body.Close()
	return true
}

func deleteNodeAfter30s(ctx context.Context, key string) {
	select {
	case <-ctx.Done():
		klog.Infof("Node[%v] updated, reset cancel func", key)
		return
	case <-time.After(30 * time.Second):
		mtx.Lock()
		klog.Infof("Node[%v] disconnected", key)
		node := nodeMap[key]
		node.Status.Phase = "Unknown"
		nodeMap[key] = node
		buf, _ := json.Marshal(node)
		apiclient.Rest(node.UID, string(buf), apiclient.OBJ_NODE, apiclient.OP_PUT)
		mtx.Unlock()
	}
}
