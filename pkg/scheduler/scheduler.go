package scheduler

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"k8s.io/klog/v2"
	"minik8s.com/minik8s/config"
	v1 "minik8s.com/minik8s/pkg/api/v1"
	"minik8s.com/minik8s/pkg/apiclient"
	"net/http"
	"strconv"
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
var nodeMap map[string]v1.Node

var shed func(pod v1.Pod) bool

func Init() {
	shed = shed_simple
	podMap = make(map[string]v1.Pod)
	nodeMap = make(map[string]v1.Node)
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
		klog.Infof("\nselecting chan...\n")
		select {
		case rawBytes := <-podChan:
			req := &PodRequest{}
			err := json.Unmarshal(rawBytes, req)

			if err != nil {

				klog.Error("Unmarshal Pod Change Req Failed: %v", err)
			} else {
				handlePodChanRequest(req)
			}
		case rawBytes := <-nodeChan:
			req := &NodeRequest{}
			err := json.Unmarshal(rawBytes, req)

			if err != nil {
				klog.Error("Unmarshal Node Change Req Failed: %v", err)
			} else {
				handleNodeChanRequest(req)
			}
		}
	}
}

func handlePodChanRequest(req *PodRequest) {
	switch req.Type {
	case "PUT":
		if _, exist := podMap[req.Key]; exist {
			podMap[req.Key] = req.Pod
			klog.Infof("Pod Changed: Key[%v] Value[...]", req.Key)
			klog.Infof("Current pod num: %v", len(podMap))
		} else {
			podMap[req.Key] = req.Pod
			klog.Infof("New Pod Added: Key[%v] Value[...]", req.Key)
			klog.Infof("Current pod num: %v", len(podMap))
			shed(req.Pod)
		}

	case "DELETE":
		pod := podMap[req.Key]
		podUID := pod.UID
		nodeUID := pod.Spec.NodeName

		cli := http.Client{}
		url := config.AC_ServerAddr + ":" + strconv.Itoa(config.AC_ServerPort)
		buf, _ := json.Marshal(req.Pod)
		http_req, _ := http.NewRequest(http.MethodDelete, url+"/innode/"+nodeUID+"/pod/"+podUID, bytes.NewReader(buf))
		_, _ = cli.Do(http_req)
		klog.Infof("Delete Pod[%v] in Node[%v]", podUID, nodeUID)

		delete(podMap, req.Key)
		klog.Infof("Current pod num: %v", len(podMap))
	}
}

func handleNodeChanRequest(req *NodeRequest) {
	switch req.Type {
	case "PUT":
		nodeMap[req.Key] = req.Node
		klog.Infof("New Node Register: Key[%v] Value[...]", req.Key)
		klog.Infof("Current node num: %v", len(nodeMap))
	case "DELETE":
		delete(nodeMap, req.Key)
		klog.Infof("Node Deregister: Key[%v]", req.Key)
		klog.Infof("Current node num: %v", len(nodeMap))
	}
}

func shed_simple(pod v1.Pod) bool {
	klog.Infof("\nScheduling Pod: UID[%v] NodeName[%v]", pod.UID, pod.Spec.NodeName)
	if pod.Spec.NodeName == "" {
		min := -1
		for _, node := range nodeMap {
			resp, _ := http.Get(config.AC_ServerAddr + ":" + strconv.Itoa(config.AC_ServerPort) + "/innode/" + node.UID + "/pods")
			var pods []PodRequest
			buf, _ := io.ReadAll(resp.Body)
			_ = json.Unmarshal(buf, &pods)
			if len(pods) < min || min == -1 {
				min = len(pods)
				pod.Spec.NodeName = node.UID
			}
		}
		klog.Infof("Sched handle pod UID[%v]: set appointed Node[%v]", pod.UID, pod.Spec.NodeName)
	}
	cli := http.Client{}
	url := config.AC_ServerAddr + ":" + strconv.Itoa(config.AC_ServerPort)
	buf, _ := json.Marshal(pod)
	req, _ := http.NewRequest(http.MethodPut, url+"/innode/"+pod.Spec.NodeName+"/pod/"+pod.UID, bytes.NewReader(buf))
	resp, _ := cli.Do(req)
	if resp.StatusCode != http.StatusOK {
		klog.Errorf("Sched error: Cannot Assign Pod[%v] to Node[%v]", pod.UID, pod.Spec.NodeName)
		return false
	}

	buf, _ = json.Marshal(pod)
	req, _ = http.NewRequest(http.MethodPut, url+"/pod/"+pod.UID, bytes.NewReader(buf))
	resp, _ = cli.Do(req)
	klog.Infof("Sched ok with pod UID[%v] to Node UID[%v]", pod.UID, pod.Spec.NodeName)
	return resp.StatusCode == http.StatusOK
}
