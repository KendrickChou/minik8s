package container

import (
	"context"
	"errors"
	"time"

	"github.com/containerd/containerd"
	"github.com/containerd/containerd/api/services/tasks/v1"
	"github.com/containerd/containerd/api/types"
	"github.com/containerd/containerd/cio"
	"k8s.io/klog/v2"
)

type ContainerManager interface {
	CreateContainer(ctx context.Context, id string, opts ...containerd.NewContainerOpts) (string, error)
	StartContainer(ctx context.Context, containerID string, opts ...cio.Opt) error
	StopContainer(ctx context.Context, containerID string) error
	PauseContainer(ctx context.Context, containerID string) error
	ResumeContainer(ctx context.Context, containerID string) error
	RemoveContainer(ctx context.Context, containerID string) error
	ListContainers(ctx context.Context) ([]containerd.Container, error)
	ContainerStatus(ctx context.Context, containerID string) (containerd.ProcessStatus, error)
}

type containerManager struct {
	client *containerd.Client

	containers map[string]containerd.Container
}

func NewContainerManager(client *containerd.Client) (containerManager, error) {
	containerManager := containerManager{client: client}
	containerManager.containers = make(map[string]containerd.Container)

	return containerManager, nil
}

func (manager *containerManager) CreateContainer(ctx context.Context, id string, opts ...containerd.NewContainerOpts) (string, error) {
	container, err := manager.client.NewContainer(ctx, id, opts...)

	if err != nil {
		klog.Errorln("[Container] create container ", id, " failed : ", err)
		return "", err
	}

	manager.containers[container.ID()] = container

	return container.ID(), err
}

func (manager *containerManager) StartContainer(ctx context.Context, containerID string, opts ...cio.Opt) error {
	container, ok := manager.containers[containerID]

	if !ok {
		return errors.New("Container " + containerID + " does not exist")
	}

	// we can only run a single task at a container
	if task, _ := container.Task(ctx, nil); task != nil {
		return errors.New("Task " + containerID + " already started")
	}

	// if opts is empty, use standard in/out

	if len(opts) == 0 {
		opts = append(opts, cio.WithStdio)
	}

	// we can start a user defined binary when a new task is created
	_, err := container.NewTask(ctx, cio.NewCreator(opts...))
	
	if err != nil {
		return nil
	}

	return nil
}

func (manager *containerManager) StopContainer(ctx context.Context, containerID string) error {
	container, ok := manager.containers[containerID]

	if !ok {
		return errors.New("Container " + containerID + " does not exist")
	}

	task, _ := container.Task(ctx, nil)
	if task == nil {
		return errors.New("Task " + containerID + " does not exist")
	}

	err := task.Kill(ctx, 9, containerd.WithKillAll)

	if err != nil {
		return err
	}

	return nil
}

func (manager *containerManager) PauseContainer(ctx context.Context, containerID string) error {
	container, ok := manager.containers[containerID]

	if !ok {
		return errors.New("Container " + containerID + " does not exist")
	}

	task, _ := container.Task(ctx, nil)
	if task == nil {
		return errors.New("Task " + containerID + " does not exist")
	}

	return task.Pause(ctx)
}

func (manager *containerManager) ResumeContainer(ctx context.Context, containerID string) error {
	container, ok := manager.containers[containerID]

	if !ok {
		return errors.New("Container " + containerID + " does not exist")
	}

	task, _ := container.Task(ctx, nil)
	if task == nil {
		return errors.New("Task " + containerID + " does not exist")
	}

	return task.Resume(ctx)
}

func (manager *containerManager) RemoveContainer(ctx context.Context, containerID string) error {
	manager.StopContainer(ctx, containerID)

	err := manager.client.ContainerService().Delete(ctx, containerID)

	if err != nil {
		return err
	}

	delete(manager.containers, containerID)

	return nil
}

func (manager *containerManager) ListContainers(ctx context.Context) ([]containerd.Container, error) {
	return manager.client.Containers(ctx)
}

func (manager *containerManager) ContainerStatus(ctx context.Context, containerID string) (containerd.Status, error) {
	container, ok := manager.containers[containerID]

	if !ok {
		return containerd.Status{}, errors.New("Container " + containerID + " does not exist")
	}

	task, err := container.Task(ctx, nil)

	if err != nil {
		return containerd.Status{}, err
	}

	return task.Status(ctx)
}


