package main

import (
	"github.com/gin-gonic/gin"
	"k8s.io/klog"
	"minik8s.com/minik8s/utils/random"
)

var id string

func runHttpServer() {
	gin.SetMode(gin.ReleaseMode)
	r := gin.Default()
	_ = r.SetTrustedProxies([]string{"127.0.0.1"})

	//a simple handler to test connection
	r.GET("/hello", handleGetTest)

	err := r.Run(":80")
	if err != nil {
		klog.Errorf("gin server failed to start, err: %v", err)
	}
}

func handleGetTest(c *gin.Context){
	c.String(200, "server_id: " + id + ", message: hello, world!\n")
}

func main(){
	random.Init()
	id = random.String(4)
	runHttpServer()
}