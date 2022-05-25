package apiserver

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"k8s.io/klog"
	"log"
	"minik8s.com/minik8s/config"
	"net/http"
	"strconv"
)

var objCount int = 1000

func nextObjNum() int {
	objCount++
	if objCount > 9999 {
		objCount = 1000
	}
	return objCount
}

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

	r.GET("/innode/:nname/podstatuses", handleGetPodStatusesByNode)
	r.GET("/innode/:nname/podstatus/:pname", handleGetPodStatusByNode)
	r.PUT("/innode/:nname/podstatus/:pname", handlePutPodStatusByNode)
	r.DELETE("/innode/:nname/podstatus/:pname", handleDeletePodStatusByNode)

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

	// endpoint
	r.GET("/endpoints", handleGetEndpoints)
	r.GET("/endpoint/:name", handleGetEndpoint)
	r.POST("/endpoint", handlePostEndpoint)
	r.PUT("/endpoint/:name", handlePutEndpoint)
	r.DELETE("/endpoint/:name", handleDeleteEndpoint)

	// dns
	r.GET("/dnss", handleGetDNSs)
	r.GET("/dns/:name/", handleGetDNS)
	r.POST("/dns", handlePostDNS)
	r.PUT("/dns/:name", handlePutDNS)
	r.DELETE("/dns/:name", handleDeleteDNS)

	// gpu
	r.GET("/gpus", handleGetGPUs)
	r.GET("/gpu/:name/", handleGetGPU)
	r.POST("/gpu", handlePostGPU)
	r.PUT("/gpu/:name", handlePutGPU)
	r.DELETE("/gpu/:name", handleDeleteGPU)

	//clear all
	r.DELETE("/", handleDeleteAll)

	//------------------ WATCH API ----------------------
	r.GET("/watch/services", handleWatchServices)
	r.GET("/watch/service/:name", handleWatchService)

	r.GET("/watch/endpoints", handleWatchEndpoints)
	r.GET("/watch/endpoint/:name", handleWatchEndpoint)

	r.GET("/watch/dnss", handleWatchDNSs)
	r.GET("/watch/gpus", handleWatchGPUs)

	r.GET("/watch/pods", handleWatchPods)
	r.GET("/watch/pod/:name", handleWatchPod)

	r.GET("/watch/nodes", handleWatchNodes)
	r.GET("/watch/node/:name", handleWatchNode)

	r.GET("/watch/innode/:nname/pods", handleWatchPodsByNode)

	r.GET("/watch/replicas", handleWatchReplicas)
	r.GET("/watch/replica/:name", handleWatchReplica)

	//------------------ HEARTBEAT -----------------------
	r.GET("/heartbeat/:name/:num", handleHeartbeat)

	r.POST("/upload", func(c *gin.Context) {
		file, _ := c.FormFile("file")
		log.Println(file.Filename)
		dst := "./uploads/" + file.Filename
		err := c.SaveUploadedFile(file, dst)
		if err != nil {
			c.String(http.StatusOK, fmt.Sprintf("error: %v", err.Error()))
			return
		}
		c.String(http.StatusOK, fmt.Sprintf("'%s' uploaded!", file.Filename))
	})

	r.Static("/public", "./uploads")

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
