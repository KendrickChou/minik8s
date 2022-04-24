package v1

import "time"

// ContainerStateWaiting is a waiting state of a container.
type ContainerStateWaiting struct {
	// (brief) reason the container is not yet running.
	Reason string `json:"reason,omitempty"`
	// Message regarding why the container is not yet running.
	Message string `json:"message,omitempty"`
}

// ContainerStateRunning is a running state of a container.
type ContainerStateRunning struct {
	// Time at which the container was last (re-)started
	StartedAt time.Time `json:"startedAt,omitempty"`
}

// ContainerStateTerminated is a terminated state of a container.
type ContainerStateTerminated struct {
	// Exit status from the last termination of the container
	ExitCode int32 `json:"exitCode"`
	// Signal from the last termination of the container
	Signal int32 `json:"signal,omitempty"`
	// (brief) reason from the last termination of the container
	Reason string `json:"reason,omitempty"`
	// Message regarding the last termination of the container
	Message string `json:"message,omitempty"`
	// Time at which previous execution of the container started
	StartedAt time.Time `json:"startedAt,omitempty"`
	// Time at which the container last terminated
	FinishedAt time.Time `json:"finishedAt,omitempty"`
	// Container's ID in the format '<type>://<container_id>'
	ContainerID string `json:"containerID,omitempty"`
}

// ContainerState holds a possible state of container.
// Only one of its members may be specified.
// If none of them is specified, the default one is ContainerStateWaiting.
type ContainerState struct {
	// Details about a waiting container
	// +optional
	Waiting *ContainerStateWaiting `json:"waiting,omitempty"`
	// Details about a running container
	// +optional
	Running *ContainerStateRunning `json:"running,omitempty"`
	// Details about a terminated container
	// +optional
	Terminated *ContainerStateTerminated `json:"terminated,omitempty"`
}

type Container struct {
	Name string `json:"Name,omitempty"`

	Namespace string `json:"Namespace,omitempty"`

	ID string `json:"ID,omitempty"`

	Image string `json:"Image,omitempty"`

	//"Always" means that kubelet always attempts to pull the latest image. Container will fail If the pull fails.
	//"IfNotPresent" means that kubelet pulls if the image isn't present on disk. Container will fail if the image isn't present and the pull fails.
	//"Never" means that kubelet never pulls an image, but only uses a local image. Container will fail if the image isn't present
	//default: IfNotPresent
	ImagePullPolicy string `json:"ImagePolicy,omitempty" default:"IfNotPresent"`

	// Command to run when starting the container
	Command []string `json:"Command,omitempty"`

	Entrypoint []string `json:"Entrypoint,omitempty"`

	// Container's working directory.
	// If not specified, the container runtime's default will be used, which
	// might be configured in the container image.
	WorkingDir string `json:"WorkingDir,omitempty"`

	Env []string `json:"Env,omitempty"`

	// mount volumes
	Mounts []Mount `json:"Mounts,omitempty"`
}

type ContainerStatus struct {
	Name string `json:"Name,omitempty"`

	State ContainerState `json:"State,omitempty"`
}

type Mount struct {
	Type        string `json:"MountType,omitempty"`
	Source      string `json:"MountSource,omitempty"`
	Target      string `json:"MountTarget,omitempty"`
	Consistency string `json:"MountConsistency,omitempty"`
}

const (
	AlwaysImagePullPolicy       string = "Always"
	IfNotPresentImagePullPolicy string = "IfNotPresent"
	NeverPullPolicy             string = "Never"

	// TypeBind is the type for mounting host dir
	TypeBind string = "bind"
	// TypeVolume is the type for remote storage volumes
	TypeVolume string = "volume"

	// ConsistencyFull guarantees bind mount-like consistency
	ConsistencyFull string = "consistent"
	// ConsistencyCached mounts can cache read data and FS structure
	ConsistencyCached string = "cached"
	// ConsistencyDelegated mounts can cache read and written data and structure
	ConsistencyDelegated string = "delegated"
	// ConsistencyDefault provides "consistent" behavior unless overridden
	ConsistencyDefault string = "default"
)
