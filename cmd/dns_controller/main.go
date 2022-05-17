package main

import "minik8s.com/minik8s/pkg/controller/dns_controller"

func main() {
	dns_controller.Init()
	dns_controller.Run()
}
