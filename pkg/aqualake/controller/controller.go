package controller

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"net/http"

	"github.com/gin-gonic/gin"
	"minik8s.com/minik8s/pkg/aqualake/apis/constants"
	"minik8s.com/minik8s/pkg/aqualake/apis/function"
	"minik8s.com/minik8s/pkg/aqualake/couchdb"
)

// need to do some init in a script(maybe a set env script)

func SetUpRouter() *gin.Engine {
	router := gin.Default()

	router.GET("/function/:id", func(ctx *gin.Context) {
		funcId := ctx.Params.ByName("id")
		file, err := couchdb.GetFile(context.TODO(), constants.FunctionDBId, funcId, funcId)

		if err != nil {
			ctx.JSON(http.StatusOK, gin.H{"error": err.Error()})
		}else {
			ctx.JSON(http.StatusOK, gin.H{"function": file})
		}
	})

	router.PUT("/function/:id", func(ctx *gin.Context) {
		funcId := ctx.Params.ByName("id")
		rev, err := couchdb.PutDoc(context.TODO(), constants.FunctionDBId, funcId, "{}")

		if err != nil {
			ctx.JSON(http.StatusOK, gin.H{"error": err.Error()})
			return
		}

		buf, err := ioutil.ReadAll(ctx.Request.Body)

		if err != nil {
			ctx.JSON(http.StatusOK, gin.H{"error": err.Error()})
			return
		}

		err = couchdb.PutFile(context.TODO(), constants.FunctionDBId, funcId, funcId, rev, string(buf))

		if err != nil {
			ctx.JSON(http.StatusOK, gin.H{"error": err.Error()})
		} else {
			ctx.JSON(http.StatusOK, gin.H{"ok": true})
		}
	})

	router.DELETE("/function/:id", func(ctx *gin.Context) {
		funcId := ctx.Params.ByName("id")
		bytes, err := couchdb.GetDoc(context.TODO(), constants.FunctionDBId, funcId)

		var f function.Function
		json.Unmarshal(bytes, &f)

		err = couchdb.DelDoc(context.TODO(), constants.FunctionDBId, f.ID, f.Reversion)

		if err != nil {
			ctx.JSON(http.StatusOK, gin.H{"error": err.Error()})
		}else {
			ctx.JSON(http.StatusOK, gin.H{"ok": true})
		}
	})

	return router
}