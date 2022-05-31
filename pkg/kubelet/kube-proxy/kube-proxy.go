package kubeproxy

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os/exec"
	"regexp"
	"strings"

	"k8s.io/klog"
	v1 "minik8s.com/minik8s/pkg/api/v1"
	"minik8s.com/minik8s/pkg/kubelet/apis/config"
	"minik8s.com/minik8s/pkg/kubelet/apis/constants"
	"minik8s.com/minik8s/pkg/kubelet/apis/httpresponse"

	"github.com/coreos/go-iptables/iptables"
	"github.com/dchest/uniuri"
)

type KubeProxy interface {
	AddEndpoint(ctx context.Context, uid string, endpoint v1.Endpoint) error

	RemoveEndpoint(ctx context.Context, name string) error

	UpdateEndpoint(ctx context.Context, uid string, endpoint v1.Endpoint) error
}

type kubeProxy struct {
	endpoints map[string]*v1.Endpoint // key is endpoints uid

	ipt *iptables.IPTables

	SVCChain map[string][]string // map service name -> iptables svc chain name array
	SEPChain map[string][]string // map iptables svc chain name -> a sep chain name array
}

func NewKubeProxy() (KubeProxy, error) {
	kp := &kubeProxy{
		endpoints: make(map[string]*v1.Endpoint),
		SVCChain:  make(map[string][]string),
		SEPChain:  make(map[string][]string),
	}

	ipt, err := iptables.New()

	if err != nil {
		return &kubeProxy{}, err
	}

	kp.ipt = ipt

	err = initIPtables(ipt)
	if err != nil {
		klog.Errorf("Init IP tables Error: %s", err.Error())
		return nil, err
	}
	klog.Info("Init IP Tables successfully!")

	// get all existed endpoints
	resp, err := http.Get(config.ApiServerAddress + constants.GetAllEndpointsRequest())

	if err != nil {
		klog.Errorf("Get All Existed Endpoints Error: %s", err.Error())
		return nil, err
	}

	buf, err := io.ReadAll(resp.Body)

	resp.Body.Close()

	if err != nil {
		klog.Errorf("Get All Existed Endpoints Error: %s", err.Error())
		return nil, err
	}

	var epArray []httpresponse.EndpointChangeRequest
	json.Unmarshal(buf, &epArray)

	for _, ep := range epArray {
		klog.Infof("Add Endpoint %s, key: %s ", ep.Endpoint.Name, ep.Key)

		parsedPath := strings.Split(ep.Key, "/")
		uid := parsedPath[len(parsedPath) - 1]

		kp.AddEndpoint(context.TODO(), uid, ep.Endpoint)
	}

	klog.Infof("All existed endpoints are added!")
	klog.Infof("Kube-proxy is ready to go")

	return kp, err
}

func (kp *kubeProxy) AddEndpoint(ctx context.Context, uid string, endpoint v1.Endpoint) error {
	klog.Infof("Add Endpoint %s", endpoint.Name)
	_, ok := kp.endpoints[uid]

	if ok {
		klog.Infof("Add endpoint %s already exists, try to update it", endpoint.Name)
		kp.UpdateEndpoint(ctx, uid, endpoint)
		return nil
	}

	if len(endpoint.Subset) == 0 {
		err := fmt.Sprintf("Add endpoint error: endpoint %s subset is null", endpoint.Name)
		klog.Error(err)
		return errors.New(err)
	}

	kp.endpoints[uid] = &endpoint

	svcChains := []string{}
	serviceChainMap := make(map[string]string) // map port/protocol -> svcID.

	for _, subset := range endpoint.Subset {
		for _, port := range subset.Ports {
			key := fmt.Sprintf("%d/%s", port.ServicePort, port.Protocol)
			_, ok := serviceChainMap[key]

			if !ok {
				// create svc chain
				svcID := createAServiceChainID()
				err := kp.ipt.NewChain(constants.NATTableName, svcID)
				svcChains = append(svcChains, svcID)
				
				if err != nil {
					klog.Errorf("Create Service Chain %s Error: %s", svcID, err.Error())
				}
				
				// append to K8S-SERVICE chain
				rule := fmt.Sprintf("-d %s -p %s --dport %d -m comment --comment %s -j %s", endpoint.ServiceIp, port.Protocol, port.ServicePort, endpoint.Name, svcID)
				err = kp.ipt.Append(constants.NATTableName, constants.ServiceChainName, strings.Split(rule, " ")...)
				
				if err != nil {
					klog.Errorf("Append %s to K8S-SERVICE Chain Error: %s", svcID, err.Error())
				}

				serviceChainMap[key] = svcID
			}
		}
	}

	for _, subset := range endpoint.Subset {
		for _, port := range subset.Ports {
			svcID := serviceChainMap[fmt.Sprintf("%d/%s", port.ServicePort, port.Protocol)]

			sepChains := []string{}
			for i, addr := range subset.Addresses {
				// create sep chain
				sepID := createASEPChainID()
				err := kp.ipt.NewChain(constants.NATTableName, sepID)
				sepChains = append(sepChains, sepID)
	
				if err != nil {
					klog.Errorf("Create SEP Chain %s Error: %s", sepID, err.Error())
				}
	
				// append to K8S-SVC-xxx chain
				p := 1 / float32(len(subset.Addresses))
				var rule string
				if i != len(subset.Addresses)-1 {
					rule = fmt.Sprintf("-m statistic --mode random --probability %f -j %s", p, sepID)
				} else {
					rule = fmt.Sprintf("-j %s", sepID)
				}
	
				err = kp.ipt.Append(constants.NATTableName, svcID, strings.Split(rule, " ")...)
	
				if err != nil {
					klog.Errorf("Append %s to SVC Chain %s Error: %s", sepID, svcID, err.Error())
				}
	
				// append to K8S-SEP-xxx
				rule = fmt.Sprintf("-j DNAT -p %s --to %s:%d", port.Protocol, addr.IP, port.Port)
				err = kp.ipt.Append(constants.NATTableName, sepID, strings.Split(rule, " ")...)
	
				if err != nil {
					klog.Errorf("Append DNAT Rule to SEP Chain %s Error: %s", sepID, err.Error())
				}
			}
	
			kp.SEPChain[svcID] = sepChains
		}
	}

	kp.SVCChain[uid] = svcChains

	return nil
}

func (kp *kubeProxy) RemoveEndpoint(ctx context.Context, name string) error {
	klog.Infof("Remove Endpoint %s", name)
	_, ok := kp.endpoints[name]

	if !ok {
		err := fmt.Sprintf("Remove endpoint error: endpoint %s doesn't exist", name)
		klog.Error(err)
		return errors.New(err)
	}

	delete(kp.endpoints, name)

	for _, svcID := range kp.SVCChain[name] {

		res, err := deleteRulesByTarget(constants.NATTableName, constants.ServiceChainName, svcID)

		if err != nil {
			klog.Errorf("Remove rules %s from K8S-SERVICE error: %s %s", svcID, err.Error(), res)
		}

		err = kp.ipt.ClearAndDeleteChain(constants.NATTableName, svcID)

		if err != nil {
			klog.Errorf("Remove Chain %s error: %s", svcID, err.Error())
		}

		for _, sepID := range kp.SEPChain[svcID] {
			err = kp.ipt.ClearAndDeleteChain(constants.NATTableName, sepID)
			if err != nil {
				klog.Errorf("Remove Chain %s error: %s", sepID, err.Error())
			}
		}

		delete(kp.SEPChain, svcID)
	}

	delete(kp.SVCChain, name)

	return nil
}

func (kp *kubeProxy) UpdateEndpoint(ctx context.Context, uid string, endpoint v1.Endpoint) error {
	klog.Infof("Update Endpoint %s, UID: %s", endpoint.Name, uid)

	err := kp.RemoveEndpoint(context.TODO(), uid)

	if err != nil {
		return err
	}

	_ = kp.AddEndpoint(context.TODO(), uid, endpoint)

	return nil
}

func createAServiceChainID() string {
	return constants.ServiceChainPrefix + strings.ToUpper(uniuri.New())
}

func createASEPChainID() string {
	return constants.SepChainPrefix + strings.ToUpper(uniuri.New())
}

func initIPtables(ipt *iptables.IPTables) error {
	// clear all remanent K8S-SVC-xxx K8S-SEP-xxx
	chains, err := ipt.ListChains(constants.NATTableName)

	if err != nil {
		errInfo := fmt.Sprintf("Check K8S-SERVICE chain error: %s", err.Error())
		klog.Error(errInfo)
		return errors.New(errInfo)
	}

	svcChain := regexp.MustCompile("K8S-SVC")
	sepChain := regexp.MustCompile("K8S-SEP")

	// clear K8S-SERVICE first
	for _, chain := range chains {
		if chain == constants.ServiceChainName {
			ipt.ClearChain(constants.NATTableName, chain)
		}
	}

	// then clean and delete K8S-SVC-
	for _, chain := range chains {
		if svcChain.MatchString(chain) {
			ipt.ClearAndDeleteChain(constants.NATTableName, chain)
		}
	}

	// clean and delete K8S-SEP-*
	for _, chain := range chains {
		if sepChain.MatchString(chain) {
			ipt.ClearAndDeleteChain(constants.NATTableName, chain)
		}
	}

	// return if already exist
	exist, err := ipt.ChainExists(constants.NATTableName, constants.ServiceChainName)

	if err != nil {
		errInfo := fmt.Sprintf("Check K8S-SERVICE chain error: %s", err.Error())
		return errors.New(errInfo)
	}

	if exist {
		// klog.Info("K8S-SERVICE Chain already exists")
		return nil
	}

	klog.Info("Create K8S-SERVICE Chain")

	// create K8S-SERVICE
	err = ipt.NewChain(constants.NATTableName, constants.ServiceChainName)

	if err != nil {
		return err
	}

	// append it to OUTPUT Chain & PREROUTING chain
	rule := "-j " + constants.ServiceChainName

	err = ipt.Insert(constants.NATTableName, "OUTPUT", 1, strings.Split(rule, " ")...)
	if err != nil {
		return err
	}

	err = ipt.Insert(constants.NATTableName, "PREROUTING", 1, strings.Split(rule, " ")...)
	if err != nil {
		return err
	}

	klog.Info("Add K8S-SERVICE Chain to OUTPUT & PREROUTING")

	return nil
}

func deleteRulesByTarget(table, chain, target string) (string, error) {
	command := fmt.Sprintf("iptables -t %s -D %s $(iptables -t %s --line-number -nL %s | grep %s | awk '{print $1}' | tac)",
		table, chain, table, chain, target)

	cmd := exec.Command("bash", "-c", command)

	res, err := cmd.CombinedOutput()

	return string(res), err
}
