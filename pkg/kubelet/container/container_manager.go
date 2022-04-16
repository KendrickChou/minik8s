package container

import (
	"context"
	"errors"

	"github.com/containerd/containerd"
	"github.com/containerd/containerd/cio"
	v1 "minik8s.com/minik8s/pkg/api/v1"
)

type ContainerManager interface {
	CreateContainer(ctx context.Context, container v1.Container) error
	StartContainer(ctx context.Context, container v1.Container) error
	StopContainer(ctx context.Context, container v1.Container) error
	PauseContainer(ctx context.Context, container v1.Container) error
	ResumeContainer(ctx context.Context, container v1.Container) error
	RemoveContainer(ctx context.Context, container v1.Container) error
	ListContainers(ctx context.Context) ([]v1.Container, error)
	ContainerStatus(ctx context.Context, containerID string) (containerd.ProcessStatus, error)
}

type containerManager struct {
	client *containerd.Client
}

func NewContainerManager(client *containerd.Client) (containerManager, error) {
	containerManager := containerManager{client: client}

	return containerManager, nil
}

func (manager *containerManager) CreateContainer(ctx context.Context, container v1.Container) error {

	var opts []containerd.NewContainerOpts

	switch {
	case container.Image != "":
		opts = append(opts, containerd.WithImageName(container.Image))
	default:
	}

	_, err := manager.client.NewContainer(ctx, container.ID, opts...)

	return err
}

func (manager *containerManager) StartContainer(ctx context.Context, container v1.Container) error {
	baseContainer, err := manager.client.LoadContainer(ctx, container.ID)

	if err != nil {
		return err
	}

	// we can only run a single task at a container
	if task, _ := baseContainer.Task(ctx, nil); task != nil {
		return errors.New("Task " + container.ID + " already started")
	}

	// if opts is empty, use standard in/out

	// if len(opts) == 0 {
	// 	opts = append(opts, cio.WithStdio)
	// }

	// we can start a user defined binary when a new task is created
	_, err = baseContainer.NewTask(ctx, cio.NewCreator(cio.WithStdio))
	
	return err
}

func (manager *containerManager) StopContainer(ctx context.Context, container v1.Container) error {
	baseContainer, err := manager.client.LoadContainer(ctx, container.ID)

	if err != nil {
		return err
	}

	task, _ := baseContainer.Task(ctx, nil)
	if task == nil {
		return errors.New("Task " + container.ID + " does not exist")
	}

	err = task.Kill(ctx, 9, containerd.WithKillAll)

	if err != nil {
		return err
	}

	return nil
}

func (manager *containerManager) PauseContainer(ctx context.Context, container v1.Container) error {
	baseContainer, err := manager.client.LoadContainer(ctx, container.ID)

	if err != nil {
		return err
	}


	task, _ := baseContainer.Task(ctx, nil)
	if task == nil {
		return errors.New("Task " + container.ID + " does not exist")
	}

	return task.Pause(ctx)
}

func (manager *containerManager) ResumeContainer(ctx context.Context, container v1.Container) error {
	baseContainer, err := manager.client.LoadContainer(ctx, container.ID)

	if err != nil {
		return err
	}

	task, _ := baseContainer.Task(ctx, nil)
	if task == nil {
		return errors.New("Task " + container.ID + " does not exist")
	}

	return task.Resume(ctx)
}

func (manager *containerManager) RemoveContainer(ctx context.Context, container v1.Container) error {
	manager.StopContainer(ctx, container)

	err := manager.client.ContainerService().Delete(ctx, container.ID)

	if err != nil {
		return err
	}

	return nil
}

func (manager *containerManager) ListContainers(ctx context.Context) ([]containerd.Container, error) {
	return manager.client.Containers(ctx)
}

func (manager *containerManager) ContainerStatus(ctx context.Context, containerID string) (containerd.Status, error) {
	baseContainer, err := manager.client.LoadContainer(ctx, containerID)

	if err != nil {
		return containerd.Status{}, err
	}

	task, err := baseContainer.Task(ctx, nil)

	if err != nil {
		return containerd.Status{}, err
	}

	return task.Status(ctx)
}


