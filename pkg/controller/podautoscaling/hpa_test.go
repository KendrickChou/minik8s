package podautoscaling

import (
	"encoding/json"
	"io/ioutil"
	v1 "minik8s.com/minik8s/pkg/api/v1"
	"minik8s.com/minik8s/pkg/apiclient"
	"minik8s.com/minik8s/pkg/controller/component"
	rs "minik8s.com/minik8s/pkg/controller/replicaset"
	"os"
	"testing"
	"time"
)

func runInformerAndController() (podInformer *component.Informer, rsInformer *component.Informer, hpaInformer *component.Informer) {
	podInformer = component.NewInformer("Pod")
	podStopChan := make(chan bool)
	go podInformer.Run(podStopChan)

	rsInformer = component.NewInformer("ReplicaSet")
	rsStopChan := make(chan bool)
	go rsInformer.Run(rsStopChan)

	hpaInformer = component.NewInformer("HorizontalPodAutoscaler")
	hpStopChan := make(chan bool)
	go hpaInformer.Run(hpStopChan)

	rsController := rs.NewReplicaSetController(podInformer, rsInformer)
	go rsController.Run()

	hpaController := NewHorizontalController(hpaInformer, podInformer, rsInformer)
	go hpaController.Run()

	return podInformer, rsInformer, hpaInformer
}

func readFile(path string) []byte {
	file, err := os.Open(path)
	if err != nil {
		panic(err)
	}
	defer func(file *os.File) {
		closeErr := file.Close()
		if closeErr != nil {
			panic(closeErr)
		}
	}(file)
	content, err := ioutil.ReadAll(file)
	return content
}

func rest(id string, value string, objType apiclient.ObjType, opType apiclient.OpType) string {
	responseBytes := apiclient.Rest(id, value, objType, opType)
	var responseBody apiclient.HttpResponse
	err := json.Unmarshal(responseBytes, &responseBody)
	if err != nil {
		return ""
	}

	if responseBody.Status == "OK" {
		return responseBody.ID
	} else {
		return ""
	}
}

func podsOwnedByRS(podInformer *component.Informer, rs *v1.ReplicaSet) []v1.Pod {
	pods := podInformer.List()
	result := make([]v1.Pod, 0)
	for _, item := range pods {
		pod := item.(v1.Pod)
		if v1.GetOwnerReplicaSet(&pod) == rs.UID {
			result = append(result, pod)
		}
	}
	return result
}

func TestHPAIncreaseReplica(t *testing.T) {
	podInformer, rsInformer, hpaInformer := runInformerAndController()
	time.Sleep(time.Second)

	// create a replicaSet
	RSExample := readFile("./testExamples/default_replicaSet.json")
	rsId := rest("", string(RSExample), apiclient.OBJ_REPLICAS, apiclient.OP_POST)
	time.Sleep(time.Second)

	// wait until observe the rs
	rsItem := rsInformer.GetItem(rsId)
	if rsItem == nil {
		t.Fatalf("can't find ReplicaSet %s", rsId)
	}
	rs1 := rsItem.(v1.ReplicaSet)

	time.Sleep(time.Second)
	// should observe 3 pods owned by the rs
	podsOwnedByRS1 := podsOwnedByRS(podInformer, &rs1)
	t.Log(podsOwnedByRS1)
	if len(podsOwnedByRS1) != rs1.Spec.Replicas {
		t.Fatal("not enough Pod for ReplicaSet")
	}

	t.Run("podCpuIncrease", func(t *testing.T) {
		// create a hpa: cpu, averageUtilization 50%
		HPAExample := readFile("./testExamples/hpa_cpu_increase.json")
		hpaId := rest("", string(HPAExample), apiclient.OBJ_HPA, apiclient.OP_POST)

		time.Sleep(time.Second)

		hpaItem := hpaInformer.GetItem(hpaId)
		if hpaItem == nil {
			t.Errorf("can't observe HPA %s in hpaInformer", hpaId)
		}
		hpa1 := hpaItem.(v1.HorizontalPodAutoscaler)

		// the pods created by ReplicaSet have no status, so HPA decrease the num to "minReplicas"
		time.Sleep(time.Second * 3)

		podsCreated := podsOwnedByRS(podInformer, &rs1)
		if len(podsCreated) != 1 {
			t.Errorf("After create HPA, there are %d pods existing, minReplicas is %d", len(podsCreated), hpa1.Spec.MinReplicas)
		}

		containerState := v1.ContainerState{
			CPUPerc:  "50%",
			Status:   "running",
			ExitCode: 0,
		}

		// set all the pods' cpuperc to 50%
		for _, pod := range podsCreated {
			pod.Status.ContainerStatuses = make([]v1.ContainerStatus, 1)
			pod.Status.ContainerStatuses[0] = v1.ContainerStatus{
				Name:  "cState",
				State: containerState,
			}
			if !apiclient.UpdatePod(&pod) {
				t.Errorf("update Pod %s failed", pod.UID)
			}
		}

		// the pods add by HPA have 0 cpuperc, and the default scalingUp rule's default
		// stableWindowSeconds is 0, so HPA will keep adding pods
		t.Logf("replicas: %v", len(podsOwnedByRS(podInformer, &rs1)))

		// delete HPA
		rest(hpaId, "", apiclient.OBJ_HPA, apiclient.OP_DELETE)
		time.Sleep(time.Second)

		rsPods := podsOwnedByRS(podInformer, &rs1)
		if len(rsPods) != rs1.Spec.Replicas {
			t.Fatalf("after delete HPA, should have three pods, get: %v", len(rsPods))
		}
	})

	podsToDelete := podsOwnedByRS(podInformer, &rs1)
	rest(rs1.UID, "", apiclient.OBJ_REPLICAS, apiclient.OP_DELETE)

	for _, pod := range podsToDelete {
		rest(pod.UID, "", apiclient.OBJ_POD, apiclient.OP_DELETE)
	}
}
