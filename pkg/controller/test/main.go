package test

import v1 "minik8s.com/minik8s/pkg/api/v1"

func main() {

}

func createService() {
	var service v1.Service
	service.APIVersion = "v1"
	service.Kind = "Service"
	service.Name = "test"
	service.Spec = v1.ServiceSpec{}
}
