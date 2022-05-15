package component

import (
	"context"
	"encoding/json"
	"k8s.io/klog"
	"minik8s.com/minik8s/pkg/apiclient"
)

type Reflector struct {
	// object type the reflector list/watch
	Kind       string
	NotifyChan chan Delta
}

// Run list and watch
func (r *Reflector) Run(stopChan chan bool, syncChan chan bool) {
	r.list()
	syncChan <- true
	r.watch(stopChan)
}

func (r *Reflector) list() {
	var objType apiclient.ObjType

	switch r.Kind {
	case "Pod":
		objType = apiclient.OBJ_ALL_PODS
	case "Service":
		objType = apiclient.OBJ_ALL_SERVICES
	case "ReplicaSet":
		objType = apiclient.OBJ_ALL_REPLICAS
	case "Endpoint":
		objType = apiclient.OBJ_ALL_ENDPOINTS
	}

	objects := apiclient.GetAll(objType)
	switch r.Kind {
	case "Pod":
		var fmtObjs []PodObject

		err := json.Unmarshal(objects, &fmtObjs)
		if err != nil {
			klog.Error("Reflector parse error\n")
		}

		for _, podObj := range fmtObjs {
			podObj.Type = "PUT"
			podObj.StripKey()
			r.NotifyChan <- &podObj
		}
	case "ReplicaSet":
		var fmtObjs []ReplicaSetObject

		err := json.Unmarshal(objects, &fmtObjs)
		if err != nil {
			klog.Error("Reflector parse error\n")
		}

		for _, rsObj := range fmtObjs {
			rsObj.Type = "PUT"
			rsObj.StripKey()
			r.NotifyChan <- &rsObj
		}
	case "Service":
		var fmtObjs []ServiceObject

		err := json.Unmarshal(objects, &fmtObjs)
		if err != nil {
			klog.Error("Reflector parse error\n")
		}

		for _, serviceObj := range fmtObjs {
			serviceObj.Type = "PUT"
			serviceObj.StripKey()
			r.NotifyChan <- &serviceObj
		}
	case "Endpoint":
		var fmtObjs []EndpointObject

		err := json.Unmarshal(objects, &fmtObjs)
		if err != nil {
			klog.Error("Reflector parse error\n")
		}

		for _, epObj := range fmtObjs {
			epObj.Type = "PUT"
			epObj.StripKey()
			r.NotifyChan <- &epObj
		}
	}
}

func (r *Reflector) watch(stopChan chan bool) {
	var objType apiclient.ObjType

	switch r.Kind {
	case "Pod":
		objType = apiclient.OBJ_ALL_PODS
	case "Service":
		objType = apiclient.OBJ_ALL_SERVICES
	case "ReplicaSet":
		objType = apiclient.OBJ_ALL_REPLICAS
	case "Endpoint":
		objType = apiclient.OBJ_ALL_ENDPOINTS
	}

	ctx, cl := context.WithCancel(context.Background())
	watchChan := make(chan []byte)
	go apiclient.Watch(ctx, watchChan, objType)
	for {
		select {
		case <-stopChan:
			cl()
			return
		case bytes := <-watchChan:
			r.parseJsonAndNotify(bytes)
		}
	}
}

func (r *Reflector) parseJsonAndNotify(jsonObj []byte) {
	switch r.Kind {
	case "Pod":
		obj := &PodObject{}
		err := json.Unmarshal(jsonObj, obj)
		if err != nil {
			klog.Error("Reflector parse error\n")
		}

		obj.StripKey()
		r.NotifyChan <- obj
	case "ReplicaSet":
		obj := &ReplicaSetObject{}
		err := json.Unmarshal(jsonObj, obj)
		if err != nil {
			klog.Error("Reflector parse error\n")
		}

		obj.StripKey()
		r.NotifyChan <- obj
	case "Service":
		obj := &ServiceObject{}
		err := json.Unmarshal(jsonObj, obj)
		if err != nil {
			klog.Error("Reflector parse error\n")
		}

		obj.StripKey()
		r.NotifyChan <- obj
	case "Endpoint":
		obj := &EndpointObject{}
		err := json.Unmarshal(jsonObj, obj)
		if err != nil {
			klog.Error("Reflector parse error\n")
		}

		obj.StripKey()
		r.NotifyChan <- obj
	}
}
