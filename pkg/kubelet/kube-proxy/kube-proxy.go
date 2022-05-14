package kubeproxy

import (
	"context"
	"errors"
	"fmt"

	"k8s.io/klog"
	v1 "minik8s.com/minik8s/pkg/api/v1"
	"minik8s.com/minik8s/pkg/kubelet/apis/constants"

	"github.com/coreos/go-iptables/iptables"
	"github.com/google/uuid"
)

type KubeProxy interface {
	AddEndpoint(ctx context.Context, endpoint v1.Endpoint) error

	RemoveEndpoint(ctx context.Context, name string) error

	UpdateEndpoint(ctx context.Context, endpoint v1.Endpoint) error
}

type kubeProxy struct {
	endpoints map[string]*v1.Endpoint // key is endpoints name

	ipt *iptables.IPTables

	SVCChain map[string][]string	// map service name -> iptables svc chain name array
	SEPChain map[string][]string	// map iptables svc chain name -> a sep chain name array
}

func NewKubeProxy() (KubeProxy, error) {
	kp := &kubeProxy{
		endpoints: make(map[string]*v1.Endpoint),
		SVCChain: make(map[string][]string),
		SEPChain: make(map[string][]string),
	}

	ipt, err := iptables.New()

	if err != nil {
		return &kubeProxy{}, err
	}

	kp.ipt = ipt

	return kp, nil
}

func (kp *kubeProxy) AddEndpoint(ctx context.Context, endpoint v1.Endpoint) error {
	klog.Infof("Add Endpoint %s", endpoint.Name)
	_, ok := kp.endpoints[endpoint.Name]

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

	if !ok {
		err := fmt.Sprintf("Add endpoint error: no corresponding service %s", endpoint.Name)
		klog.Error(err)
		return errors.New(err)
	}

	kp.endpoints[endpoint.Name] = &endpoint

	svcChains := []string{}
	for _, port := range endpoint.Subset.Ports {
		// create svc chain
		svcID := createAServiceChainID()
		kp.ipt.NewChain(constants.NATTableName, svcID)
		svcChains = append(svcChains, svcID)

		// append to K8S-SERVICE chain
		rule := fmt.Sprintf("-d %s -p %s --dport %d -j %s", endpoint.ServiceIp, port.Protocol, port.ServicePort, svcID)
		kp.ipt.Append(constants.NATTableName, constants.ServiceChainName, rule)

		sepChains := []string{}
		for i, addr := range endpoint.Subset.Addresses {
			// create sep chain
			sepID := createASEPChainID()
			kp.ipt.NewChain(constants.NATTableName, sepID)
			sepChains = append(sepChains, sepID)

			// append to K8S-SVC-xxx chain
			p := 1 / float32(len(endpoint.Subset.Addresses))
			if i != len(endpoint.Subset.Addresses) - 1 {
				rule = fmt.Sprintf("-m statistics --mode random --probability %f -j %s", p, sepID)
			} else {
				rule = fmt.Sprintf("-j %s", sepID)
			}

			kp.ipt.Append(constants.NATTableName, svcID, rule)

			// append to K8S-SEP-xxx
			rule = fmt.Sprintf("--to %s:%d -j DNAT", addr.IP, port.Port)
			kp.ipt.Append(constants.NATTableName, sepID, rule)
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
		rule := fmt.Sprintf("-j %s", svcID)

		kp.ipt.Delete(constants.NATTableName, constants.ServiceChainName, rule)

		kp.ipt.DeleteChain(constants.NATTableName, svcID)

		for _, sepID := range kp.SEPChain[svcID] {
			kp.ipt.DeleteChain(constants.NATTableName, sepID)
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
	return constants.ServiceChainPrefix + uuid.NewString()
}

func createASEPChainID() string {
	return constants.SepChainPrefix + uuid.NewString()
}
