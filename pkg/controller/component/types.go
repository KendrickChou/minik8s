package component

import (
	v1 "minik8s.com/minik8s/pkg/api/v1"
	"strings"
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

func (dp *DeltaPart) StripKey() {
	slices := strings.Split(dp.Key, "/")
	dp.Key = slices[len(slices)-1]
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

type PodStatusObject struct {
	DeltaPart
	PodStatus v1.PodStatus `json:"value"`
}

type HPAObject struct {
	DeltaPart
	HPA v1.HorizontalPodAutoscaler `json:"value"`
}

type JobObject struct {
	DeltaPart
	Job v1.GPUJob `json:"value"`
}

func (dp DeltaPart) GetType() string {
	return dp.Type
}

func (dp DeltaPart) GetKey() string {
	return dp.Key
}

func (p PodObject) GetValue() any {
	return p.Pod
}

func (rs ReplicaSetObject) GetValue() any {
	return rs.ReplicaSet
}

func (s ServiceObject) GetValue() any {
	return s.Service
}

func (ed EndpointObject) GetValue() any {
	return ed.Endpoint
}

func (ps PodStatusObject) GetValue() any {
	return ps.PodStatus
}

func (h HPAObject) GetValue() any {
	return h.HPA
}

func (j JobObject) GetValue() any {
	return j.Job
}
