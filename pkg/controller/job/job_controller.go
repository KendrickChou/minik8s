package job

import (
	"k8s.io/klog"
	"minik8s.com/minik8s/config"
	v1 "minik8s.com/minik8s/pkg/api/v1"
	"minik8s.com/minik8s/pkg/controller/component"
	"strconv"
)

const (
	image           string = "xayahhh/minik8s_gpu_demo:v1.0"
	imagePullPolicy string = "IfNotPresent"
)

type JobController struct {
	podInformer *component.Informer
	jobInformer *component.Informer
	queue       component.WorkQueue
}

func NewJobController(podInf *component.Informer, jobInf *component.Informer) *JobController {
	return &JobController{
		podInformer: podInf,
		jobInformer: jobInf,
	}
}

func (jc *JobController) Run() {
	jc.queue.Init()

	jc.jobInformer.AddEventHandler(component.EventHandler{
		OnAdd:    jc.addJob,
		OnDelete: jc.deleteJob,
		OnUpdate: jc.updateJob,
	})

	go jc.worker()
}

func (jc *JobController) worker() {
	for jc.processNextWorkItem() {

	}
	klog.Error("Job Controller stopped")
}

func (jc *JobController) processNextWorkItem() bool {
	key := jc.queue.Fetch().(string)

	item := jc.jobInformer.GetItem(key)
	if item == nil {
		klog.Warningf("Job %s not found", key)
		return false
	}

	if !jc.queue.Process(key) {
		return true
	}
	job := item.(v1.GPUJob)

	err := jc.syncJob(&job)
	if err != nil {
		klog.Error(err.Error())
		return false
	}
	return true
}

func (jc *JobController) syncJob(job *v1.GPUJob) error {
	ownedPod := jc.getOwnedPod(job)

	if ownedPod == nil {
		containers := make([]*v1.Container, 1)
		containers[0] = &v1.Container{
			Name:            job.Name + "-container",
			Namespace:       job.Namespace,
			Image:           image,
			ImagePullPolicy: imagePullPolicy,
			Entrypoint: []string{
				"/bin/bash",
				"-c",
				"/apps/main " + config.AC_ServerAddr + ":" + strconv.Itoa(config.AC_ServerPort) + " " + job.UID,
			},
		}

		pod := v1.Pod{
			TypeMeta: v1.TypeMeta{
				Kind:       "pod",
				APIVersion: job.APIVersion,
			},
			ObjectMeta: v1.ObjectMeta{
				Name:      job.Name + "-pod",
				Namespace: job.Namespace,
				OwnerReferences: []v1.OwnerReference{
					{
						APIVersion: job.APIVersion,
						Name:       job.Name,
						Kind:       job.Kind,
						UID:        job.UID,
					},
				},
			},
			Spec: v1.PodSpec{
				Containers: containers,
			},
		}

		jc.podInformer.AddItem(pod)
	} else if job.Output != "" || job.Error != "" {
		jc.podInformer.DeleteItem(ownedPod.UID)
		jc.jobInformer.DeleteItem(job.UID)
	}

	return nil
}

func (jc *JobController) getOwnedPod(job *v1.GPUJob) *v1.Pod {
	pods := jc.podInformer.List()
	for _, item := range pods {
		pod := item.(v1.Pod)
		if v1.CheckOwner(pod.OwnerReferences, job.UID) >= 0 {
			return &pod
		}
	}

	return nil
}

func (jc *JobController) addJob(obj any) {
	job := obj.(v1.GPUJob)
	jc.queue.Push(job.UID)
}

func (jc *JobController) updateJob(newObj any, oldObj any) {
	newJob := newObj.(v1.GPUJob)

	if newJob.Error != "" || newJob.Output != "" {
		jc.queue.Push(newJob.UID)
	}
}

func (jc *JobController) deleteJob(obj any) {
	klog.Warningf("delete Job %v, this shouldn't happen", obj)
}
