package server

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"minik8s.com/minik8s/pkg/api/v1"
)

type HandlerInterface interface {
	GetPods() ([]*v1.Pod, error)
	GetPodByUID(UID v1.UID) (v1.Pod, error)

	CreatePod(pod v1.Pod) (v1.Pod, error)
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
}

func getAllPods(ctx *gin.Context, handler HandlerInterface) {
	pods, err := handler.GetPods()

	if err != nil {
		ctx.IndentedJSON(http.StatusNotFound, gin.H{"message": err})
		return
	}

	ctx.IndentedJSON(http.StatusOK, pods)
}

func getPodByUID(ctx *gin.Context, handler HandlerInterface) {
	pod, err := handler.GetPodByUID(v1.UID(ctx.Param("UID")))

	if err != nil {
		ctx.IndentedJSON(http.StatusNotFound, gin.H{"message": err})
		return
	}

	ctx.IndentedJSON(http.StatusOK, pod)
}

func createPod(ctx *gin.Context, handler HandlerInterface){
	var pod v1.Pod

	if err := ctx.BindJSON(&pod); err != nil {
		ctx.IndentedJSON(http.StatusBadRequest, gin.H{"message": "illegal request body"})
		return
	}

	pod, err := handler.CreatePod(pod)

	if (err != nil) {
		ctx.IndentedJSON(http.StatusBadRequest, gin.H{"message": err})
		return
	}

	ctx.IndentedJSON(http.StatusCreated, pod)
}
