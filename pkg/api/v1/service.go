package v1

type Service struct {
	TypeMeta
	ObjectMeta `json:"metadata,omitempty"`
	Spec       ServiceSpec   `json:"spec,omitempty"`
	Status     ServiceStatus `json:"status,omitempty"`
}

type ServiceSpec struct {
	Ports     []ServicePort     `json:"ports,omitempty"`
	Selector  map[string]string `json:"selector,omitempty"`
	ClusterIP string            `json:"clusterIP,omitempty"`
	Type      string            `json:"type,omitempty"`
}

type ServiceStatus struct {
	// TODO: implement ServiceStatus
}

type ServicePort struct {
	Name       string `json:"name,omitempty"`
	Protocol   string `json:"protocol,omitempty"`
	Port       int32  `json:"port"`
	TargetPort int32  `json:"targetPort,omitempty"`
	NodePort   int32  `json:"nodePort,omitempty"`
}
