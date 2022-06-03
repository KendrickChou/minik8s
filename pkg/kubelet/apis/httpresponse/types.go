package httpresponse

import (
	v1 "minik8s.com/minik8s/pkg/api/v1"
)

type PodChangeRequest struct {
	Key string `json:"key"`

	Pod v1.Pod `json:"value"`

	Type string `json:"type"`
}

type RegistResponse struct {
	UID string `json:"id"`
}

type WatchNodeResponse struct {
	Key string `json:"key"`

	Node v1.Node `json:"value"`

	Type string `json:"type"`
}

type EndpointChangeRequest struct {
	Key string `json:"key"`

	Endpoint v1.Endpoint `json:"value"`

	Type string `json:"type"`
}
