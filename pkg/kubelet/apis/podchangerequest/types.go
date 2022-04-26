package podchangerequest

import v1 "minik8s.com/minik8s/pkg/api/v1"

type PodChangeRequest struct {
	Key string

	Pod v1.Pod `json:"Value"`

	Type string
}