package constants

import (
	"strings"

	"k8s.io/klog"
	v1 "minik8s.com/minik8s/pkg/api/v1"
	"minik8s.com/minik8s/pkg/aqualake/apis/config"

	"github.com/dchest/uniuri"
)

func CouchPutDBRequest(name string) string {
	return config.CouchDBAddr + "/" + name
}

func CouchGetDocRequest(db, id string) string {
	return config.CouchDBAddr + "/" + db + "/" + id
}

func CouchGetAllDocsRequest(db string) string {
	return config.CouchDBAddr + "/" + db + "/_all_docs"
}

func CouchPutDocRequest(db, id string) string {
	return config.CouchDBAddr + "/" + db + "/" + id
}

func CouchDelDocRequest(db, id, rev string) string {
	return config.CouchDBAddr + "/" + db + "/" + id + "?rev=" + rev
}

func CouchPutFileRequest(db, docId, fileId, rev string) string {
	return config.CouchDBAddr + "/" + db + "/" + docId + "/" + fileId + "?rev=" + rev
}

func CouchGetFileRequest(db, docId, fileId string) string {
	return config.CouchDBAddr + "/" + db + "/" + docId + "/" + fileId
}

const (
	FunctionDBId string = "aqualake-function"
	ActionDBId   string = "aqualake-actionchain"
)

const (
	DefaultPoolSetSize = 3
)

const (
	PythonEnv string = "python"
	GoEnv     string = "go"

	PythonContainerName string = "python"
	GoContainerName     string = "go"

	PythonImageName string = "docker.io/kendrickchou/aqualake-python:latest"
	GoImageName     string = "docker.io/kendrickchou/aqualake-go:latest"
)

func NewPodConfig(podtype string) *v1.Pod {
	podname := "Aqualake-" + podtype + "-" + strings.ToUpper(uniuri.New())

	template := &v1.Pod{
		TypeMeta: v1.TypeMeta{
			Kind:       "pod",
			APIVersion: "v1",
		},
		ObjectMeta: v1.ObjectMeta{
			Name:      podname,
			Namespace: "default",
		},
		Spec: v1.PodSpec{
			Containers: []*v1.Container{},
		},
	}

	switch podtype {
	case PythonEnv:
		template.Spec.Containers = append(template.Spec.Containers,
			&v1.Container{
				Name:            PythonContainerName,
				Namespace:       "example",
				Image:           PythonImageName,
				ImagePullPolicy: "IfNotPresent",
			})
	case GoEnv:
		template.Spec.Containers = append(template.Spec.Containers,
			&v1.Container{
				Name:            GoContainerName,
				Namespace:       "example",
				Image:           GoImageName,
				ImagePullPolicy: "IfNotPresent",
			})
	default:
		klog.Errorf("Unsupported Pod Type %s", podtype)
	}

	return template
}

var SupportedEnvs []string = []string{PythonEnv, GoEnv}
