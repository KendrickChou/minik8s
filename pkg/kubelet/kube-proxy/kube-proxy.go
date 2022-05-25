package kubeproxy

import (
	"context"
	"errors"
	"fmt"
	"os/exec"
	"regexp"
	"strings"

	"k8s.io/klog"
	v1 "minik8s.com/minik8s/pkg/api/v1"
	"minik8s.com/minik8s/pkg/kubelet/apis/constants"

	"github.com/coreos/go-iptables/iptables"
	"github.com/dchest/uniuri"
)

type KubeProxy interface {
	AddEndpoint(ctx context.Context, uid string, endpoint v1.Endpoint) error

	RemoveEndpoint(ctx context.Context, name string) error

	UpdateEndpoint(ctx context.Context, endpoint v1.Endpoint) error
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
	klog.Info("Init IP Tables successfully!")

	return kp, err
}

func (kp *kubeProxy) AddEndpoint(ctx context.Context, uid string, endpoint v1.Endpoint) error {
	klog.Infof("Add Endpoint %s", endpoint.Name)
	_, ok := kp.endpoints[uid]

	if ok {
		err := fmt.Sprintf("Add endpoint error: endpoint %s already exists", endpoint.Name)
		klog.Error(err)
		return errors.New(err)
	}

	if len(endpoint.Subset.Addresses) == 0 || len(endpoint.Subset.Ports) == 0 {
		err := fmt.Sprintf("Add endpoint error: endpoint %s subset is null", endpoint.Name)
		klog.Error(err)
		return errors.New(err)
	}

	kp.endpoints[uid] = &endpoint

	svcChains := []string{}
	for _, port := range endpoint.Subset.Ports {
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

		sepChains := []string{}
		for i, addr := range endpoint.Subset.Addresses {
			// create sep chain
			sepID := createASEPChainID()
			err = kp.ipt.NewChain(constants.NATTableName, sepID)
			sepChains = append(sepChains, sepID)

			if err != nil {
				klog.Errorf("Create SEP Chain %s Error: %s", sepID, err.Error())
			}

			// append to K8S-SVC-xxx chain
			p := 1 / float32(len(endpoint.Subset.Addresses))
			if i != len(endpoint.Subset.Addresses)-1 {
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

	kp.SVCChain[endpoint.Name] = svcChains

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

func (kp *kubeProxy) UpdateEndpoint(ctx context.Context, endpoint v1.Endpoint) error {
	klog.Infof("Update Endpoint %s", endpoint.Name)

	err := kp.RemoveEndpoint(context.TODO(), endpoint.Name)

	if err != nil {
		return err
	}

	err = kp.AddEndpoint(context.TODO(), endpoint)

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
		return errors.New(errInfo)
	}

	for _, chain := range chains {
		match, _ := regexp.MatchString("K8S", chain)
		if match {
			if chain == constants.ServiceChainName {
				ipt.ClearChain(constants.NATTableName, chain)
			} else {
				ipt.ClearAndDeleteChain(constants.NATTableName, chain)
			}
		}
	}

	// return if already exist
	exist, err := ipt.ChainExists(constants.NATTableName, constants.ServiceChainName)

	if err != nil {
		errInfo := fmt.Sprintf("Check K8S-SERVICE chain error: %s", err.Error())
		return errors.New(errInfo)
	}

	if exist {
		klog.Info("K8S-SERVICE Chain already exists")
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
