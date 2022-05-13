package v1

type RestartPolicy string
type PodPhase string

const (
	RestartPolicyAlways    RestartPolicy = "Always"
	RestartPolicyOnFailure RestartPolicy = "OnFailure"
	RestartPolicyNever     RestartPolicy = "Never"
)

const (
	// PodPending means the pod has been accepted by the system, but one or more of the containers
	// has not been started. This includes time before being bound to a node, as well as time spent
	// pulling images onto the host.
	PodPending PodPhase = "Pending"
	// PodRunning means the pod has been bound to a node and all of the containers have been started.
	// At least one container is still running or is in the process of being restarted.
	PodRunning PodPhase = "Running"
	// PodSucceeded means that all containers in the pod have voluntarily terminated
	// with a container exit code of 0, and the system is not going to restart any of these containers.
	PodSucceeded PodPhase = "Succeeded"
	// PodFailed means that all containers in the pod have terminated, and at least one container has
	// terminated in a failure (exited with a non-zero exit code or was stopped by the system).
	PodFailed PodPhase = "Failed"
	// PodUnknown means that for some reason the state of the pod could not be obtained, typically due
	// to an error in communicating with the host of the pod.
	PodUnknown PodPhase = "Unknown"
)

type Pod struct {
	TypeMeta

	ObjectMeta

	Spec PodSpec `json:"Spec,omitempty"`

	Status PodStatus `json:"Status,omitempty"`
}

type PodSpec struct {
	InitialContainers map[string]Container

	Containers []*Container `json:"Containers,omitempty"`

	RestartPolicy RestartPolicy `json:"RestartPolicy,omitempty"`

	NodeName string `json:"NodeName,omitempty"`
}

type PodStatus struct {
	Phase PodPhase `json:"Phase,omitempty"`

	PodIP string `json:"PodIP,omitempty"`

	PodNetworkID string `json:"-"`

	ContainerStatuses []ContainerStatus `json:"ContainerStatuses,omitempty"`
}
