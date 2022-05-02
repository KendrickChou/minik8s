package apiserver

import (
	"github.com/gin-gonic/gin"
	"k8s.io/klog"
	"minik8s.com/minik8s/config"
	"strconv"
)

var objCount int = 0

func runHttpServer() {
	gin.SetMode(gin.ReleaseMode)
	r := gin.Default()

	_ = r.SetTrustedProxies([]string{"127.0.0.1"})

	//a simple handler to test connection
	r.GET("/test", handleGetTest)

	//------------------ REST API ----------------------
	// service
	r.GET("/services", handleGetServices)
	r.GET("/service/:name", handleGetService)
	r.POST("/service", handlePostService)
	r.PUT("/service/:name", handlePutService)
	r.DELETE("/service/:name", handleDeleteService)

	// pod
	r.GET("/pods", handleGetPods)
	r.GET("/pod/:name", handleGetPod)
	r.POST("/pod", handlePostPod)
	r.PUT("/pod/:name", handlePutPod)
	r.DELETE("/pod/:name", handleDeletePod)

	// pod in certain namespace
	r.GET("/innode/:nname/pods", handleGetPodsByNode)
	r.GET("/innode/:nname/pod/:pname", handleGetPodByNode)
	r.POST("/innode/:nname/pod", handlePostPodByNode)
	r.PUT("/innode/:nname/pod/:pname", handlePutPodByNode)
	r.DELETE("/innode/:nname/pod/:pname", handleDeletePodByNode)

	// replica
	r.GET("/replicas", handleGetReplicas)
	r.GET("/replica/:name", handleGetReplica)
	r.POST("/replica", handlePostReplica)
	r.PUT("/replica/:name", handlePutReplica)
	r.DELETE("/replica/:name", handleDeleteReplica)

	// node
	r.GET("/nodes", handleGetNodes)
	r.GET("/node/:name/", handleGetNode)
	r.POST("/node", handlePostNode)
	r.PUT("/node/:name", handlePutNode)
	r.DELETE("/node/:name", handleDeleteNode)

	//------------------ WATCH API ----------------------
	r.GET("/watch/services", handleWatchServices)
	r.GET("/watch/service/:name", handleWatchService)

	r.GET("/watch/pods", handleWatchPods)
	r.GET("/watch/pod/:name", handleWatchPod)

	r.GET("/watch/innode/:nname/pods", handleWatchPodsByNode)

	r.GET("/watch/replicas", handleWatchReplicas)
	r.GET("/watch/replica/:name", handleWatchReplica)

	//------------------ HEARTBEAT -----------------------
	r.GET("/heartbeat/:name/:num", handleHeartbeat)

	err := r.Run(":" + strconv.Itoa(config.AS_HttpListenPort))
	if err != nil {
		klog.Errorf("gin server failed to start, err: %v", err)
	}
}

func handleGetTest(c *gin.Context) {
	c.JSON(200, gin.H{
		"message": "hello world",
	})
}
