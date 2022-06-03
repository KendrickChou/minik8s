package podchangerequest

import v1 "minik8s.com/minik8s/pkg/api/v1"

type PodChangeRequest struct {
	Key string `json:"key"`

	Pod v1.Pod `json:"value"`

	Type string `json:"type"`
}
