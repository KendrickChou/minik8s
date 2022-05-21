package main

import (
	"minik8s.com/minik8s/pkg/aqualake/apis/config"
	"minik8s.com/minik8s/pkg/aqualake/controller"
)

func main() {
	r := controller.SetUpRouter()

	r.Run(config.ServeAddr + ":8699")
}
