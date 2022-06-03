package controller

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"net/http"

	"minik8s.com/minik8s/pkg/aqualake/invoker"

	"github.com/gin-gonic/gin"
	"minik8s.com/minik8s/pkg/aqualake/apis/actionchain"
	"minik8s.com/minik8s/pkg/aqualake/apis/constants"
	"minik8s.com/minik8s/pkg/aqualake/apis/couchobject"
	"minik8s.com/minik8s/pkg/aqualake/couchdb"
)

// need to do some init in a script(maybe a set env script)
var ivk *invoker.Invoker

func SetUpRouter() *gin.Engine {
	ivk = invoker.NewInvoker()

	router := gin.Default()

	// Function Related
	router.GET("/function/:id", func(ctx *gin.Context) {
		funcId := ctx.Params.ByName("id")
		file, err := couchdb.GetFile(context.TODO(), constants.FunctionDBId, funcId, funcId)

		if err != nil {
			ctx.JSON(http.StatusOK, gin.H{"error": err.Error()})
		} else {
			ctx.JSON(http.StatusOK, gin.H{"function": file})
		}
	})

	router.GET("/functions", func(ctx *gin.Context) {
		docs, err := couchdb.GetAllDoc(context.TODO(), constants.FunctionDBId)

		if err != nil {
			ctx.JSON(http.StatusOK, gin.H{"error": err.Error()})
		} else {
			var body map[string]interface{}
			var res []string
			json.Unmarshal(docs, &body)

			if _, ok := body["total_rows"]; ok {
				rows := body["rows"].([]interface{})

				for _, function := range rows {
					if value, ok := function.(map[string]interface{})["id"]; ok {
						res = append(res, value.(string))
					}
				}
			}
			ctx.JSON(http.StatusOK, gin.H{"functions": res})
		}
	})

	router.POST("/function/:id", func(ctx *gin.Context) {
		funcId := ctx.Params.ByName("id")
		rev, err := couchdb.PutDoc(context.TODO(), constants.FunctionDBId, funcId, []byte("{}"))

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
		if err != nil {
			ctx.JSON(500, gin.H{"error": err.Error()})
			return
		}

		var f couchobject.Function
		json.Unmarshal(bytes, &f)

		err = couchdb.DelDoc(context.TODO(), constants.FunctionDBId, f.ID, f.Reversion)

		if err != nil {
			ctx.JSON(http.StatusOK, gin.H{"error": err.Error()})
		} else {
			ctx.JSON(http.StatusOK, gin.H{"ok": true})
		}
	})

	router.PUT("/function/:id", func(ctx *gin.Context) {
		funcId := ctx.Params.ByName("id")
		bytes, err := couchdb.GetDoc(context.TODO(), constants.FunctionDBId, funcId)
		if err != nil {
			ctx.JSON(500, gin.H{"error": err.Error()})
			return
		}

		var f couchobject.Function
		json.Unmarshal(bytes, &f)

		buf, err := ioutil.ReadAll(ctx.Request.Body)

		if err != nil {
			ctx.JSON(http.StatusOK, gin.H{"error": err.Error()})
			return
		}

		err = couchdb.PutFile(context.TODO(), constants.FunctionDBId, funcId, funcId, f.Reversion, string(buf))

		if err != nil {
			ctx.JSON(http.StatusOK, gin.H{"error": err.Error()})
		} else {
			ctx.JSON(http.StatusOK, gin.H{"ok": true})
		}
	})

	// Action Related
	router.GET("/actionchain/:id", func(ctx *gin.Context) {
		acId := ctx.Params.ByName("id")
		bytes, err := couchdb.GetDoc(context.TODO(), constants.ActionDBId, acId)

		if err != nil {
			ctx.JSON(http.StatusOK, gin.H{"error": err.Error()})
			return
		}

		var ac couchobject.ActionChain
		err = json.Unmarshal(bytes, &ac)

		if err != nil {
			ctx.JSON(http.StatusOK, gin.H{"error": err.Error()})
		} else {
			ctx.JSON(http.StatusOK, ac.ActionChain)
		}

	})

	router.GET("/actionchains", func(ctx *gin.Context) {
		docs, err := couchdb.GetAllDoc(context.TODO(), constants.ActionDBId)

		if err != nil {
			ctx.JSON(http.StatusOK, gin.H{"error": err.Error()})
		} else {
			var body map[string]interface{}
			res := make(map[string]actionchain.ActionChain)
			json.Unmarshal(docs, &body)

			if _, ok := body["total_rows"]; ok {
				rows := body["rows"].([]interface{})

				for _, row := range rows {
					if id, ok := row.(map[string]interface{})["id"]; ok {
						var actionChain actionchain.ActionChain
						buf, _ := json.Marshal(row.(map[string]interface{})["value"].(map[string]interface {}))
						err = json.Unmarshal(buf, &actionChain)
						if err != nil {
							ctx.JSON(http.StatusOK, gin.H{"error": err.Error()})
							return
						}

						res[id.(string)] = actionChain
					}
				}
			}
			ctx.JSON(http.StatusOK, gin.H{"actionchains": res})
		}
	})

	router.POST("/actionchain/:id", func(ctx *gin.Context) {
		acId := ctx.Params.ByName("id")

		buf, err := ioutil.ReadAll(ctx.Request.Body)

		// without check, maybe I should add some error check

		if err != nil {
			ctx.JSON(http.StatusOK, gin.H{"error": err.Error()})
			return
		}

		_, err = couchdb.PutDoc(context.TODO(), constants.ActionDBId, acId, buf)

		if err != nil {
			ctx.JSON(http.StatusOK, gin.H{"error": err.Error()})
		} else {
			ctx.JSON(http.StatusOK, gin.H{"ok": true})
		}
	})

	router.DELETE("/actionchain/:id", func(ctx *gin.Context) {
		acId := ctx.Params.ByName("id")
		bytes, err := couchdb.GetDoc(context.TODO(), constants.ActionDBId, acId)

		if err != nil {
			ctx.JSON(http.StatusOK, gin.H{"error": err.Error()})
			return
		}

		var ac couchobject.ActionChain
		json.Unmarshal(bytes, &ac)

		err = couchdb.DelDoc(context.TODO(), constants.ActionDBId, ac.ID, ac.Reversion)
		if err != nil {
			ctx.JSON(http.StatusOK, gin.H{"error": err.Error()})
		} else {
			ctx.JSON(http.StatusOK, gin.H{"ok": true})
		}
	})

	router.PUT("/actionchain/:id", func(ctx *gin.Context) {
		acId := ctx.Params.ByName("id")
		bytes, err := couchdb.GetDoc(context.TODO(), constants.ActionDBId, acId)

		if err != nil {
			ctx.JSON(http.StatusOK, gin.H{"error": err.Error()})
			return
		}

		var ac couchobject.ActionChain
		json.Unmarshal(bytes, &ac)

		err = couchdb.DelDoc(context.TODO(), constants.ActionDBId, ac.ID, ac.Reversion)

		if err != nil {
			ctx.JSON(http.StatusOK, gin.H{"error": err.Error()})
		} else {
			ctx.JSON(http.StatusOK, gin.H{"ok": true})
		}

		buf, err := ioutil.ReadAll(ctx.Request.Body)

		// without check, maybe I should add some error check

		if err != nil {
			ctx.JSON(http.StatusOK, gin.H{"error": err.Error()})
			return
		}

		_, err = couchdb.PutDoc(context.TODO(), constants.ActionDBId, acId, buf)

		if err != nil {
			ctx.JSON(http.StatusOK, gin.H{"error": err.Error()})
		} else {
			ctx.JSON(http.StatusOK, gin.H{"ok": true})
		}
	})

	router.GET("/trigger/:id", func(ctx *gin.Context) {
		buf, _ := ioutil.ReadAll(ctx.Request.Body)
		acId := ctx.Params.ByName("id")
		bytes, err := couchdb.GetDoc(context.TODO(), constants.ActionDBId, acId)

		if err != nil {
			ctx.JSON(http.StatusOK, gin.H{"error": err.Error()})
			return
		}

		var ac couchobject.ActionChain
		err = json.Unmarshal(bytes, &ac)
		if err != nil {
			ctx.JSON(500, gin.H{"error": err.Error()})
			return
		}
		var arg interface{}
		err = json.Unmarshal(buf, &arg)
		if err != nil {
			ctx.JSON(500, gin.H{"error": err.Error()})
			return
		} else {
			ret, err := ivk.InvokeActionChain(ac.ActionChain, arg)
			if err != nil {
				ctx.JSON(500, gin.H{"error": err.Error()})
			} else {
				ctx.JSON(http.StatusOK, gin.H{"staus": "OK", "result": ret})
			}
		}
	})

	return router
}
