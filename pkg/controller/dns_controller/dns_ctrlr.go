package dns_controller

import (
	"context"
	"encoding/json"
	"k8s.io/klog/v2"
	v1 "minik8s.com/minik8s/pkg/api/v1"
	"minik8s.com/minik8s/pkg/apiclient"
	"os/exec"
)

type DNSRequest struct {
	Key string `json:"key"`

	DNS v1.DNS `json:"value"`

	Type string `json:"type"`
}
type ServiceRequest struct {
	Key string `json:"key"`

	Service v1.Service `json:"value"`

	Type string `json:"type"`
}

var dnsMap map[string]v1.DNS

func Init() {
	dnsMap = make(map[string]v1.DNS)
}

func Run() {
	ctx, cancle := context.WithCancel(context.Background())
	defer cancle()

	dnsChan := make(chan []byte)
	//watch pod and node
	go apiclient.Watch(ctx, dnsChan, apiclient.OBJ_ALL_DNSS)

	//add existed dnss
	dnss_raw := apiclient.Rest("", "", apiclient.OBJ_ALL_DNSS, apiclient.OP_GET)
	var dnss []DNSRequest
	err := json.Unmarshal(dnss_raw, &dnss)
	if err != nil {
		klog.Error("Unmarshal DNSs Failed: %v", err)
	} else {
		for _, dnsReq := range dnss {
			dnsMap[dnsReq.Key] = dnsReq.DNS
			installDNS(dnsReq.DNS)
		}
		klog.Infof("Current dns num: %v", len(dnsMap))
	}

	//handle watch results
	for {
		select {
		case rawBytes := <-dnsChan:
			req := &DNSRequest{}
			err := json.Unmarshal(rawBytes, req)

			if err != nil {
				klog.Error("Unmarshal Dns Change Req Failed: %v", err)
			} else {
				handleDNSChanRequest(req)
			}
		}
	}
}

func handleDNSChanRequest(req *DNSRequest) {
	switch req.Type {
	case "PUT":
		if _, exist := dnsMap[req.Key]; exist {
			dnsMap[req.Key] = req.DNS
			klog.Infof("New DNS Added: Key[%v] Value[...]", req.Key)
			klog.Infof("Current dns num: %v", len(dnsMap))
			installDNS(req.DNS)
		} else {
			dnsMap[req.Key] = req.DNS
			klog.Infof("DNS Changed: Key[%v] Value[...]", req.Key)
			klog.Infof("Current dns num: %v", len(dnsMap))
		}

	case "DELETE":
		delete(dnsMap, req.Key)
		klog.Infof("Dns Deleted: Key[%v]", req.Key)
		klog.Infof("Current dns num: %v", len(dnsMap))
	}
}

func installDNS(dns v1.DNS) {

	services_raw := apiclient.Rest("", "", apiclient.OBJ_ALL_SERVICES, apiclient.OP_GET)
	var svcs []ServiceRequest
	err := json.Unmarshal(services_raw, &svcs)
	if err != nil {
		klog.Error("Unmarshal Service Failed: %v", err)
	} else {
		for _, path := range dns.Paths {
			klog.Infof("Installing DNS Rule Path[%v] ServiceName[%v]", path.Path, path.ServiceName)
			exist := false
			for _, svcReq := range svcs {
				if svcReq.Service.Name == path.ServiceName {
					cmd := exec.Command("weave",
						"dns-add",
						svcReq.Service.Spec.ClusterIP,
						"-h",
						dns.Host+path.Path+".weave.local")
					if err := cmd.Start(); err != nil {
						klog.Error("Error:The command is err,", err)
						return
					}
					exist = true
					break
				}
			}
			if !exist {
				klog.Error("Cannot Find ServiceName[%v]", path.ServiceName)
			}
		}
	}

}
