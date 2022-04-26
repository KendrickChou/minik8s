package server

import (
	"bufio"
	"io"
	"net/http"

	"github.com/gin-gonic/gin"
	v1 "minik8s.com/minik8s/pkg/api/v1"
	"minik8s.com/minik8s/pkg/kubelet/apis/config"
)

type HandlerInterface interface {
	GetPods() ([]v1.Pod, error)
	GetPodByUID(UID string) (v1.Pod, error)

	CreatePod(pod v1.Pod) (v1.Pod, error)
	DeletePod(UID string) error
}

func InstallDefaultHandlers(server *gin.Engine, handler HandlerInterface) {
	server.GET("/pods", func(ctx *gin.Context) {
		getAllPods(ctx, handler)
	})

	server.GET("/pods/:UID", func(ctx *gin.Context) {
		getPodByUID(ctx, handler)
	})

	server.PUT("/pods", func(ctx *gin.Context) {
		createPod(ctx, handler)
	})

	server.DELETE("/pods/:UID", func(ctx *gin.Context) {
		deletePod(ctx, handler)
	})

	server.GET("/watch/:UID",func(ctx *gin.Context) {
		watchPod(ctx, handler)
	})
}

func getAllPods(ctx *gin.Context, handler HandlerInterface) {
	pods, err := handler.GetPods()

	if err != nil {
		ctx.IndentedJSON(http.StatusNotFound, gin.H{"message": err.Error()})
		return
	}

	ctx.IndentedJSON(http.StatusOK, pods)
}

func getPodByUID(ctx *gin.Context, handler HandlerInterface) {
	pod, err := handler.GetPodByUID(ctx.Param("UID"))

	if err != nil {
		ctx.IndentedJSON(http.StatusNotFound, gin.H{"message": err.Error()})
		return
	}

	ctx.IndentedJSON(http.StatusOK, pod)
}

func createPod(ctx *gin.Context, handler HandlerInterface) {
	var pod v1.Pod

	if err := ctx.BindJSON(&pod); err != nil {
		ctx.IndentedJSON(http.StatusBadRequest, gin.H{"message": "illegal request body"})
		return
	}

	pod, err := handler.CreatePod(pod)

	if err != nil {
		ctx.IndentedJSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}

	ctx.IndentedJSON(http.StatusCreated, pod)
}

func deletePod(ctx *gin.Context, handler HandlerInterface) {
	uid := ctx.Param("UID")
	err := handler.DeletePod(uid)

	if err != nil {
		ctx.IndentedJSON(http.StatusNotFound, gin.H{"message": err})
		return
	}

	ctx.IndentedJSON(http.StatusOK, uid)
}

func watchPod(ctx *gin.Context, handler HandlerInterface) {
	uid := ctx.Param("UID")

	resp, err := http.Get(config.ApiServerAddress + "/watch/pod/" + uid)

	if err != nil {
		ctx.IndentedJSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}

	go watchPodHandler(resp)

	ctx.IndentedJSON(http.StatusOK, uid)
}

func watchPodHandler(resp *http.Response) {
	defer resp.Body.Close()
	io.PipeRea
	reader := bufio.NewReader(resp.Body)
	for {
		pod, err := reader.ReadSt
	}
}
