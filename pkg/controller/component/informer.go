package component

import (
	"k8s.io/klog"
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
	/*
		TODO:
		   use queue to process the events, channel will congest, in such case the stored
		 objects may not be the newest
	*/
	for {
		select {
		case delta := <-inf.notifyChan:
			{
				switch delta.GetType() {
				case "PUT":
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
					inf.store.Delete(delta.GetKey())

					for _, handler := range inf.Handlers {
						handler.OnDelete(delta.GetValue())
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
