package apiclient

import (
	"context"
	"fmt"
	"k8s.io/klog"
	"minik8s.com/minik8s/config"
	"net/http"
	"strconv"
	"time"
)

func getPodByName(name string) {

}

func TestChunkedWatch() {
	ctx, cl := context.WithCancel(context.Background())
	watchChan := make(chan string)
	go watchPod(ctx, watchChan)
	select {
	case <-time.After(time.Second * 5):
		cl()
	}
}

func watchPod(ctx context.Context, ch chan string) {
	resp, err := http.Get(config.AC_ServerAddr + ":" + strconv.Itoa(config.AC_ServerPort) + config.AC_WatchServices_Path)
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
