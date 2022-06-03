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
	"github.com/gin-gonic/gin"
	"io"
	"k8s.io/klog"
	"minik8s.com/minik8s/config"
	v1 "minik8s.com/minik8s/pkg/api/v1"
	"minik8s.com/minik8s/utils/random"
	"net/http"
	"strconv"
)

//------------ Service Rest API -----------
func handleGetServices(c *gin.Context) {
	services, _ := etcdGetPrefix("/service")
	c.JSON(200, services)
}

func handleGetService(c *gin.Context) {
	kv, err := etcdGet("/service/" + c.Param("name"))
	if err != nil {
		c.JSON(500, gin.H{"status": "ERR", "error": err.Error()})
	} else if kv.Type == config.AS_OP_ERROR_String {
		c.JSON(404, gin.H{"status": "ERR", "error": "No such service"})
	} else {
		c.JSON(200, kv)
	}
}

func handlePostService(c *gin.Context) {
	buf, err := io.ReadAll(c.Request.Body)
	if err != nil {
		klog.Errorf("error: %v", err)
		c.JSON(500, gin.H{"status": "ERR", "error": err.Error()})
		return
	}
	name := "S" + strconv.Itoa(nextObjNum()) + "-" + random.String(8)
	var service v1.Service
	err = json.Unmarshal(buf, &service)
	if err != nil {
		c.JSON(500, gin.H{"status": "ERR", "error": err.Error()})
		return
	}
	service.UID = name
	buf, _ = json.Marshal(service)
	err = etcdPut("/service/"+name, string(buf))
	if err != nil {
		c.JSON(500, gin.H{"status": "ERR", "error": err.Error()})
	} else {
		c.JSON(200, gin.H{"status": "OK", "id": name})
	}
}

func handlePutService(c *gin.Context) {
	buf, err := io.ReadAll(c.Request.Body)
	if err != nil {
		klog.Errorf("error: %v", err)
		c.JSON(500, gin.H{"status": "ERR", "error": err.Error()})
		return
	}
	name := c.Param("name")

	if !etcdTest("/service/" + name) {
		c.JSON(404, gin.H{"status": "ERR", "error": "No such service"})
	} else {
		err := etcdPut(""+
			"/service/"+name, string(buf))
		if err != nil {
			c.JSON(500, gin.H{"status": "ERR", "error": err.Error()})
		} else {
			c.JSON(200, gin.H{"status": "OK"})
		}
	}
}

func handleDeleteService(c *gin.Context) {
	name := c.Param("name")

	if !etcdTest("/service/" + name) {
		c.JSON(404, gin.H{"status": "ERR", "error": "No such service"})
	} else {
		err := etcdDel("/service/" + name)
		if err != nil {
			c.JSON(500, gin.H{"status": "ERR", "error": err.Error()})
		} else {
			c.JSON(200, gin.H{"status": "OK"})
		}
	}
}

func handleWatchServices(c *gin.Context) {
	wch, cancel := etcdWatchPrefix("/service")
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
			c.Writer.Write(info)
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

func handleWatchService(c *gin.Context) {
	name := c.Param("name")
	if !etcdTest("/service/" + name) {
		c.JSON(404, gin.H{"status": "ERR", "error": "No such service"})
	} else {
		wch, cancel := etcdWatch("/service" + name)
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
				c.Writer.Write(info)
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
}

//------------ Pod Rest API -----------
func handleGetPods(c *gin.Context) {
	pods, _ := etcdGetPrefix("/pod")
	c.JSON(200, pods)
}

func handleGetPod(c *gin.Context) {
	kv, err := etcdGet("/pod/" + c.Param("name"))
	if err != nil {
		c.JSON(500, gin.H{"status": "ERR", "error": err.Error()})
	} else if kv.Type == config.AS_OP_ERROR_String {
		c.JSON(404, gin.H{"status": "ERR", "error": "No such pod"})
	} else {
		c.JSON(200, kv)
	}
}

func handlePostPod(c *gin.Context) {
	buf, err := io.ReadAll(c.Request.Body)
	if err != nil {
		klog.Errorf("error: %v", err)
		c.JSON(500, gin.H{"status": "ERR", "error": err.Error()})
		return
	}
	name := "P" + strconv.Itoa(nextObjNum()) + "-" + random.String(8)
	var pod v1.Pod
	err = json.Unmarshal(buf, &pod)
	if err != nil {
		c.JSON(500, gin.H{"status": "ERR", "error": err.Error()})
		return
	}
	pod.UID = name
	buf, _ = json.Marshal(pod)
	err = etcdPut("/pod/"+name, string(buf))
	if err != nil {
		c.JSON(500, gin.H{"status": "ERR", "error": err.Error()})
	} else {
		c.JSON(200, gin.H{"status": "OK", "id": name})
	}
}

func handlePutPod(c *gin.Context) {
	buf, err := io.ReadAll(c.Request.Body)
	if err != nil {
		klog.Errorf("error: %v", err)
		c.JSON(500, gin.H{"status": "ERR", "error": err.Error()})
		return
	}
	name := c.Param("name")
	if !etcdTest("/pod/" + name) {
		c.JSON(404, gin.H{"status": "ERR", "error": "No such pod"})
	} else {
		err := etcdPut("/pod/"+name, string(buf))
		if err != nil {
			c.JSON(500, gin.H{"status": "ERR", "error": err.Error()})
		} else {
			c.JSON(200, gin.H{"status": "OK"})
		}
	}
}

func handleDeletePod(c *gin.Context) {
	name := c.Param("name")
	if !etcdTest("/pod/" + name) {
		c.JSON(404, gin.H{"status": "ERR", "error": "No such pod"})
	} else {
		err := etcdDel("/pod/" + name)
		if err != nil {
			c.JSON(500, gin.H{"status": "ERR", "error": err.Error()})
		} else {
			c.JSON(200, gin.H{"status": "OK"})
		}
	}
}

func handleWatchPods(c *gin.Context) {
	wch, cancel := etcdWatchPrefix("/pod")
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
			c.Writer.Write(info)
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
	if !etcdTest("/pod/" + name) {
		c.JSON(404, gin.H{"status": "ERR", "error": "No such pod"})
	} else {
		wch, cancel := etcdWatch("/pod" + name)
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
				c.Writer.Write(info)
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
}

func handleGetPodsByNode(c *gin.Context) {
	nname := c.Param("nname")
	if !etcdTest("/node/" + nname) {
		c.JSON(404, gin.H{"status": "ERR", "error": "No such node"})
	} else {
		pods, _ := etcdGetPrefix("/innode/" + nname + "/pod/")
		c.JSON(200, pods)
	}
}

func handleGetPodByNode(c *gin.Context) {
	nname := c.Param("nname")
	if !etcdTest("/node/" + nname) {
		c.JSON(404, gin.H{"status": "ERR", "error": "No such node"})
	} else {
		pname := c.Param("pname")
		kv, err := etcdGet("/innode/" + nname + "/pod/" + pname)
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
	buf, err := io.ReadAll(c.Request.Body)
	if err != nil {
		klog.Errorf("error: %v", err)
		c.JSON(500, gin.H{"status": "ERR", "error": err.Error()})
		return
	}
	nname := c.Param("nname")
	pname := "P" + strconv.Itoa(nextObjNum()) + "-" + random.String(8)
	if !etcdTest("/node/" + nname) {
		c.JSON(404, gin.H{"status": "ERR", "error": "No such pod"})
	} else {
		err := etcdPut("/innode/"+nname+"/pod/"+pname, string(buf))
		if err != nil {
			c.JSON(500, gin.H{"status": "ERR", "error": err.Error()})
		} else {
			c.JSON(200, gin.H{"status": "OK", "id": pname})
		}
	}
}

func handlePutPodByNode(c *gin.Context) {
	buf, err := io.ReadAll(c.Request.Body)
	if err != nil {
		klog.Errorf("error: %v", err)
		c.JSON(500, gin.H{"status": "ERR", "error": err.Error()})
		return
	}
	nname := c.Param("nname")
	pname := c.Param("pname")
	if !etcdTest("/node/"+nname) || !etcdTest("/pod/"+pname) {
		c.JSON(404, gin.H{"status": "ERR", "error": "No such pod"})
	} else {
		err := etcdPut("/innode/"+nname+"/pod/"+pname, string(buf))
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
	if !etcdTest("/node/"+nname) || !etcdTest("/innode/"+nname+"/pod/"+pname) {
		c.JSON(404, gin.H{"status": "ERR", "error": "No such pod"})
	} else {
		err := etcdDel("/innode/" + nname + "/pod/" + pname)
		if err != nil {
			c.JSON(500, gin.H{"status": "ERR", "error": err.Error()})
		} else {
			c.JSON(200, gin.H{"status": "OK"})
		}
	}
}

func handleGetPodStatusesByNode(c *gin.Context) {
	nname := c.Param("nname")
	if !etcdTest("/node/" + nname) {
		c.JSON(404, gin.H{"status": "ERR", "error": "No such node"})
	} else {
		pods, _ := etcdGetPrefix("/innode/" + nname + "/podstatus/")
		c.JSON(200, pods)
	}
}

func handleGetPodStatusByNode(c *gin.Context) {
	nname := c.Param("nname")
	if !etcdTest("/node/" + nname) {
		c.JSON(404, gin.H{"status": "ERR", "error": "No such node"})
	} else {
		pname := c.Param("pname")
		kv, err := etcdGet("/innode/" + nname + "/podstatus/" + pname)
		if err != nil {
			c.JSON(500, gin.H{"status": "ERR", "error": err.Error()})
		} else if kv.Type == config.AS_OP_ERROR_String {
			c.JSON(404, gin.H{"status": "ERR", "error": "No such pod"})
		} else {
			c.JSON(200, kv)
		}
	}
}

func handlePutPodStatusByNode(c *gin.Context) {
	buf, err := io.ReadAll(c.Request.Body)
	if err != nil {
		klog.Errorf("error: %v", err)
		c.JSON(500, gin.H{"status": "ERR", "error": err.Error()})
		return
	}
	nname := c.Param("nname")
	pname := c.Param("pname")
	if !etcdTest("/node/"+nname) || !etcdTest("/pod/"+pname) {
		c.JSON(404, gin.H{"status": "ERR", "error": "No such pod"})
	} else {
		err := etcdPut("/innode/"+nname+"/podstatus/"+pname, string(buf))
		if err != nil {
			c.JSON(500, gin.H{"status": "ERR", "error": err.Error()})
		} else {
			var kv KV
			var pod v1.Pod
			var podStatus v1.PodStatus
			kv, err = etcdGet("/pod/" + pname)
			if err != nil {
				c.JSON(500, gin.H{"status": "ERR", "error": err.Error()})
				return
			}
			err = json.Unmarshal(kv.Value, &pod)
			if err != nil {
				c.JSON(500, gin.H{"status": "ERR", "error": err.Error()})
				return
			}
			err = json.Unmarshal(buf, &podStatus)
			if err != nil {
				c.JSON(500, gin.H{"status": "ERR", "error": err.Error()})
				return
			}
			pod.Status = podStatus
			podBuf, err := json.Marshal(pod)
			if err != nil {
				c.JSON(500, gin.H{"status": "ERR", "error": err.Error()})
				return
			}
			err = etcdPut("/pod/"+pname, string(podBuf))
			if err != nil {
				c.JSON(500, gin.H{"status": "ERR", "error": err.Error()})
				return
			}
			c.JSON(200, gin.H{"status": "OK"})
		}
	}
}

func handleDeletePodStatusByNode(c *gin.Context) {
	nname := c.Param("nname")
	pname := c.Param("pname")
	if !etcdTest("/node/"+nname) || !etcdTest("/innode/"+nname+"/podstatus/"+pname) {
		c.JSON(404, gin.H{"status": "ERR", "error": "No such pod"})
	} else {
		err := etcdDel("/innode/" + nname + "/podstatus/" + pname)
		if err != nil {
			c.JSON(500, gin.H{"status": "ERR", "error": err.Error()})
		} else {
			c.JSON(200, gin.H{"status": "OK"})
		}
	}
}

func handleWatchPodsByNode(c *gin.Context) {
	nname := c.Param("nname")
	wch, cancel := etcdWatchPrefix("/innode/" + nname + "/pod/")
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
			c.Writer.Write(info)
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

//------------ Replica Rest API -----------
func handleGetReplicas(c *gin.Context) {
	replicas, _ := etcdGetPrefix("/replica")
	c.JSON(200, replicas)
}

func handleGetReplica(c *gin.Context) {
	kv, err := etcdGet("/replica/" + c.Param("name"))
	if err != nil {
		c.JSON(500, gin.H{"status": "ERR", "error": err.Error()})
	} else if kv.Type == config.AS_OP_ERROR_String {
		c.JSON(404, gin.H{"status": "ERR", "error": "No such replica"})
	} else {
		c.JSON(200, kv)
	}
}

func handlePostReplica(c *gin.Context) {
	buf, err := io.ReadAll(c.Request.Body)
	if err != nil {
		klog.Errorf("error: %v", err)
		c.JSON(500, gin.H{"status": "ERR", "error": err.Error()})
		return
	}
	name := "R" + strconv.Itoa(nextObjNum()) + "-" + random.String(8)
	var rep v1.ReplicaSet
	err = json.Unmarshal(buf, &rep)
	if err != nil {
		c.JSON(500, gin.H{"status": "ERR", "error": err.Error()})
		return
	}
	rep.UID = name
	buf, _ = json.Marshal(rep)
	err = etcdPut("/replica/"+name, string(buf))
	if err != nil {
		c.JSON(500, gin.H{"status": "ERR", "error": err.Error()})
	} else {
		c.JSON(200, gin.H{"status": "OK", "id": name})
	}
}

func handlePutReplica(c *gin.Context) {
	buf, err := io.ReadAll(c.Request.Body)
	if err != nil {
		klog.Errorf("error: %v", err)
		c.JSON(500, gin.H{"status": "ERR", "error": err.Error()})
		return
	}
	name := c.Param("name")
	if !etcdTest("/replica/" + name) {
		c.JSON(404, gin.H{"status": "ERR", "error": "No such replica"})
	} else {
		err := etcdPut("/replica/"+name, string(buf))
		if err != nil {
			c.JSON(500, gin.H{"status": "ERR", "error": err.Error()})
		} else {
			c.JSON(200, gin.H{"status": "OK"})
		}
	}
}

func handleDeleteReplica(c *gin.Context) {
	name := c.Param("name")
	if !etcdTest("/replica/" + name) {
		c.JSON(404, gin.H{"status": "ERR", "error": "No such replica"})
	} else {
		err := etcdDel("/replica/" + name)
		if err != nil {
			c.JSON(500, gin.H{"status": "ERR", "error": err.Error()})
		} else {
			c.JSON(200, gin.H{"status": "OK"})
		}
	}
}

func handleWatchReplicas(c *gin.Context) {
	wch, cancel := etcdWatchPrefix("/replica")
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
			c.Writer.Write(info)
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

func handleWatchReplica(c *gin.Context) {
	name := c.Param("name")
	if !etcdTest("/replica/" + name) {
		c.JSON(404, gin.H{"status": "ERR", "error": "No such replica"})
	} else {
		wch, cancel := etcdWatch("/replica" + name)
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
				c.Writer.Write(info)
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
}

//------------ HPA Rest API -----------
func handleGetHPAs(c *gin.Context) {
	replicas, _ := etcdGetPrefix("/hpa")
	c.JSON(200, replicas)
}

func handleGetHPA(c *gin.Context) {
	kv, err := etcdGet("/hpa/" + c.Param("name"))
	if err != nil {
		c.JSON(500, gin.H{"status": "ERR", "error": err.Error()})
	} else if kv.Type == config.AS_OP_ERROR_String {
		c.JSON(404, gin.H{"status": "ERR", "error": "No such hpa"})
	} else {
		c.JSON(200, kv)
	}
}

func handlePostHPA(c *gin.Context) {
	buf, err := io.ReadAll(c.Request.Body)
	if err != nil {
		klog.Errorf("error: %v", err)
		c.JSON(500, gin.H{"status": "ERR", "error": err.Error()})
		return
	}
	name := "H" + strconv.Itoa(nextObjNum()) + "-" + random.String(8)
	var hpa v1.HorizontalPodAutoscaler
	err = json.Unmarshal(buf, &hpa)
	if err != nil {
		c.JSON(500, gin.H{"status": "ERR", "error": err.Error()})
		return
	}
	hpa.UID = name
	buf, _ = json.Marshal(hpa)
	err = etcdPut("/hpa/"+name, string(buf))
	if err != nil {
		c.JSON(500, gin.H{"status": "ERR", "error": err.Error()})
	} else {
		c.JSON(200, gin.H{"status": "OK", "id": name})
	}
}

func handlePutHPA(c *gin.Context) {
	buf, err := io.ReadAll(c.Request.Body)
	if err != nil {
		klog.Errorf("error: %v", err)
		c.JSON(500, gin.H{"status": "ERR", "error": err.Error()})
		return
	}
	name := c.Param("name")
	if !etcdTest("/hpa/" + name) {
		c.JSON(404, gin.H{"status": "ERR", "error": "No such hpa"})
	} else {
		err := etcdPut("/hpa/"+name, string(buf))
		if err != nil {
			c.JSON(500, gin.H{"status": "ERR", "error": err.Error()})
		} else {
			c.JSON(200, gin.H{"status": "OK"})
		}
	}
}

func handleDeleteHPA(c *gin.Context) {
	name := c.Param("name")
	if !etcdTest("/hpa/" + name) {
		c.JSON(404, gin.H{"status": "ERR", "error": "No such hpa"})
	} else {
		err := etcdDel("/hpa/" + name)
		if err != nil {
			c.JSON(500, gin.H{"status": "ERR", "error": err.Error()})
		} else {
			c.JSON(200, gin.H{"status": "OK"})
		}
	}
}

func handleWatchHPAs(c *gin.Context) {
	wch, cancel := etcdWatchPrefix("/hpa")
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
			c.Writer.Write(info)
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

func handleWatchHPA(c *gin.Context) {
	name := c.Param("name")
	if !etcdTest("/hpa/" + name) {
		c.JSON(404, gin.H{"status": "ERR", "error": "No such hpa"})
	} else {
		wch, cancel := etcdWatch("/hpa" + name)
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
				c.Writer.Write(info)
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
}

//------------ Endpoint Rest API -----------
func handleGetEndpoints(c *gin.Context) {
	replicas, _ := etcdGetPrefix("/endpoint")
	c.JSON(200, replicas)
}

func handleGetEndpoint(c *gin.Context) {
	kv, err := etcdGet("/endpoint/" + c.Param("name"))
	if err != nil {
		c.JSON(500, gin.H{"status": "ERR", "error": err.Error()})
	} else if kv.Type == config.AS_OP_ERROR_String {
		c.JSON(404, gin.H{"status": "ERR", "error": "No such endpoint"})
	} else {
		c.JSON(200, kv)
	}
}

func handlePostEndpoint(c *gin.Context) {
	buf, err := io.ReadAll(c.Request.Body)
	if err != nil {
		klog.Errorf("error: %v", err)
		c.JSON(500, gin.H{"status": "ERR", "error": err.Error()})
		return
	}
	name := "E" + strconv.Itoa(nextObjNum()) + "-" + random.String(8)
	var ep v1.Endpoint
	err = json.Unmarshal(buf, &ep)
	if err != nil {
		c.JSON(500, gin.H{"status": "ERR", "error": err.Error()})
		return
	}
	ep.UID = name
	buf, _ = json.Marshal(ep)
	err = etcdPut("/endpoint/"+name, string(buf))
	if err != nil {
		c.JSON(500, gin.H{"status": "ERR", "error": err.Error()})
	} else {
		c.JSON(200, gin.H{"status": "OK", "id": name})
	}
}

func handlePutEndpoint(c *gin.Context) {
	buf, err := io.ReadAll(c.Request.Body)
	if err != nil {
		klog.Errorf("error: %v", err)
		c.JSON(500, gin.H{"status": "ERR", "error": err.Error()})
		return
	}
	name := c.Param("name")
	if !etcdTest("/endpoint/" + name) {
		c.JSON(404, gin.H{"status": "ERR", "error": "No such endpoint"})
	} else {
		err := etcdPut("/endpoint/"+name, string(buf))
		if err != nil {
			c.JSON(500, gin.H{"status": "ERR", "error": err.Error()})
		} else {
			c.JSON(200, gin.H{"status": "OK"})
		}
	}
}

func handleDeleteEndpoint(c *gin.Context) {
	name := c.Param("name")
	if !etcdTest("/endpoint/" + name) {
		c.JSON(404, gin.H{"status": "ERR", "error": "No such endpoint"})
	} else {
		err := etcdDel("/endpoint/" + name)
		if err != nil {
			c.JSON(500, gin.H{"status": "ERR", "error": err.Error()})
		} else {
			c.JSON(200, gin.H{"status": "OK"})
		}
	}
}

func handleWatchEndpoints(c *gin.Context) {
	wch, cancel := etcdWatchPrefix("/endpoint")
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
			c.Writer.Write(info)
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

func handleWatchEndpoint(c *gin.Context) {
	name := c.Param("name")
	if !etcdTest("/endpoint/" + name) {
		c.JSON(404, gin.H{"status": "ERR", "error": "No such endpoint"})
	} else {
		wch, cancel := etcdWatch("/endpoint" + name)
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
				c.Writer.Write(info)
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
}

//------------ Endpoint Rest API -----------
func handleGetGPUs(c *gin.Context) {
	replicas, _ := etcdGetPrefix("/gpu/")
	c.JSON(200, replicas)
}

func handleGetGPU(c *gin.Context) {
	kv, err := etcdGet("/gpu/" + c.Param("name"))
	if err != nil {
		c.JSON(500, gin.H{"status": "ERR", "error": err.Error()})
	} else if kv.Type == config.AS_OP_ERROR_String {
		c.JSON(404, gin.H{"status": "ERR", "error": "No such gpu"})
	} else {
		c.JSON(200, kv)
	}
}

func handlePostGPU(c *gin.Context) {
	buf, err := io.ReadAll(c.Request.Body)
	if err != nil {
		klog.Errorf("error: %v", err)
		c.JSON(500, gin.H{"status": "ERR", "error": err.Error()})
		return
	}
	name := "G" + strconv.Itoa(nextObjNum()) + "-" + random.String(8)
	var gj v1.GPUJob
	err = json.Unmarshal(buf, &gj)
	if err != nil {
		c.JSON(500, gin.H{"status": "ERR", "error": err.Error()})
		return
	}
	gj.UID = name
	buf, _ = json.Marshal(gj)
	err = etcdPut("/gpu/"+name, string(buf))
	if err != nil {
		c.JSON(500, gin.H{"status": "ERR", "error": err.Error()})
	} else {
		c.JSON(200, gin.H{"status": "OK", "id": name})
	}
}

func handlePutGPU(c *gin.Context) {
	buf, err := io.ReadAll(c.Request.Body)
	if err != nil {
		klog.Errorf("error: %v", err)
		c.JSON(500, gin.H{"status": "ERR", "error": err.Error()})
		return
	}
	name := c.Param("name")
	if !etcdTest("/gpu/" + name) {
		c.JSON(404, gin.H{"status": "ERR", "error": "No such gpu job"})
	} else {
		err := etcdPut("/gpu/"+name, string(buf))
		if err != nil {
			c.JSON(500, gin.H{"status": "ERR", "error": err.Error()})
		} else {
			c.JSON(200, gin.H{"status": "OK"})
		}
	}
}

func handleDeleteGPU(c *gin.Context) {
	name := c.Param("name")
	if !etcdTest("/gpu/" + name) {
		c.JSON(404, gin.H{"status": "ERR", "error": "No such gpu"})
	} else {
		err := etcdDel("/gpu/" + name)
		if err != nil {
			c.JSON(500, gin.H{"status": "ERR", "error": err.Error()})
		} else {
			c.JSON(200, gin.H{"status": "OK"})
		}
	}
}

func handleWatchGPUs(c *gin.Context) {
	wch, cancel := etcdWatchPrefix("/gpu")
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
			c.Writer.Write(info)
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

//------------ Endpoint Rest API -----------
func handleGetDNSs(c *gin.Context) {
	replicas, _ := etcdGetPrefix("/dns")
	c.JSON(200, replicas)
}

func handleGetDNS(c *gin.Context) {
	kv, err := etcdGet("/dns/" + c.Param("name"))
	if err != nil {
		c.JSON(500, gin.H{"status": "ERR", "error": err.Error()})
	} else if kv.Type == config.AS_OP_ERROR_String {
		c.JSON(404, gin.H{"status": "ERR", "error": "No such dns"})
	} else {
		c.JSON(200, kv)
	}
}

func handlePostDNS(c *gin.Context) {
	buf, err := io.ReadAll(c.Request.Body)
	if err != nil {
		klog.Errorf("error: %v", err)
		c.JSON(500, gin.H{"status": "ERR", "error": err.Error()})
		return
	}
	name := "D" + strconv.Itoa(nextObjNum()) + "-" + random.String(8)
	var dns v1.DNS
	err = json.Unmarshal(buf, &dns)
	if err != nil {
		c.JSON(500, gin.H{"status": "ERR", "error": err.Error()})
		return
	}
	dns.UID = name
	buf, _ = json.Marshal(dns)
	err = etcdPut("/dns/"+name, string(buf))
	if err != nil {
		c.JSON(500, gin.H{"status": "ERR", "error": err.Error()})
	} else {
		c.JSON(200, gin.H{"status": "OK", "id": name})
	}
}

func handlePutDNS(c *gin.Context) {
	buf, err := io.ReadAll(c.Request.Body)
	if err != nil {
		klog.Errorf("error: %v", err)
		c.JSON(500, gin.H{"status": "ERR", "error": err.Error()})
		return
	}
	name := c.Param("name")
	if etcdTest("/dns/" + name) {
		c.JSON(404, gin.H{"status": "ERR", "error": "No such dns"})
	} else {
		err := etcdPut("/dns/"+name, string(buf))
		if err != nil {
			c.JSON(500, gin.H{"status": "ERR", "error": err.Error()})
		} else {
			c.JSON(200, gin.H{"status": "OK"})
		}
	}
}

func handleDeleteDNS(c *gin.Context) {
	name := c.Param("name")
	if !etcdTest("/dns/" + name) {
		c.JSON(404, gin.H{"status": "ERR", "error": "No such dns"})
	} else {
		err := etcdDel("/dns/" + name)
		if err != nil {
			c.JSON(500, gin.H{"status": "ERR", "error": err.Error()})
		} else {
			c.JSON(200, gin.H{"status": "OK"})
		}
	}
}

func handleWatchDNSs(c *gin.Context) {
	wch, cancel := etcdWatchPrefix("/dns")
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
			c.Writer.Write(info)
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

//------------ Node Rest API -----------
func handleGetNodes(c *gin.Context) {
	nodes, _ := etcdGetPrefix("/node")
	c.JSON(200, nodes)
}

func handleGetNode(c *gin.Context) {
	kv, err := etcdGet("/node/" + c.Param("name"))
	if err != nil {
		c.JSON(500, gin.H{"status": "ERR", "error": err.Error()})
	} else if kv.Type == config.AS_OP_ERROR_String {
		c.JSON(404, gin.H{"status": "ERR", "error": "No such node"})
	} else {
		c.JSON(200, kv)
	}
}

func handlePostNode(c *gin.Context) {
	_, err := io.ReadAll(c.Request.Body)
	if err != nil {
		klog.Errorf("error: %v", err)
		c.JSON(500, gin.H{"status": "ERR", "error": err.Error()})
		return
	}
	name := "N" + strconv.Itoa(nextObjNum()) + "-" + random.String(8)
	node := v1.Node{}
	//err = json.Unmarshal(buf, &node)
	//if err != nil {
	//	c.JSON(500, gin.H{"status": "ERR", "error": err.Error()})
	//	return
	//}
	node.UID = name
	node.Name = name
	buf, _ := json.Marshal(node)
	err = etcdPut("/node/"+name, string(buf))
	if err != nil {
		c.JSON(500, gin.H{"status": "ERR", "error": err.Error()})
	} else {
		c.JSON(200, gin.H{"status": "OK", "id": name})
	}
}

func handlePutNode(c *gin.Context) {
	buf, err := io.ReadAll(c.Request.Body)
	if err != nil {
		klog.Errorf("error: %v", err)
		c.JSON(500, gin.H{"status": "ERR", "error": err.Error()})
		return
	}
	name := c.Param("name")
	kv, err := etcdGet("/node/" + name)
	if err != nil {
		c.JSON(500, gin.H{"status": "ERR", "error": err.Error()})
	} else if kv.Type == config.AS_OP_ERROR_String {
		c.JSON(404, gin.H{"status": "ERR", "error": "No such node"})
	} else {
		err = etcdPut("/node/"+name, string(buf))
		if err != nil {
			c.JSON(500, gin.H{"status": "ERR", "error": err.Error()})
		} else {
			c.JSON(200, gin.H{"status": "OK"})
		}
	}
}

func handleDeleteNode(c *gin.Context) {
	name := c.Param("name")
	if !etcdTest("/node/" + name) {
		c.JSON(404, gin.H{"status": "ERR", "error": "No such node"})
	} else {
		err := etcdDel("/node/" + name)
		if err != nil {
			c.JSON(500, gin.H{"status": "ERR", "error": err.Error()})
		} else {
			c.JSON(200, gin.H{"status": "OK"})
		}
	}
}

func handleWatchNodes(c *gin.Context) {
	wch, cancel := etcdWatchPrefix("/node")
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
			c.Writer.Write(info)
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

func handleWatchNode(c *gin.Context) {
	name := c.Param("name")
	if !etcdTest("/node/" + name) {
		c.JSON(404, gin.H{"status": "ERR", "error": "No such node"})
	} else {
		wch, cancel := etcdWatch("/node/" + name)
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
				c.Writer.Write(info)
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
}

//------------ Other API -----------

func handleHeartbeat(c *gin.Context) {
	c.JSON(200, gin.H{"status": "OK"})
}

func handleDeleteAll(c *gin.Context) {
	err := etcdDelPrefix("/")
	if err != nil {
		c.JSON(500, gin.H{"status": "ERR", "error": err.Error()})
	} else {
		c.JSON(200, gin.H{"status": "OK"})
	}
}
