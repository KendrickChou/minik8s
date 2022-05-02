/*
	文件包含所有http REST接口的handler
	所有接口供http_server使用，任何其他的外部调用都是不被允许的

	创建日期：4月15日
	修改日期：4月16日
	作者：蒋哲
*/
package apiserver

import (
	"encoding/json"
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
	} else if kv.Type == config.AS_OP_ERROR_String {
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
		c.JSON(200, gin.H{"status": "OK", "id": name})
	}
}

func handlePutService(c *gin.Context) {
	buf := make([]byte, c.Request.ContentLength)
	_, err := c.Request.Body.Read(buf)
	name := c.Param("name")

	if etcdTest("/minik8s/service/" + name) {
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

	if etcdTest("/minik8s/service/" + name) {
		c.JSON(404, gin.H{"status": "ERR", "error": "No such service"})
	} else {
		err := etcdDel("/minik8s/service/" + name)
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
			info, err := json.Marshal(kv)
			if err != nil {
				klog.Infof("json parse error, cancel watch task...\n")
				cancel()
				return
			}
			_, err = fmt.Fprintf(c.Writer, string(info))
			writeN, err := c.Writer.Write([]byte{26})
			fmt.Printf("wN: %v, err: %v", writeN, err)
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
	} else if kv.Type == config.AS_OP_ERROR_String {
		c.JSON(404, gin.H{"status": "ERR", "error": "No such service"})
	} else {
		wch, cancel := etcdWatch("/minik8s/service")
		flusher, _ := c.Writer.(http.Flusher)
		for {
			select {
			case <-c.Request.Context().Done():
				klog.Infof("connection closed, cancel watch task...\n")
				cancel()
				return
			case kv := <-wch:
				info, err := json.Marshal(kv)
				if err != nil {
					klog.Infof("json parse error, cancel watch task...\n")
					cancel()
					return
				}
				_, err = fmt.Fprintf(c.Writer, string(info))
				if err != nil {
					klog.Infof("fail to write to client, cancel watch task...\n")
					cancel()
					return
				}
				flusher.Flush()
			}
		}
	}
}

//------------ Pod Rest API -----------
func handleGetPods(c *gin.Context) {
	pods, _ := etcdGetPrefix("/minik8s/pod")
	c.JSON(200, gin.H{
		"status":   "OK",
		"pod_num":  len(pods),
		"services": pods})
}

func handleGetPod(c *gin.Context) {
	kv, err := etcdGet("/minik8s/pod/" + c.Param("name"))
	if err != nil {
		c.JSON(500, gin.H{"status": "ERR", "error": err.Error()})
	} else if kv.Type == config.AS_OP_ERROR_String {
		c.JSON(404, gin.H{"status": "ERR", "error": "No such pod"})
	} else {
		c.JSON(200, kv)
	}
}

func handlePostPod(c *gin.Context) {
	buf := make([]byte, c.Request.ContentLength)
	_, err := c.Request.Body.Read(buf)
	name := "P" + strconv.Itoa(objCount) + "-" + random.String(8)
	objCount++
	err = etcdPut("/minik8s/pod/"+name, string(buf))
	if err != nil {
		c.JSON(500, gin.H{"status": "ERR", "error": err.Error()})
	} else {
		c.JSON(200, gin.H{"status": "OK", "id": name})
	}
}

func handlePutPod(c *gin.Context) {
	buf := make([]byte, c.Request.ContentLength)
	_, err := c.Request.Body.Read(buf)
	name := c.Param("name")
	if !etcdTest("/minik8s/pod/" + name) {
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
	if !etcdTest("/minik8s/pod/" + name) {
		c.JSON(404, gin.H{"status": "ERR", "error": "No such pod"})
	} else {
		err := etcdDel("/minik8s/pod/" + name)
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
	for {
		select {
		case <-c.Request.Context().Done():
			klog.Infof("connection closed, cancel watch task...\n")
			cancel()
			return
		case kv := <-wch:
			info, err := json.Marshal(kv)
			if err != nil {
				klog.Infof("json parse error, cancel watch task...\n")
				cancel()
				return
			}
			_, err = fmt.Fprintf(c.Writer, string(info))
			_, err = c.Writer.Write([]byte{26})
			if err != nil {
				klog.Infof("fail to write to client, cancel watch task...\n")
				cancel()
				return
			}
			flusher.Flush()
		}
	}
}

func handleWatchPod(c *gin.Context) {
	name := c.Param("name")
	if !etcdTest("/minik8s/pod/" + name) {
		c.JSON(404, gin.H{"status": "ERR", "error": "No such pod"})
	} else {
		wch, cancel := etcdWatch("/minik8s/pod")
		flusher, _ := c.Writer.(http.Flusher)
		for {
			select {
			case <-c.Request.Context().Done():
				klog.Infof("connection closed, cancel watch task...\n")
				cancel()
				return
			case kv := <-wch:
				info, err := json.Marshal(kv)
				if err != nil {
					klog.Infof("json parse error, cancel watch task...\n")
					cancel()
					return
				}
				_, err = fmt.Fprintf(c.Writer, string(info))
				if err != nil {
					klog.Infof("fail to write to client, cancel watch task...\n")
					cancel()
					return
				}
				flusher.Flush()
			}
		}
	}
}

func handleGetPodsByNode(c *gin.Context) {
	nname := c.Param("nname")
	if !etcdTest("/minik8s/node/" + nname) {
		c.JSON(404, gin.H{"status": "ERR", "error": "No such node"})
	} else {
		pods, _ := etcdGetPrefix("/minik8s/node/" + nname + "/pod")
		c.JSON(200, gin.H{
			"status":  "OK",
			"pod_num": len(pods),
			"pods":    pods})
	}
}

func handleGetPodByNode(c *gin.Context) {
	nname := c.Param("nname")
	if !etcdTest("/minik8s/node/" + nname) {
		c.JSON(404, gin.H{"status": "ERR", "error": "No such node"})
	} else {
		pname := c.Param("pname")
		kv, err := etcdGet("/minik8s/node/" + nname + "/pod/" + pname)
		if err != nil {
			c.JSON(500, gin.H{"status": "ERR", "error": err.Error()})
		} else if kv.Type == config.AS_OP_ERROR_String {
			c.JSON(404, gin.H{"status": "ERR", "error": "No such pod"})
		} else {
			c.JSON(200, kv)
		}
	}
}

func handlePostPodByNode(c *gin.Context) {
	buf := make([]byte, c.Request.ContentLength)
	_, err := c.Request.Body.Read(buf)
	nname := c.Param("nname")
	pname := c.Param("pname")
	if !etcdTest("/minik8s/node/"+nname) || !etcdTest("/minik8s/pod/"+pname) {
		c.JSON(404, gin.H{"status": "ERR", "error": "No such pod"})
	} else {
		err = etcdPut("/minik8s/node/"+nname+"/pod/"+pname, string(buf))
		if err != nil {
			c.JSON(500, gin.H{"status": "ERR", "error": err.Error()})
		} else {
			c.JSON(200, gin.H{"status": "OK"})
		}
	}
}

func handlePutPodByNode(c *gin.Context) {
	buf := make([]byte, c.Request.ContentLength)
	_, err := c.Request.Body.Read(buf)
	nname := c.Param("nname")
	pname := c.Param("pname")
	if !etcdTest("/minik8s/node/"+nname) || !etcdTest("/minik8s/pod/"+pname) {
		c.JSON(404, gin.H{"status": "ERR", "error": "No such pod"})
	} else {
		err = etcdPut("/minik8s/node/"+nname+"/pod/"+pname, string(buf))
		if err != nil {
			c.JSON(500, gin.H{"status": "ERR", "error": err.Error()})
		} else {
			c.JSON(200, gin.H{"status": "OK"})
		}
	}
}

func handleDeletePodByNode(c *gin.Context) {
	nname := c.Param("nname")
	pname := c.Param("pname")
	if !etcdTest("/minik8s/node/"+nname) || !etcdTest("/minik8s/node/"+nname+"/pod/"+pname) {
		c.JSON(404, gin.H{"status": "ERR", "error": "No such pod"})
	} else {
		err := etcdDel("/minik8s/node/" + nname + "/pod/" + pname)
		if err != nil {
			c.JSON(500, gin.H{"status": "ERR", "error": err.Error()})
		} else {
			c.JSON(200, gin.H{"status": "OK"})
		}
	}
}

func handleWatchPodsByNode(c *gin.Context) {
	nname := c.Param("nname")
	wch, cancel := etcdWatchPrefix("/minik8s/node/" + nname + "/pod")
	flusher, _ := c.Writer.(http.Flusher)
	for {
		select {
		case <-c.Request.Context().Done():
			klog.Infof("connection closed, cancel watch task...\n")
			cancel()
			return
		case kv := <-wch:
			info, err := json.Marshal(kv)

			if err != nil {
				klog.Infof("json parse error, cancel watch task...\n")
				cancel()
				return
			}
			_, err = c.Writer.Write(append(info, 26))
			if err != nil {
				klog.Infof("fail to write to client, cancel watch task...\n")
				cancel()
				return
			}
			flusher.Flush()
		}
	}
}

//
//func handleWatchPod(c *gin.Context) {
//	name := c.Param("name")
//	kv, err := etcdGet("/minik8s/pod/" + name)
//	if err != nil {
//		c.JSON(500, gin.H{"status": "ERR", "error": err.Error()})
//	} else if kv.Type == config.AS_OP_ERROR_String {
//		c.JSON(404, gin.H{"status": "ERR", "error": "No such pod"})
//	} else {
//		wch, cancel := etcdWatch("/minik8s/pod")
//		flusher, _ := c.Writer.(http.Flusher)
//		for {
//			select {
//			case <-c.Request.Context().Done():
//				klog.Infof("connection closed, cancel watch task...\n")
//				cancel()
//				return
//			case kv := <-wch:
//				info, err := json.Marshal(kv)
//				if err != nil {
//					klog.Infof("json parse error, cancel watch task...\n")
//					cancel()
//					return
//				}
//				_, err = fmt.Fprintf(c.Writer, string(info))
//				if err != nil {
//					klog.Infof("fail to write to client, cancel watch task...\n")
//					cancel()
//					return
//				}
//				flusher.Flush()
//			}
//		}
//	}
//}

//------------ Replica Rest API -----------
func handleGetReplicas(c *gin.Context) {
	services, _ := etcdGetPrefix("/minik8s/replicas")
	c.JSON(200, gin.H{
		"status":      "OK",
		"replica_num": len(services),
		"replicas":    services})
}

func handleGetReplica(c *gin.Context) {
	kv, err := etcdGet("/minik8s/replica/" + c.Param("name"))
	if err != nil {
		c.JSON(500, gin.H{"status": "ERR", "error": err.Error()})
	} else if kv.Type == config.AS_OP_ERROR_String {
		c.JSON(404, gin.H{"status": "ERR", "error": "No such replica"})
	} else {
		c.JSON(200, kv)
	}
}

func handlePostReplica(c *gin.Context) {
	buf := make([]byte, c.Request.ContentLength)
	_, err := c.Request.Body.Read(buf)
	name := "R" + strconv.Itoa(objCount) + "-" + random.String(8)
	objCount++
	err = etcdPut("/minik8s/replica/"+name, string(buf))
	if err != nil {
		c.JSON(500, gin.H{"status": "ERR", "error": err.Error()})
	} else {
		c.JSON(200, gin.H{"status": "OK", "id": name})
	}
}

func handlePutReplica(c *gin.Context) {
	buf := make([]byte, c.Request.ContentLength)
	_, err := c.Request.Body.Read(buf)
	name := c.Param("name")
	if etcdTest("/minik8s/replica/" + name) {
		c.JSON(404, gin.H{"status": "ERR", "error": "No such replica"})
	} else {
		err = etcdPut("/minik8s/replica/"+name, string(buf))
		if err != nil {
			c.JSON(500, gin.H{"status": "ERR", "error": err.Error()})
		} else {
			c.JSON(200, gin.H{"status": "OK"})
		}
	}
}

func handleDeleteReplica(c *gin.Context) {
	name := c.Param("name")
	kv, err := etcdGet("/minik8s/replica/" + name)
	if err != nil {
		c.JSON(500, gin.H{"status": "ERR", "error": err.Error()})
	} else if kv.Type == config.AS_OP_ERROR_String {
		c.JSON(404, gin.H{"status": "ERR", "error": "No such replica"})
	} else {
		err = etcdDel("/minik8s/replica/" + name)
		if err != nil {
			c.JSON(500, gin.H{"status": "ERR", "error": err.Error()})
		} else {
			c.JSON(200, gin.H{"status": "OK"})
		}
	}
}

func handleWatchReplicas(c *gin.Context) {
	wch, cancel := etcdWatchPrefix("/minik8s/replica")
	flusher, _ := c.Writer.(http.Flusher)
	for {
		select {
		case <-c.Request.Context().Done():
			klog.Infof("connection closed, cancel watch task...\n")
			cancel()
			return
		case kv := <-wch:
			info, err := json.Marshal(kv)
			if err != nil {
				klog.Infof("json parse error, cancel watch task...\n")
				cancel()
				return
			}
			_, err = fmt.Fprintf(c.Writer, string(info))
			if err != nil {
				klog.Infof("fail to write to client, cancel watch task...\n")
				cancel()
				return
			}
			flusher.Flush()
		}
	}
}

func handleWatchReplica(c *gin.Context) {
	name := c.Param("name")
	kv, err := etcdGet("/minik8s/replica/" + name)
	if err != nil {
		c.JSON(500, gin.H{"status": "ERR", "error": err.Error()})
	} else if kv.Type == config.AS_OP_ERROR_String {
		c.JSON(404, gin.H{"status": "ERR", "error": "No such replica"})
	} else {
		wch, cancel := etcdWatch("/minik8s/replica")
		flusher, _ := c.Writer.(http.Flusher)
		for {
			select {
			case <-c.Request.Context().Done():
				klog.Infof("connection closed, cancel watch task...\n")
				cancel()
				return
			case kv := <-wch:
				info, err := json.Marshal(kv)
				if err != nil {
					klog.Infof("json parse error, cancel watch task...\n")
					cancel()
					return
				}
				_, err = fmt.Fprintf(c.Writer, string(info))
				if err != nil {
					klog.Infof("fail to write to client, cancel watch task...\n")
					cancel()
					return
				}
				flusher.Flush()
			}
		}
	}
}

//------------ Node Rest API -----------
func handleGetNodes(c *gin.Context) {
	services, _ := etcdGetPrefix("/minik8s/nodes")
	c.JSON(200, gin.H{
		"status":   "OK",
		"node_num": len(services),
		"nodes":    services})
}

func handleGetNode(c *gin.Context) {
	kv, err := etcdGet("/minik8s/node/" + c.Param("name"))
	if err != nil {
		c.JSON(500, gin.H{"status": "ERR", "error": err.Error()})
	} else if kv.Type == config.AS_OP_ERROR_String {
		c.JSON(404, gin.H{"status": "ERR", "error": "No such node"})
	} else {
		c.JSON(200, kv)
	}
}

func handlePostNode(c *gin.Context) {
	buf := make([]byte, c.Request.ContentLength)
	_, err := c.Request.Body.Read(buf)
	name := "N" + strconv.Itoa(objCount) + "-" + random.String(8)
	objCount++
	err = etcdPut("/minik8s/node/"+name, string(buf))
	if err != nil {
		c.JSON(500, gin.H{"status": "ERR", "error": err.Error()})
	} else {
		c.JSON(200, gin.H{"status": "OK", "id": name})
	}
}

func handlePutNode(c *gin.Context) {
	buf := make([]byte, c.Request.ContentLength)
	_, err := c.Request.Body.Read(buf)
	name := c.Param("name")
	kv, err := etcdGet("/minik8s/node/" + name)
	if err != nil {
		c.JSON(500, gin.H{"status": "ERR", "error": err.Error()})
	} else if kv.Type == config.AS_OP_ERROR_String {
		c.JSON(404, gin.H{"status": "ERR", "error": "No such node"})
	} else {
		err = etcdPut("/minik8s/node/"+name, string(buf))
		if err != nil {
			c.JSON(500, gin.H{"status": "ERR", "error": err.Error()})
		} else {
			c.JSON(200, gin.H{"status": "OK"})
		}
	}
}

func handleDeleteNode(c *gin.Context) {
	name := c.Param("name")
	kv, err := etcdGet("/minik8s/node/" + name)
	if err != nil {
		c.JSON(500, gin.H{"status": "ERR", "error": err.Error()})
	} else if kv.Type == config.AS_OP_ERROR_String {
		c.JSON(404, gin.H{"status": "ERR", "error": "No such replica"})
	} else {
		err = etcdDel("/minik8s/node/" + name)
		if err != nil {
			c.JSON(500, gin.H{"status": "ERR", "error": err.Error()})
		} else {
			c.JSON(200, gin.H{"status": "OK"})
		}
	}
}

//------------ Heartbeat API -----------

func handleHeartbeat(c *gin.Context) {
	//name := c.Param("name")
	//num := c.Param("num")
	//klog.Infof("heartbeat from %v, num %v\n", name, num)
	c.JSON(200, gin.H{"status": "OK"})
}
