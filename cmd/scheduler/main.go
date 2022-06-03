package main

import "minik8s.com/minik8s/pkg/scheduler"

func main() {
	scheduler.Init()
	scheduler.Run()
}
