package main

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"os"
	"strconv"
	"time"

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

	// start heartbeat
	go sendHeartBeat(ctx, kl.NodeName, errChan)

	for {
		select {
		case e := <-errChan:
			klog.Fatalf("Node Failed: %s", e)
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

func sendHeartBeat(ctx context.Context, nodeName string, errChan chan string) {
	counter := 0
	errorCounter := 0
	lastReportTime := time.Now()

	for {
		if errorCounter >= constants.MaxErrorHeartBeat {
			errChan <-"Send heartbeat failed successively for " + strconv.Itoa(constants.MaxErrorHeartBeat) + " times"
			return
		}

		time.Sleep(time.Duration(constants.HeartBeatInterval) * time.Second)

		counter++
		lastReportTime = time.Now()

		klog.Infof("Send Heartbeat %d, time: %s", counter, lastReportTime)

		resp, err := http.Get(constants.HeartBeatRequest + nodeName + "/" + strconv.Itoa(counter))

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
