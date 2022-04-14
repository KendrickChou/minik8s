package apiserver

import (
	"fmt"
)

type ApiServer struct {
}

func CreateNewApiServer() *ApiServer {
	var as = new(ApiServer)
	return as
}

func TestEtcd() {
	initEtcd()
	defer closeEtcd()
	wch := etcdWatch("hello")
	go handleWatchResult(wch)
	etcdPut("hello", "world1")
	etcdPut("hello", "world2")
	etcdPut("hello", "world3")
	etcdGet("hello")
}

func handleWatchResult(wch chan KV) {
	for kv := range wch {
		fmt.Printf("user watch key: %v, val: %v\n", kv.key, kv.value)
	}
}

func RegisterPodWatcher() {

}
