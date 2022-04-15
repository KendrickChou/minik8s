package apiserver

import (
	"github.com/gin-gonic/gin"
	"k8s.io/klog"
	"minik8s.com/minik8s/config"
	"minik8s.com/minik8s/utils/random"
	"strconv"
)

var serviceCount int = 0

func runHttpServer() {
	gin.SetMode(gin.ReleaseMode)
	r := gin.Default()

	_ = r.SetTrustedProxies([]string{"127.0.0.1"})

	//a simple handler to test connection
	r.GET("/test", handleGetTest)

	//REST API
	r.GET("/services", handleGetServices)
	r.GET("/service/:name", handleGetService)
	r.POST("/service", handlePostService)
	r.PUT("/service/:name", handlePutService)
	r.DELETE("/service/:name", handleDeleteService)

	//Watch
	r.GET("/watch/services", handleWatchServices)
	r.GET("/watch/service/:name", handleWatchService)

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

func handleGetServices(c *gin.Context) {
	services, _ := etcdGetPrefix("/minik8s/service")
	c.JSON(200, gin.H{
		"status":      "OK",
		"service_num": len(services),
		"services":    services})
}

func handleGetService(c *gin.Context) {
	kv, err := etcdGet("/minik8s/service/" + c.Param("name"))
	if err != nil {
		c.JSON(500, gin.H{"status": "ERR", "error": err.Error()})
	} else if kv.ty == config.AS_OP_ERROR_String {
		c.JSON(404, gin.H{"status": "ERR", "error": "No such service"})
	} else {
		c.JSON(200, kv)
	}
}

func handlePostService(c *gin.Context) {
	buf := make([]byte, c.Request.ContentLength)
	_, err := c.Request.Body.Read(buf)
	name := "S" + strconv.Itoa(serviceCount) + "-" + random.RandomString(8)
	serviceCount++
	err = etcdPut("/minik8s/service/"+name, string(buf))
	if err != nil {
		c.JSON(500, gin.H{"status": "ERR", "error": err.Error()})
	} else {
		c.JSON(200, gin.H{"status": "OK"})
	}
}

func handlePutService(c *gin.Context) {
	buf := make([]byte, c.Request.ContentLength)
	_, err := c.Request.Body.Read(buf)
	name := c.Param("name")

	kv, err := etcdGet("/minik8s/service/" + name)
	if err != nil {
		c.JSON(500, gin.H{"status": "ERR", "error": err.Error()})
	} else if kv.ty == config.AS_OP_ERROR_String {
		c.JSON(404, gin.H{"status": "ERR", "error": "No such service"})
	} else {
		err = etcdPut("/minik8s/service/"+name, string(buf))
		if err != nil {
			c.JSON(500, gin.H{"status": "ERR", "error": err.Error()})
		} else {
			c.JSON(200, gin.H{"status": "OK"})
		}
	}
}

func handleDeleteService(c *gin.Context) {
	name := c.Param("name")

	kv, err := etcdGet("/minik8s/service/" + name)
	if err != nil {
		c.JSON(500, gin.H{"status": "ERR", "error": err.Error()})
	} else if kv.ty == config.AS_OP_ERROR_String {
		c.JSON(404, gin.H{"status": "ERR", "error": "No such service"})
	} else {
		err = etcdDel("/minik8s/service/" + name)
		if err != nil {
			c.JSON(500, gin.H{"status": "ERR", "error": err.Error()})
		} else {
			c.JSON(200, gin.H{"status": "OK"})
		}
	}
}

func handleWatchServices(c *gin.Context) {
}

func handleWatchService(c *gin.Context) {

}
