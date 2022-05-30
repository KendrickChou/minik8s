package apiserver

import (
	"context"
	"encoding/json"
	clientv3 "go.etcd.io/etcd/client/v3"
	"k8s.io/klog"
	"minik8s.com/minik8s/config"
	"strconv"
	"time"
)

type KV struct {
	Key   string          `json:"key"`
	Value json.RawMessage `json:"value"`
	Type  string          `json:"type"`
}

var etcdClient *clientv3.Client

func initEtcd() {
	var err error
	etcdClient, err = clientv3.New(clientv3.Config{
		Endpoints:   []string{config.AS_EtcdAddr + ":" + strconv.Itoa(config.AS_EtcdPort)},
		DialTimeout: 10 * time.Second,
	})
	if err != nil {
		klog.Errorf("create etcd client failed, err:%v\n", err)
	} else {
		klog.Info("successfully started etcd client\n\n")
	}
}

func closeEtcd() {
	err := etcdClient.Close()
	if err != nil {
		klog.Errorf("close etcd client failed, err:%v\n", err)
	} else {
		klog.Info("etcd client closed\n")
	}
}

func etcdPut(key, val string) error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	_, err := etcdClient.Put(ctx, key, val)
	cancel()
	if err != nil {
		klog.Errorf("etcd put failed, err: %v", err)
	} else {
		klog.Infof("etcd put key: %v, value: %v\n", key, val)
	}
	return err
}

func etcdGet(key string) (KV, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	resp, err := etcdClient.Get(ctx, key)
	cancel()
	if err != nil || resp.Count == 0 {
		klog.Errorf("etcd get failed, err: %v", err)
		return KV{
			Key:   "",
			Value: []byte{},
			Type:  config.AS_OP_ERROR_String,
		}, err
	} else {
		val := resp.Kvs[0].Value
		klog.Infof("etcd get key: %v, value: %s\n", key, val)
		return KV{
			Key:   key,
			Value: val,
			Type:  config.AS_OP_GET_String,
		}, err
	}
}

func etcdTest(key string) bool {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	resp, err := etcdClient.Get(ctx, key, clientv3.WithCountOnly())
	cancel()
	if err != nil || resp.Count == 0 {
		klog.Errorf("etcd test failed, key: %v err: %v", key, err)
		return false
	} else {
		klog.Infof("etcd test key: %v\n", key)
		return true
	}
}

func etcdGetPrefix(key string) ([]KV, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	resp, err := etcdClient.Get(ctx, key, clientv3.WithPrefix())
	cancel()
	if err != nil {
		klog.Errorf("etcd get failed, err: %v", err)
		return []KV{}, err
	} else {
		var kvList []KV
		for _, kv := range resp.Kvs {
			kvList = append(kvList, KV{string(kv.Key), kv.Value, config.AS_OP_GET_String})
			klog.Infof("etcd get with prefix: %s, key: %s, value: %s\n", key, kv.Key, kv.Value)
		}
		return kvList, err
	}
}

func etcdTestPrefix(key string) bool {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	resp, err := etcdClient.Get(ctx, key, clientv3.WithCountOnly(), clientv3.WithPrefix())
	cancel()
	if err != nil || resp.Count == 0 {
		klog.Errorf("etcd test failed, err: %v", err)
		return false
	} else {
		klog.Infof("etcd test key: %v\n", key)
		return true
	}
}

func etcdDel(key string) error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	_, err := etcdClient.Delete(ctx, key)
	cancel()
	if err != nil {
		klog.Errorf("etcd delete failed, err: %v", err)
	} else {
		klog.Infof("etcd delete key: %v\n", key)
	}
	return err
}

func etcdDelPrefix(key string) error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	_, err := etcdClient.Delete(ctx, key, clientv3.WithPrefix())
	cancel()
	if err != nil {
		klog.Errorf("etcd delete failed, err: %v", err)
	} else {
		klog.Infof("etcd delete prefix: %v\n", key)
	}
	return err
}

func etcdWatch(key string) (chan *KV, context.CancelFunc) {
	ch := make(chan *KV)
	ctx, cancel := context.WithCancel(context.Background())
	rch := etcdClient.Watch(ctx, key)
	go startWatch(ch, rch)
	klog.Infof("etcd watch start key: %v\n", key)
	return ch, cancel
}

func etcdWatchPrefix(key string) (chan *KV, context.CancelFunc) {
	ch := make(chan *KV)
	ctx, cancel := context.WithCancel(context.Background())
	rch := etcdClient.Watch(ctx, key, clientv3.WithPrefix())
	go startWatch(ch, rch)
	klog.Infof("etcd watch start prefix: %v\n", key)
	return ch, cancel
}

func startWatch(ch chan *KV, rch clientv3.WatchChan) {
	for resp := range rch {
		for _, ev := range resp.Events {
			klog.Infof("etcd watch emitted -- type: %s key: %s val: %s\n", ev.Type, ev.Kv.Key, ev.Kv.Value)
			ch <- &KV{Type: ev.Type.String(), Key: string(ev.Kv.Key), Value: json.RawMessage(ev.Kv.Value)}
		}
	}
}
