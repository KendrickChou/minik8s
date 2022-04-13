package apiserver

import (
	"context"
	"fmt"
	clientv3 "go.etcd.io/etcd/client/v3"
	"k8s.io/klog"
	"minik8s.com/minik8s/config"
	"strconv"
	"time"
)

var etcdClient *clientv3.Client
var err error

func initEtcd() {
	etcdClient, err = clientv3.New(clientv3.Config{
		Endpoints:   []string{config.EtcdConfig.EtcdAddr + ":" + strconv.Itoa(config.EtcdConfig.EtcdPort)},
		DialTimeout: 10 * time.Second,
	})
	if err != nil {
		fmt.Printf("connect to etcd failed, err:%v\n", err)
	}
}

func etcdPut(key, val string) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	_, err = etcdClient.Put(ctx, key, val)
	cancel()
	if err != nil {
		klog.Error("etcd put failed, err\n\t", err.Error())
		return
	} else {
		klog.Info("etcd put key: " + key + ", value: " + val + "\n")
	}
}

func etcdGet(key string) string {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	resp, err := etcdClient.Get(ctx, "hello")
	cancel()
	if err != nil {
		klog.Error("etcd get failed, err:\n\t", err.Error())
		return ""
	} else {
		val := string(resp.Kvs[0].Value)
		klog.Info("etcd get key: " + key + ", value: " + val + "\n")
		return val
	}
}

func etcdWatch() {

}
