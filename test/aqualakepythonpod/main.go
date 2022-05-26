package main

import (
	"fmt"

	pythonenv "minik8s.com/minik8s/pkg/aqualake/podenv/python"
)

func main(){
	pythonenv.FunctionName = "hello_word"

	res := pythonenv.CallPythonFunction("Kendrick")

	fmt.Print(res)
}