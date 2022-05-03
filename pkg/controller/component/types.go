package utils

import (
	v1 "minik8s.com/minik8s/pkg/api/v1"
	"minik8s.com/minik8s/pkg/controller/replicaset"
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
	ReplicaSet replicaset.ReplicaSet `json:"value"`
}

type ServiceObject struct {
	DeltaPart
	// TODO: add Service struct
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
	// TODO: return a real service object
	return nil
}
