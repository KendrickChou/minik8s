package container

import (
	"context"
	"errors"
	"fmt"
	"syscall"
	"time"

	"github.com/containerd/containerd"
	"github.com/containerd/containerd/cio"
	"github.com/containerd/containerd/namespaces"
	"github.com/containerd/containerd/oci"
	"github.com/containerd/containerd/pkg/cri"
	"k8s.io/klog/v2"
	v1 "minik8s.com/minik8s/pkg/api/v1"
)

type ContainerManager interface {
	CreateContainer(ctx context.Context, container *v1.Container) error
	StartContainer(ctx context.Context, container *v1.Container) error
	StopContainer(ctx context.Context, container *v1.Container) error
	PauseContainer(ctx context.Context, container *v1.Container) error
	ResumeContainer(ctx context.Context, container *v1.Container) error
	RemoveContainer(ctx context.Context, container *v1.Container) error
	ListContainers(ctx context.Context) ([]*v1.Container, error)
	ContainerStatus(ctx context.Context, containerID string) (containerd.Status, error)
}

type containerManager struct {
	client *containerd.Client
}

func NewContainerManager(address string) (ContainerManager, error) {
	client, err := containerd.New(address)

	containerManager := &containerManager{
		client: client,
	}

	return containerManager, err
}

func (manager *containerManager) CreateContainer(ctx context.Context, container *v1.Container) error {

	var opts []containerd.NewContainerOpts

	fmt.Println("container: ", container)

	switch {
	case container.Namespace != "":
		ctx = namespaces.WithNamespace(ctx, container.Namespace)
		fallthrough
	case container.Image != "":
		image, err := manager.client.Pull(ctx, container.Image, containerd.WithPullUnpack)
		if err != nil {
			return err
		}
		opts = append(opts, containerd.WithImage(image))
		opts = append(opts, containerd.WithNewSnapshot("hello-world-snapshot", image))
		opts = append(opts, containerd.WithNewSpec(oci.WithImageConfig(image)))
	default:
	}

	cont, err := manager.client.NewContainer(ctx, container.Name, opts...)

	if err != nil {
		klog.Error(err)
		return err
	}

	task, err := cont.NewTask(ctx, cio.NewCreator(cio.WithStdio))

	exitStatus, err := task.Wait(ctx)

	err = task.Start(ctx)

	if err != nil {
		klog.Error(err)
		return err
	}

	time.Sleep(3 * time.Second)

	task.Kill(ctx, syscall.SIGTERM)

	staus := <-exitStatus

	klog.Info(staus)
	return err
}

func (manager *containerManager) StartContainer(ctx context.Context, container *v1.Container) error {
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

func (manager *containerManager) StopContainer(ctx context.Context, container *v1.Container) error {
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

func (manager *containerManager) PauseContainer(ctx context.Context, container *v1.Container) error {
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

func (manager *containerManager) ResumeContainer(ctx context.Context, container *v1.Container) error {
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

func (manager *containerManager) RemoveContainer(ctx context.Context, container *v1.Container) error {
	manager.StopContainer(ctx, container)

	err := manager.client.ContainerService().Delete(ctx, container.ID)

	if err != nil {
		return err
	}

	return nil
}

func (manager *containerManager) ListContainers(ctx context.Context) ([]*v1.Container, error) {
	// do we really need this?

	return nil, nil
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
