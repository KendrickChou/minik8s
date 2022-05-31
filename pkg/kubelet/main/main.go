package main

import (
	"bufio"
	"bytes"
	"context"
	"encoding/gob"
	"encoding/json"
	"errors"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/patrickmn/go-cache"
	"k8s.io/klog/v2"
	"minik8s.com/minik8s/pkg/kubelet"
	"minik8s.com/minik8s/pkg/kubelet/apis/config"
	"minik8s.com/minik8s/pkg/kubelet/apis/constants"
	"minik8s.com/minik8s/pkg/kubelet/apis/httpresponse"
	kubeproxy "minik8s.com/minik8s/pkg/kubelet/kube-proxy"
)

const JsonContentType string = "application/json"

func main() {

	// read from cache if exist
	_, err := os.Stat(constants.CacheFilePath)

	var nodeUID string
	var restart bool = false

	// get node info in cache or send request to api server
	if err == nil {
		klog.Info("Read node info from cache.")

		f, err := os.Open(constants.CacheFilePath)
		if err != nil {
			klog.Errorf("Open Cache File %s Error: %s", constants.CacheFilePath, err.Error())
			os.Exit(0)
		}

		dec := gob.NewDecoder(f)
		var items map[string]cache.Item
		err = dec.Decode(&items)

		f.Close()

		if err != nil {
			klog.Errorf("Decode cache file Error: %s", err.Error())
			os.Exit(0)
		}

		nodeCache := cache.NewFrom(cache.NoExpiration, 0, items)
		value, ok := nodeCache.Get(constants.NodeCacheID)

		if !ok {
			klog.Errorf("No node cache in cache file")
			os.Exit(0)
		}

		nodeUID = value.(string)
		restart = true
	} else if errors.Is(err, os.ErrNotExist) {
		klog.Info("Send request to register a node.")

		resp, err := http.Post(config.ApiServerAddress+constants.RegistNodeRequest(), JsonContentType, bytes.NewBuffer([]byte{}))

		// regist to apiserver
		if err != nil || resp.StatusCode != 200 {
			klog.Fatalf("Node failed register to apiserver %s", config.ApiServerAddress)
			if resp != nil {
				resp.Body.Close()
			}
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

		nodeUID = registResp.UID

		// write it to cache & store to cache file
		nodeCache := cache.New(cache.NoExpiration, 0)

		nodeCache.Add(constants.NodeCacheID, nodeUID, cache.NoExpiration)

		items := nodeCache.Items()
		f, err := os.Create(constants.CacheFilePath)

		if err != nil {
			klog.Errorf("Create Cache File %s Error: %s",  constants.CacheFilePath,err.Error())
			os.Exit(0)
		}

		enc := gob.NewEncoder(f)
		err = enc.Encode(items)
		f.Close()

		if err != nil {
			klog.Error("Encode cache items %s Error: %s", err.Error())
			os.Exit(0)
		}

		klog.Info("Register node & cache finished")
	} else {
		klog.Errorf("Check File %s Error: %s", constants.CacheFilePath, err.Error())
	}

	// create kubelet
	kl, err := kubelet.NewKubelet(config.NodeName, nodeUID)

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

	// get the already exist pod when restart.
	if restart {
		klog.Info("Try to recover pods...")

		resp, err := http.Get(config.ApiServerAddress + constants.GetAllPodsRequest(nodeUID))

		if err != nil {
			klog.Errorf("Get All Existed Pods Error: %s", err.Error())
			os.Exit(0)
		}

		buf, _ := ioutil.ReadAll(resp.Body)
		resp.Body.Close()

		var podArray []httpresponse.PodChangeRequest
		err = json.Unmarshal(buf, &podArray)

		if err != nil {
			klog.Errorf("Unmarshal All Pods Error: %s", err.Error())
			os.Exit(0)
		}

		for _, pod := range podArray {
			kl.AddPodsWithoutCreate(pod.Pod)
		}

		klog.Info("Recover pods complete.")
	}

	ctx, cancel := context.WithCancel(context.TODO())
	defer cancel()

	// may some parallel bugs. maybe should not share kl

	// watch pod
	klog.Info("Begin watching pods")
	podErr := make(chan string)
	go watchingPods(ctx, &kl, podErr)

	// start heartbeat
	klog.Info("Begin send heartbeat")
	heartbeatErr := make(chan string)
	go sendHeartBeat(ctx, kl.UID, heartbeatErr)

	// start refreshPodStatus periodically
	klog.Info("Begin refresh pod status")
	go refreshAllPodStatus(ctx, &kl)

	// watch endpoints
	klog.Info("Begin watching endpoints")
	endpointsErr := make(chan string)
	go watchingEndpoints(ctx, kp, endpointsErr)

	klog.Info("Kubelet is ready to go.")

	for {
		select {
		case e := <-podErr:
			klog.Errorf("Node seems Failed: %s", e)

			go watchingPods(ctx, &kl, podErr)
		case e := <-heartbeatErr:
			klog.Errorf("Node seems Failed: %s", e)

			go sendHeartBeat(ctx, kl.UID, heartbeatErr)
		case e := <-endpointsErr:
			klog.Errorf("Node seems Failed: %s", e)

			go watchingEndpoints(ctx, kp, endpointsErr)
		}
	}

}

func watchingPods(ctx context.Context, kl *kubelet.Kubelet, errChan chan string) {
	resp, err := http.Get(config.ApiServerAddress + constants.WatchPodsRequest(kl.UID))

	if err != nil {
		klog.Errorf("Node %s Watch Pods Failed: %s", kl.UID, err.Error())
		// sleep some time before retry
		time.Sleep(time.Second * time.Duration(constants.ReconnectInterval))
		errChan <- err.Error()
		return
	}

	reader := bufio.NewReader(resp.Body)
	for {
		select {
		case <-ctx.Done():
			return
		default:
			buf, err := reader.ReadBytes(byte(constants.EOF))

			if err != nil {
				klog.Errorf("Watch Pods Error: %s", err)
				errChan <- err.Error()
				return
			}

			buf[len(buf)-1] = '\n'
			req := &httpresponse.PodChangeRequest{}
			err = json.Unmarshal(buf, req)

			if err != nil {
				klog.Errorf("Unmarshal APIServer Data Failed: %s", err.Error())
			} else {
				handlePodChangeRequest(kl, req)
			}
		}
	}
}

func handlePodChangeRequest(kl *kubelet.Kubelet, req *httpresponse.PodChangeRequest) {
	parsedPath := strings.Split(req.Key, "/")
	req.Pod.UID = parsedPath[len(parsedPath)-1]

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

		resp.Body.Close()

		if resp.StatusCode != 200 {
			klog.Warningf("Send Heartbeat %d Failed, response status %s", counter, resp.Status)
			errorCounter++
			continue
		}

		errorCounter = 0
	}
}

func refreshAllPodStatus(ctx context.Context, kl *kubelet.Kubelet) {

	cli := http.Client{}

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

			req, _ := http.NewRequest(http.MethodPut, config.ApiServerAddress+constants.RefreshPodRequest(kl.UID, pod.UID), bytes.NewReader(body))
			resp, err := cli.Do(req)

			if err != nil {
				klog.Errorf("Error When refresh Pod Status: %s", err.Error())
				continue
			}

			resp.Body.Close()
		}

		time.Sleep(time.Duration(constants.RefreshPodStatusInterval) * time.Second)
	}
}

func watchingEndpoints(ctx context.Context, kp kubeproxy.KubeProxy, errChan chan string) {
	resp, err := http.Get(config.ApiServerAddress + constants.WatchEndpointsRequest())

	if err != nil {
		klog.Errorf("Node Watch Endpoints Failed: %s", err.Error())
		// sleep some time before retry
		time.Sleep(time.Second * time.Duration(constants.ReconnectInterval))
		errChan <- err.Error()
		return
	}

	reader := bufio.NewReader(resp.Body)
	for {
		select {
		case <-ctx.Done():
			return
		default:
			buf, err := reader.ReadBytes(byte(constants.EOF))

			if err != nil {
				klog.Errorf("Watch Endpoints Error: %s", err)
				errChan <- err.Error()
				return
			}

			buf[len(buf)-1] = '\n'
			req := &httpresponse.EndpointChangeRequest{}
			err = json.Unmarshal(buf, req)

			if err != nil {
				klog.Errorf("Unmarshal APIServer Data Failed: %s", err.Error())
			} else {
				handleEndpointChangeRequest(kp, req)
			}
		}
	}
}

func handleEndpointChangeRequest(kp kubeproxy.KubeProxy, req *httpresponse.EndpointChangeRequest) {
	klog.Infof("Receive %s Endpoint %s, Key %s", req.Type, req.Endpoint.Name, req.Key)
	parsedPath := strings.Split(req.Key, "/")
	uid := parsedPath[len(parsedPath)-1]
	switch req.Type {
	case "PUT":
		kp.AddEndpoint(context.TODO(), uid, req.Endpoint)
	case "DELETE":
		kp.RemoveEndpoint(context.TODO(), uid)
	default:
		klog.Errorln("Unknown Pod Change Request Type: %s", req.Type)
		return
	}
}
