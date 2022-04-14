package apiserver

import (
	"context"
	clientv3 "go.etcd.io/etcd/client/v3"
	"k8s.io/klog"
	"minik8s.com/minik8s/config"
	"strconv"
	"time"
)

type KV struct {
	key   string
	value string
	ty    string
}

var etcdClient *clientv3.Client
var err error

func initEtcd() {
	etcdClient, err = clientv3.New(clientv3.Config{
		Endpoints:   []string{config.EtcdConfig.EtcdAddr + ":" + strconv.Itoa(config.EtcdConfig.EtcdPort)},
		DialTimeout: 10 * time.Second,
	})
	if err != nil {
		klog.Error("create etcd client failed, err:%v\n", err)
	}
}

func closeEtcd() {
	err = etcdClient.Close()
	if err != nil {
		klog.Error("close etcd client failed, err:%v\n", err)
	}
}

func etcdPut(key, val string) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	_, err = etcdClient.Put(ctx, key, val)
	cancel()
	if err != nil {
		klog.Errorf("etcd put failed, err: %v", err)
		return
	} else {
		klog.Infof("etcd put key: %v, value: %v\n", key, val)
	}
}

func etcdGet(key string) string {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	resp, err := etcdClient.Get(ctx, key)
	cancel()
	if err != nil {
		klog.Errorf("etcd get failed, err: %v", err)
		return ""
	} else {
		val := string(resp.Kvs[0].Value)
		klog.Infof("etcd get key: %v, value: %v\n", key, val)
		return val
	}
}

func etcdWatch(key string) chan KV {
	ch := make(chan KV)
	rch := etcdClient.Watch(context.Background(), key)
	go startWatch(ch, rch)
	klog.Infof("etcd watch start key: %v\n", key)
	return ch
}

func startWatch(ch chan KV, rch clientv3.WatchChan) {
	for resp := range rch {
		for _, ev := range resp.Events {
			klog.Infof("etcd watch emitted type: %s key: %q val: %q\n", ev.Type, ev.Kv.Key, ev.Kv.Value)
			ch <- KV{ty: ev.Type.String(), key: string(ev.Kv.Key), value: string(ev.Kv.Value)}
		}
	}
}
