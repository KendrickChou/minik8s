package apiclient

import (
	"context"
	"fmt"
	"minik8s.com/minik8s/config"
	"net/http"
	"strconv"
	"time"
)

func getPodByName(name string) {

}

func TestChunkedWatch() {
	ctx, cl := context.WithCancel(context.Background())
	go watchService(ctx)
	select {
	case <-time.After(time.Second * 5):
		cl()
	}
}

func watchService(ctx context.Context) {
	resp, err := http.Get(config.AC_ServerAddr + ":" + strconv.Itoa(config.AC_ServerPort) + config.AC_WatchServices_Path)
	if err != nil {
		fmt.Printf("error: %v", err)
		return
	}

	buf := make([]byte, 128)
	for {
		select {
		case <-ctx.Done():
			return
		default:
			readN, err := resp.Body.Read(buf)
			if err != nil {
				return
			}
			if readN > 0 {
				println(string(buf))
			}
		}
	}
}
