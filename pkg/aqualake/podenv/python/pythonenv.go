package main

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"os/exec"
	"reflect"

	"github.com/gin-gonic/gin"
	"github.com/sbinet/go-python"
	"k8s.io/klog"
	"minik8s.com/minik8s/pkg/aqualake/apis/config"
	"minik8s.com/minik8s/pkg/aqualake/apis/podserver"
)

// const FunctionPath string = "/root/aqualake/function.py"
const FunctionPath string = "/root/functions"

var FunctionName string

func setUpRouter() *gin.Engine {
	router := gin.Default()

	router.POST("/installfunction", func(ctx *gin.Context) {
		buf, err := ioutil.ReadAll(ctx.Request.Body)

		var installFuncResp podserver.InstallFuncResp

		if err != nil {
			klog.Errorf("InstallFunction Body Err: %s", err.Error())
			installFuncResp.Ok = false
			installFuncResp.Err = err.Error()
			ctx.JSON(http.StatusOK, installFuncResp)
			return
		}

		var installFuncReq podserver.InstallFuncReq
		err = json.Unmarshal(buf, &installFuncReq)

		if err != nil {
			klog.Errorf("InstallFunction Marshal Body Err: %s", err.Error())
			installFuncResp.Ok = false
			installFuncResp.Err = err.Error()
			ctx.JSON(http.StatusOK, installFuncResp)
			return
		}

		cmd := exec.Command("wget", "-O", FunctionPath, installFuncReq.Url)

		res, err := cmd.CombinedOutput()

		if err != nil {
			klog.Errorf("Fetch Function Err: %s", err.Error())
			installFuncResp.Ok = false
			installFuncResp.Err = err.Error()
			ctx.JSON(http.StatusOK, installFuncResp)
			return
		}

		FunctionName = installFuncReq.Name

		klog.Infof("Fetch Function %s, res %s", installFuncReq.Name, res)
		installFuncResp.Ok = true
		installFuncResp.Err = ""
		ctx.JSON(http.StatusOK, installFuncResp)
	})

	router.POST("/trigger", func(ctx *gin.Context) {
		var triggerResp podserver.TriggerResp

		if FunctionName != "" {
			err := "Function is not installed"
			klog.Errorf(err)
			triggerResp.Err = err
			ctx.JSON(http.StatusOK, triggerResp)
			return
		}

		var triggerReq podserver.TriggerReq

		buf, err := ioutil.ReadAll(ctx.Request.Body)
		err = json.Unmarshal(buf, &triggerReq)

		if err != nil {
			klog.Errorf("InstallFunction Marshal Body Err: %s", err.Error())
			triggerResp.Err = err.Error()
			ctx.JSON(http.StatusOK, triggerResp)
			return
		}

		triggerResp.Ret = CallPythonFunction(triggerReq.Args)

		ctx.JSON(http.StatusOK, triggerResp)
	})

	return router
}

func main() {
	cmd := exec.Command("/bin/sh", "-c", "cat /etc/hosts | awk 'END{print $1}'")
	ip, err := cmd.CombinedOutput()

	if err != nil {
		klog.Fatal("Get Container IP error %s", err.Error())
		return
	}

	r := setUpRouter()

	klog.Infof("Server Run In %s:%s", string(ip), config.PodServePort)

	r.Run(string(ip) + ":" + config.PodServePort)
}

func CallPythonFunction(args interface{}) interface{} {
	module := ImportModule(FunctionPath, "function")
	function := module.GetAttrString(FunctionName)
	func_args := convertArgToPythonObj(args)

	klog.Info(function, func_args)

	res := function.Call(func_args, python.Py_None)

	return convertPythonObjToRet(res)
}

func convertArgToPythonObj(arg interface{}) *python.PyObject {
	value := reflect.ValueOf(arg)

	if value.IsNil() {
		return python.Py_None
	}

	switch reflect.TypeOf(arg).Kind() {
	case reflect.Slice, reflect.Array:
		args := []*python.PyObject{}
		for i := 0; i < value.Len(); i++ {
			args = append(args, convertArgToPythonObj(value.Index(i)))
		}
		return python.PyTuple_Pack(value.Len(), args...)
	case reflect.String:
		return python.PyString_FromString(value.String())
	case reflect.Int:
		return python.PyInt_FromLong(int(value.Int()))
	case reflect.Bool:
		b := 0
		if value.Bool() {
			b = 1
		}
		return python.PyBool_FromLong(b)
	default:
		klog.Infof("Unknown Type in convert arg to python obj")
		return nil
	}
}

func convertPythonObjToRet(pyObj *python.PyObject) interface{} {
	if python.Py_None == pyObj {
		return nil
	} else if python.PyInt_Check(pyObj) {
		return python.PyInt_AsLong(pyObj)
	} else if python.PyString_Check(pyObj) {
		return python.PyString_AS_STRING(pyObj)
	} else if python.PyTuple_Check(pyObj) {
		return pyObj.Repr().String()
	} else if python.PyBool_Check(pyObj) {
		return pyObj.IsTrue()
	}
	return nil
}

func ImportModule(dir, name string) *python.PyObject {
	sysModule := python.PyImport_ImportModule("sys")               // import sys
	path := sysModule.GetAttrString("path")                        // path = sys.path
	python.PyList_Insert(path, 0, python.PyString_FromString(dir)) // path.insert(0, dir)
	return python.PyImport_ImportModule(name)                      // return __import__(name)
}
