package main

import (
	"encoding/json"
	"fmt"
	"minik8s.com/minik8s/pkg/aqualake/apis/podserver"
	"reflect"
)

func main() {
	str := `{"Ret":true,"Err":""}`
	buf := []byte(str)
	var v podserver.TriggerResp
	err := json.Unmarshal(buf, &v)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(reflect.TypeOf(v.Ret))
}
