package apiserver

import (
	"github.com/gin-gonic/gin"
	"k8s.io/klog"
	"minik8s.com/minik8s/config"
	"strconv"
)

func runHttpServer() {
	gin.SetMode(gin.ReleaseMode)
	r := gin.Default()

	_ = r.SetTrustedProxies([]string{"127.0.0.1"})

	r.GET("/test", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "hello world",
		})
	})

	err := r.Run(":" + strconv.Itoa(config.ASConfig.HttpListenPort))
	if err != nil {
		klog.Errorf("gin server failed to start, err: %v", err)
	}
}
