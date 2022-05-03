package utils

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
}

func NewInformer(kind string) (inf *Informer) {
	inf = &Informer{
		Kind: kind,
	}

	inf.notifyChan = make(chan Delta)
	inf.Handlers = []EventHandler{}
	inf.reflector.Kind = kind
	inf.reflector.NotifyChan = &inf.notifyChan
	return inf
}

func (inf *Informer) Run(stopChan chan bool) {
	reflectorStopChan := make(chan bool)
	syncChan := make(chan bool)
	go inf.reflector.Run(reflectorStopChan, syncChan)

	<-syncChan

	for {
		select {
		case delta := <-inf.notifyChan:
			{
				switch delta.GetType() {
				case "Put":
				case "Post":
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
				case "Delete":
					inf.store.Delete(delta.GetKey())

					for _, handler := range inf.Handlers {
						handler.OnDelete(delta.GetValue())
					}
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
