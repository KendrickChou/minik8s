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

	ID string `json:"ID,omitempty"`

	Image string `json:"Image,omitempty"`

	// Entrypoint array. Not executed within a shell.
	// The container image's ENTRYPOINT is used if this is not provided.
	Command []string `json:"Command,omitempty"`

	// Arguments to the entrypoint.
	// The container image's CMD is used if this is not provided.
	Args []string `json:"Args,omitempty"`

	// Container's working directory.
	// If not specified, the container runtime's default will be used, which
	// might be configured in the container image.
	WorkingDir string `json:"WorkingDir,omitempty"`

}

type ContainerStatus struct {
	Name string `json:"Name,omitempty"`

	State ContainerState `json:"State,omitempty"`
}