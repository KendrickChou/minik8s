package main

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"os"

	"k8s.io/klog/v2"
	"minik8s.com/minik8s/pkg/kubelet"
	"minik8s.com/minik8s/pkg/kubelet/apis/constants"
	"minik8s.com/minik8s/pkg/kubelet/apis/podchangerequest"
)

const JsonContentType string = "application/json"

func main() {
	kl := kubelet.NewKubelet("firstnode")

	// regist to apiserver

	registBody, _ := json.Marshal(kl)
	resp, _ := http.Post(constants.ApiServerAddress+"/node", JsonContentType, bytes.NewBuffer(registBody))

	if resp.StatusCode != 200 {
		klog.Errorf("Node %s failed regist to apiserver %s", kl.NodeName, constants.ApiServerAddress)
		resp.Body.Close()
		os.Exit(0)
	}

	resp.Body.Close()

	// watch pod
	ctx, cancel := context.WithCancel(context.TODO())
	defer cancel()

	podChangeRaw := make(chan []byte)
	errChan := make(chan string)
	go watching(ctx, kl.NodeName, podChangeRaw, errChan)

	for {
		select {
		case <-errChan:
			return
		case rawBytes := <-podChangeRaw:
			req := &podchangerequest.PodChangeRequest{}
			err := json.Unmarshal(rawBytes, req)

			if err != nil {
				klog.Error("Unmarshal APIServer Data Failed: %s", err.Error())
			} else {
				handlePodChangeRequest(&kl, req)
			}
		}
	}

}

func watching(ctx context.Context, nodeName string, podChange chan []byte, errChan chan string) {
	resp, err := http.Get(constants.WatchPodsRequest + nodeName)

	if err != nil {
		klog.Errorf("Node %s Watch Pods Failed: %s", nodeName, err.Error())
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

			podChange <- buf

		}
	}
}

func handlePodChangeRequest(kl *kubelet.Kubelet, req *podchangerequest.PodChangeRequest) {
	switch req.Type {
	case "PUT":
		kl.CreatePod(req.Pod)
	case "DELETE":
		kl.DeletePod(req.Pod.UID)
	default:
		klog.Errorln("Unknown Pod Change Request Type: %s", req.Type)
		return
	}
}
