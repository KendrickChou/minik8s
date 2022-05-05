package component

import (
	v1 "minik8s.com/minik8s/pkg/api/v1"
)

type Delta interface {
	GetType() string
	GetKey() string
	GetValue() any
}

type DeltaPart struct {
	Type string `json:"type,omitempty"`
	Key  string `json:"key"`
}

type PodObject struct {
	DeltaPart
	Pod v1.Pod `json:"value"`
}

type ReplicaSetObject struct {
	DeltaPart
	ReplicaSet v1.ReplicaSet `json:"value"`
}

type ServiceObject struct {
	DeltaPart
	Service v1.Service `json:"value"`
}

type EndpointObject struct {
	DeltaPart
	Endpoint v1.Endpoint `json:"value"`
}

func (d *DeltaPart) GetType() string {
	return d.Type
}

func (d *DeltaPart) GetKey() string {
	return d.Key
}

func (p *PodObject) GetValue() any {
	return p.Pod
}

func (rs *ReplicaSetObject) GetValue() any {
	return rs.ReplicaSet
}

func (s *ServiceObject) GetValue() any {
	return s.Service
}

func (ed *EndpointObject) GetValue() any {
	return ed.Endpoint
}
