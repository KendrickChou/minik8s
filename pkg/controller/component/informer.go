package component

import (
	"k8s.io/klog"
	v1 "minik8s.com/minik8s/pkg/api/v1"
	"minik8s.com/minik8s/pkg/apiclient"
)

type EventHandler struct {
	OnAdd    func(obj any)
	OnUpdate func(newObj, oldObj any)
	OnDelete func(obj any)
}

type Informer struct {
	Kind       string
	Handlers   []EventHandler
	store      ThreadSafeStore
	reflector  Reflector
	notifyChan chan Delta
	synced     bool
}

func NewInformer(kind string) (inf *Informer) {
	inf = &Informer{
		Kind:   kind,
		synced: false,
	}

	inf.notifyChan = make(chan Delta)
	inf.Handlers = []EventHandler{}
	inf.reflector.Kind = kind
	inf.reflector.NotifyChan = inf.notifyChan
	inf.store.Init()
	return inf
}

func (inf *Informer) Run(stopChan chan bool) {
	reflectorStopChan := make(chan bool)
	syncChan := make(chan bool)
	go inf.reflector.Run(reflectorStopChan, syncChan)

	for {
		var loop = true
		select {
		case delta := <-inf.notifyChan:
			inf.store.Add(delta.GetKey(), delta.GetValue())
		case <-syncChan:
			loop = false
		}

		if !loop {
			break
		}
	}

	inf.synced = true

	for {
		select {
		case delta := <-inf.notifyChan:
			{
				switch delta.GetType() {
				case "PUT", "POST":
					oldObj, exist := inf.store.Get(delta.GetKey())
					inf.store.Add(delta.GetKey(), delta.GetValue())

					if exist {
						for _, handler := range inf.Handlers {
							handler.OnUpdate(delta.GetValue(), oldObj)
						}
					} else {
						for _, handler := range inf.Handlers {
							handler.OnAdd(delta.GetValue())
						}
					}
				case "DELETE":
					obj, exist := inf.store.Get(delta.GetKey())

					if exist {
						for _, handler := range inf.Handlers {
							handler.OnDelete(obj)
						}

						inf.store.Delete(delta.GetKey())
					}
				default:
					klog.Error("invalid delta type\n")
				}
			}
		case <-stopChan:
			reflectorStopChan <- true
			return
		}

	}
}

func (inf *Informer) AddEventHandler(handler EventHandler) {
	inf.Handlers = append(inf.Handlers, handler)
}

func (inf *Informer) HasSynced() bool {
	return inf.synced
}

func (inf *Informer) List() []any {
	return inf.store.List()
}

func (inf *Informer) GetItem(key string) any {
	item, exist := inf.store.Get(key)
	if !exist {
		return nil
	} else {
		return item
	}
}

func (inf *Informer) DeleteItem(key string) {
	var flag bool
	switch inf.Kind {
	case "Endpoint":
		{
			flag = apiclient.DeleteEndpoint(key)
		}
	case "Pod":
		{
			flag = apiclient.DeletePod(key)
		}
	default:
		klog.Warningf("Delete %s not handled", inf.Kind)
	}

	if flag {
		inf.store.Delete(key)
	} else {
		klog.Errorf("Delete %s failed", key)
	}
}

func (inf *Informer) UpdateItem(key string, obj any) {
	var flag bool
	switch inf.Kind {
	case "Pod":
		{
			pod := obj.(v1.Pod)
			flag = apiclient.UpdatePod(&pod)
		}
	case "Endpoint":
		{
			ep := obj.(v1.Endpoint)
			flag = apiclient.UpdateEndpoint(&ep)
		}
	case "ReplicaSet":
		{
			rs := obj.(v1.ReplicaSet)
			flag = apiclient.UpdateReplicaSet(&rs)
		}
	case "HorizontalPodAutoscaler":
		{
			hpa := obj.(v1.HorizontalPodAutoscaler)
			flag = apiclient.UpdateHorizontalPodAutoscaler(&hpa)
		}
	default:
		klog.Warningf("Update %s not handled", inf.Kind)
	}

	if flag {
		inf.store.Update(key, obj)
	} else {
		klog.Errorf("Update %s failed", key)
	}
}

func (inf *Informer) AddItem(obj any) {
	var uid string
	switch inf.Kind {
	case "Pod":
		{
			pod := obj.(v1.Pod)
			uid = apiclient.PostPod(&pod)
		}
	case "Endpoint":
		{
			ep := obj.(v1.Endpoint)
			uid = apiclient.PostEndpoint(&ep)
		}
	default:
		klog.Warningf("Add %s not handled", inf.Kind)
	}

	if uid != "" {
		inf.store.Add(uid, obj)
	} else {
		klog.Error("Add Object failed ", obj)
	}
}
