package main

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"k8s.io/klog/v2"
	"minik8s.com/minik8s/pkg/kubelet"
	"minik8s.com/minik8s/pkg/kubelet/apis/config"
	"minik8s.com/minik8s/pkg/kubelet/apis/constants"
	"minik8s.com/minik8s/pkg/kubelet/apis/httpresponse"
	kubeproxy "minik8s.com/minik8s/pkg/kubelet/kube-proxy"
)

const JsonContentType string = "application/json"

func main() {

	resp, err := http.Post(config.ApiServerAddress+constants.RegistNodeRequest(), JsonContentType, bytes.NewBuffer([]byte{}))

	// regist to apiserver
	if err != nil || resp.StatusCode != 200 {
		klog.Fatalf("Node failed register to apiserver %s", config.ApiServerAddress)
		resp.Body.Close()
		os.Exit(0)
	}

	registResp := &httpresponse.RegistResponse{}

	buf, _ := io.ReadAll(resp.Body)
	err = json.Unmarshal(buf, registResp)

	resp.Body.Close()

	if err != nil {
		klog.Fatal("Json parse RegistResponse failed")
		os.Exit(0)
	}

	// watch node to get CIDR and block till get response.
	// resp, err = http.Get(config.ApiServerAddress + constants.WatchNodeRequest + registResp.UID)

	// if err != nil {
	// 	klog.Fatalf("Node failed get CIDR from apiserver: %s", err.Error())
	// 	os.Exit(0)
	// }

	// watchNodeResp := httpresponse.WatchNodeResponse{}
	// buf, _ = io.ReadAll(resp.Body)
	// err = json.Unmarshal(buf, watchNodeResp)

	// resp.Body.Close()

	// if err != nil {
	// 	klog.Fatal("Json parse WatchNodeResponse failed")
	// 	os.Exit(0)
	// }

	// if watchNodeResp.Key != registResp.UID {
	// 	klog.Fatalf("Actual Node Name: %s, Received Node Name: %s", registResp.UID, watchNodeResp.Key)
	// 	os.Exit(0)
	// }

	// create kubelet
	kl, err := kubelet.NewKubelet(config.NodeName, registResp.UID)

	if err != nil {
		klog.Fatalf("Create Kubelet Failed: %s", err.Error())
		os.Exit(0)
	}

	// create kube proxy
	kp, err := kubeproxy.NewKubeProxy()

	if err != nil {
		klog.Fatalf("Create Kube-Proxy Failed: %s", err.Error())
		os.Exit(0)
	}

	ctx, cancel := context.WithCancel(context.TODO())
	defer cancel()

	// may some parallel bugs. maybe should not share kl

	// watch pod
	errChan := make(chan string)
	go watchingPods(ctx, &kl, errChan)

	// start heartbeat
	go sendHeartBeat(ctx, kl.UID, errChan)

	// start refreshPodStatus periodically
	go refreshAllPodStatus(ctx, &kl)

	// watch endpoints
	go watchingEndpoints(ctx, kp, errChan)

	for {
		select {
		case e := <-errChan:
			klog.Fatalf("Node Failed: %s", e)
			return
		}
	}

}

func watchingPods(ctx context.Context, kl *kubelet.Kubelet, errChan chan string) {
	resp, err := http.Get(config.ApiServerAddress + constants.WatchPodsRequest(kl.UID))

	if err != nil {
		klog.Errorf("Node %s Watch Pods Failed: %s", kl.UID, err.Error())
		errChan <- err.Error()
		return
	}

	for {
		select {
		case <-ctx.Done():
			return
		default:
			reader := bufio.NewReader(resp.Body)
			buf, err := reader.ReadBytes(byte(constants.EOF))

			if err != nil {
				klog.Errorf("Watch Pods Error: %s", err)
				errChan <- err.Error()
				return
			}

			req := &httpresponse.PodChangeRequest{}
			err = json.Unmarshal(buf, req)

			if err != nil {
				klog.Error("Unmarshal APIServer Data Failed: %s", err.Error())
			} else {
				handlePodChangeRequest(kl, req)
			}

			podChange <- buf
		}
	}
}

func handlePodChangeRequest(kl *kubelet.Kubelet, req *httpresponse.PodChangeRequest) {
	switch req.Type {
	case "PUT":
		parsedPath := strings.Split(req.Key, "/")
		req.Pod.UID = parsedPath[len(parsedPath) - 1]
		
		kl.CreatePod(req.Pod)
	case "DELETE":
		kl.DeletePod(req.Pod.UID)
	default:
		klog.Errorln("Unknown Pod Change Request Type: %s", req.Type)
		return
	}
}

func sendHeartBeat(ctx context.Context, nodeUID string, errChan chan string) {
	counter := 0
	errorCounter := 0
	// lastReportTime := time.Now()

	for {
		if errorCounter >= constants.MaxErrorHeartBeat {
			errChan <- "Send heartbeat failed successively for " + strconv.Itoa(constants.MaxErrorHeartBeat) + " times"
			return
		}

		time.Sleep(time.Duration(constants.HeartBeatInterval) * time.Second)

		counter++
		// lastReportTime = time.Now()

		// klog.Infof("Send Heartbeat %d, time: %s", counter, lastReportTime)

		resp, err := http.Get(config.ApiServerAddress + constants.HeartBeatRequest(nodeUID, strconv.Itoa(counter)))

		if err != nil {
			klog.Warningf("Send Heartbeat %d Failed: %s", counter, err.Error())
			errorCounter++
			continue
		}

		if resp.StatusCode != 200 {
			klog.Warningf("Send Heartbeat %d Failed, response status %s", counter, resp.Status)
			errorCounter++
			continue
		}

		errorCounter = 0
	}
}

func refreshAllPodStatus(ctx context.Context, kl *kubelet.Kubelet) {

	for {
		pods, err := kl.GetPods()

		if err != nil {
			klog.Errorf("Error When refresh Pod Status: %s", err.Error())
		}

		for _, pod := range pods {
			body, err := json.Marshal(pod.Status)

			if err != nil {
				klog.Errorf("Marshal pod status error: %s", err.Error())
			}

			resp, err := http.Post(config.ApiServerAddress+constants.RefreshPodRequest(kl.UID, pod.UID), JsonContentType, bytes.NewReader(body))

			if err != nil {
				klog.Errorf("Error When refresh Pod Status: %s", err.Error())
			}

			resp.Body.Close()
		}

		time.Sleep(time.Duration(constants.RefreshPodStatusInterval) * time.Second)
	}
}

func watchingEndpoints(ctx context.Context, kp kubeproxy.KubeProxy, errChan chan string) {
	resp, err := http.Get(config.ApiServerAddress + constants.WatchEndpointsRequest())

	if err != nil {
		klog.Errorf("Node %s Watch Endpoints Failed: %s", err.Error())
		errChan <- err.Error()
		return
	}

	for {
		select {
		case <-ctx.Done():
			return
		default:
			reader := bufio.NewReader(resp.Body)
			buf, err := reader.ReadBytes(byte(constants.EOF))

			if err != nil {
				klog.Errorf("Watch Endpoints Error: %s", err)
				errChan <- err.Error()
				return
			}

			req := &httpresponse.EndpointChangeRequest{}
			err = json.Unmarshal(buf, req)

			if err != nil {
				klog.Error("Unmarshal APIServer Data Failed: %s", err.Error())
			} else {
				handleEndpointChangeRequest(kp, req)
			}
		}
	}
}

func handleEndpointChangeRequest(kp kubeproxy.KubeProxy, req *httpresponse.EndpointChangeRequest) {
	switch req.Type {
	case "PUT":
		kp.AddEndpoint(context.TODO(), req.Endpoint)
	case "DELETE":
		kp.RemoveEndpoint(context.TODO(), req.Endpoint.Name)
	default:
		klog.Errorln("Unknown Pod Change Request Type: %s", req.Type)
		return
	}
}
