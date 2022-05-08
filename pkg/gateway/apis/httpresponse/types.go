package httpresponse

import (
	v1 "minik8s.com/minik8s/pkg/api/v1"
)

type WatchNodeResponse struct {
	Key string `json:"key"`

	Node v1.Node `json:"value"`

	Type string `json:"type"`
}
