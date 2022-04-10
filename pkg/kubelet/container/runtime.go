package container

import(
	"minik8s.com/minik8s/pkg/kubelet/types"
)

type Pod struct {
	ID types.UID

	Name string
	Namespace string

	Container []*Container
}

type PodStatus struct {
	ID types.UID
	Name string
	Namespace string
}

type Container struct{
	ID string
	Name string

	Image string
	ImageID string
	
	State string
}

/**
 * enum 4 state of container
 * ContainerStateCreated: a container that has been created (e.g. with docker create) but not started.
 * ContainerStateRunning: a currently running container.
 * ContainerStateExited: a container that ran and completed ("stopped" in other contexts, although a created container is technically also "stopped").
 * ContainerStateUnknown: encompasses all the states that we currently don't care about (like restarting, paused, dead).
 */
const (
	ContainerStateCreated string = "created"
	ContainerStateRunning string = "running"
	ContainerStateExited  string = "exited"
	ContainerStateUnknown string = "unknown"
)

type ContainerStatus struct {
	ID string
	Name string

	State string
}

type Annotation struct {
	Name string
	Value string
}

type ImageSpec struct {
	Name string
	Annotation []Annotation
}

type Image struct {
	ID string

	Size int64

	ImageSpec ImageSpec
}

type Runtime interface {
	Type() string
	
	GetPods(all bool) ([]*Pod, error)
	KillPod(pod Pod) error
	GetPodStatus(uid types.UID, name, namespace string) (*PodStatus, error)
	
	DeleteContainer(containerID string) error
}

type ImageService interface {
	PullImage(Image ImageSpec) (string, error)
	ListImages() ([]Image, error)
	RemoveImage(image ImageSpec) error
}

