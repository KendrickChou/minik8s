/*
	文件包含所有http REST接口的handler
	所有接口供http_server使用，任何其他的外部调用都是不被允许的

	创建日期：4月15日
	修改日期：4月16日
	作者：蒋哲
*/
package apiserver

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"k8s.io/klog"
	"minik8s.com/minik8s/config"
	"minik8s.com/minik8s/utils/random"
	"net/http"
	"strconv"
)

//------------ Service Rest API -----------
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
	name := "S" + strconv.Itoa(objCount) + "-" + random.String(8)
	objCount++
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
	wch, cancel := etcdWatchPrefix("/minik8s/service")
	flusher, _ := c.Writer.(http.Flusher)
	for {
		select {
		case <-c.Request.Context().Done():
			klog.Infof("connection closed, cancel watch task...\n")
			cancel()
			return
		case kv := <-wch:
			_, err := fmt.Fprintf(c.Writer, "watch result key: %v, value: %v\n", kv.Key, kv.Value)
			if err != nil {
				klog.Infof("fail to write to client, cancel watch task...\n")
				cancel()
				return
			}
			flusher.Flush()
		}
	}
}

func handleWatchService(c *gin.Context) {
	name := c.Param("name")
	kv, err := etcdGet("/minik8s/service/" + name)
	if err != nil {
		c.JSON(500, gin.H{"status": "ERR", "error": err.Error()})
	} else if kv.ty == config.AS_OP_ERROR_String {
		c.JSON(404, gin.H{"status": "ERR", "error": "No such service"})
	} else {
		wch, cancel := etcdWatch("/minik8s/service")
		flusher, _ := c.Writer.(http.Flusher)
		for kv := range wch {
			_, err := fmt.Fprintf(c.Writer, "watch result key: %v, value: %v\n", kv.Key, kv.Value)
			if err != nil {
				cancel()
				return
			}
			flusher.Flush()
		}
	}
}

//------------ Pod Rest API -----------
func handleGetPods(c *gin.Context) {
	services, _ := etcdGetPrefix("/minik8s/pod")
	c.JSON(200, gin.H{
		"status":   "OK",
		"pod_num":  len(services),
		"services": services})
}

func handleGetPod(c *gin.Context) {
	kv, err := etcdGet("/minik8s/pod/" + c.Param("name"))
	if err != nil {
		c.JSON(500, gin.H{"status": "ERR", "error": err.Error()})
	} else if kv.ty == config.AS_OP_ERROR_String {
		c.JSON(404, gin.H{"status": "ERR", "error": "No such service"})
	} else {
		c.JSON(200, kv)
	}
}

func handlePostPod(c *gin.Context) {
	buf := make([]byte, c.Request.ContentLength)
	_, err := c.Request.Body.Read(buf)
	name := "P" + strconv.Itoa(objCount) + "-" + random.String(8)
	objCount++
	err = etcdPut("/minik8s/service/"+name, string(buf))
	if err != nil {
		c.JSON(500, gin.H{"status": "ERR", "error": err.Error()})
	} else {
		c.JSON(200, gin.H{"status": "OK"})
	}
}

func handlePutPod(c *gin.Context) {
	buf := make([]byte, c.Request.ContentLength)
	_, err := c.Request.Body.Read(buf)
	name := c.Param("name")
	kv, err := etcdGet("/minik8s/pod/" + name)
	if err != nil {
		c.JSON(500, gin.H{"status": "ERR", "error": err.Error()})
	} else if kv.ty == config.AS_OP_ERROR_String {
		c.JSON(404, gin.H{"status": "ERR", "error": "No such pod"})
	} else {
		err = etcdPut("/minik8s/pod/"+name, string(buf))
		if err != nil {
			c.JSON(500, gin.H{"status": "ERR", "error": err.Error()})
		} else {
			c.JSON(200, gin.H{"status": "OK"})
		}
	}
}

func handleDeletePod(c *gin.Context) {
	name := c.Param("name")
	kv, err := etcdGet("/minik8s/pod/" + name)
	if err != nil {
		c.JSON(500, gin.H{"status": "ERR", "error": err.Error()})
	} else if kv.ty == config.AS_OP_ERROR_String {
		c.JSON(404, gin.H{"status": "ERR", "error": "No such service"})
	} else {
		err = etcdDel("/minik8s/pod/" + name)
		if err != nil {
			c.JSON(500, gin.H{"status": "ERR", "error": err.Error()})
		} else {
			c.JSON(200, gin.H{"status": "OK"})
		}
	}
}

func handleWatchPods(c *gin.Context) {
	wch, cancel := etcdWatchPrefix("/minik8s/pod")
	flusher, _ := c.Writer.(http.Flusher)
	for kv := range wch {
		_, err := fmt.Fprintf(c.Writer, "watch result key: %v, value: %v\n", kv.Key, kv.Value)
		if err != nil {
			cancel()
			return
		}
		flusher.Flush()
	}
}

func handleWatchPod(c *gin.Context) {
	name := c.Param("name")
	kv, err := etcdGet("/minik8s/pod/" + name)
	if err != nil {
		c.JSON(500, gin.H{"status": "ERR", "error": err.Error()})
	} else if kv.ty == config.AS_OP_ERROR_String {
		c.JSON(404, gin.H{"status": "ERR", "error": "No such service"})
	} else {
		wch, cancel := etcdWatch("/minik8s/pod")
		flusher, _ := c.Writer.(http.Flusher)
		for kv := range wch {
			_, err := fmt.Fprintf(c.Writer, "watch result key: %v, value: %v\n", kv.Key, kv.Value)
			if err != nil {
				cancel()
				return
			}
			flusher.Flush()
		}
	}
}
