package kubeproxy

import (
	"context"
	"errors"
	"fmt"

	"k8s.io/klog"
	v1 "minik8s.com/minik8s/pkg/api/v1"
)

type KubeProxy interface {
	AddService(ctx context.Context, service v1.Service) error

	RemoveService(ctx context.Context, name string) error

	AddEndpoint(ctx context.Context, endpoint v1.Endpoint) error

	RemoveEndpoint(ctx context.Context, name string) error

	UpdateEndpoint(ctx context.Context, endpoint v1.Endpoint) error
}

type kubeProxy struct {
	services  map[string]*v1.Service  // key is service name
	endpoints map[string]*v1.Endpoint // key is endpoints name
}

func NewKubeProxy() (KubeProxy, error) {
	kp := &kubeProxy{
		services:  make(map[string]*v1.Service),
		endpoints: make(map[string]*v1.Endpoint),
	}

	return kp, nil
}

func (kp *kubeProxy) AddService(ctx context.Context, service v1.Service) error {
	_, ok := kp.services[service.Name]

	if ok {
		err := fmt.Sprintf("Add service error: service %s already exists", service.Name)
		klog.Error(err)
		return errors.New(err)
	}

	if service.Spec.ClusterIP == "" {
		err := fmt.Sprintf("Add service error: service %s cluster ip is null", service.Name)
		klog.Error(err)
		return errors.New(err)
	}

	klog.Infof("Add service %s, cluster IP: %s", service.Name, service.Spec.ClusterIP)
	kp.services[service.Name] = &service

	return nil
}

func (kp *kubeProxy) RemoveService(ctx context.Context, name string) error {
	service, ok := kp.services[name]

	if !ok {
		err := fmt.Sprintf("Remove service error: service %s doesn't exist", service.Name)
		klog.Error(err)
		return errors.New(err)
	}

	klog.Infof("Add service %s, cluster IP: %s", service.Name, service.Spec.ClusterIP)
	delete(kp.services, name)

	return nil
}

func (kp *kubeProxy) AddEndpoint(ctx context.Context, endpoint v1.Endpoint) error {
	_, ok := kp.endpoints[endpoint.Name]

	if ok {
		err := fmt.Sprintf("Add endpoint error: endpoint %s already exists", endpoint.Name)
		klog.Error(err)
		return errors.New(err)
	}

	if len(endpoint.Subsets) == 0 {
		err := fmt.Sprintf("Add endpoint error: endpoint %s subset is null", endpoint.Name)
		klog.Error(err)
		return errors.New(err)
	}

	kp.endpoints[endpoint.Name] = &endpoint
	
	return nil
}

func (kp *kubeProxy) RemoveEndpoint(ctx context.Context, name string) error {
	return nil
}

func (kp *kubeProxy) UpdateEndpoint(ctx context.Context, endpoint v1.Endpoint) error {
	return nil
}
