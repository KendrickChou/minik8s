package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"minik8s.com/minik8s/pkg/aqualake/apis/actionchain"
)

func main() {
	funcId := "simple-func"
	actionId := "simple-actionchain"
	actionChain := actionchain.ActionChain{
		StartAt: "Begin",
		Chain: map[string]actionchain.Action{
			"Begin": {
				Type:     actionchain.ACT_TASK,
				Env:      "Python",
				Function: funcId,
				End:      true,
			},
		},
	}

	buf, err := json.Marshal(actionChain)

	if err != nil {
		fmt.Print(err.Error())
		return
	}

	fmt.Printf("%s\n", string(buf))

	req, _ := http.NewRequest("PUT", "http://127.0.0.1:8699/actionchain/"+actionId, bytes.NewReader(buf))

	resp, err := http.DefaultClient.Do(req)

	if err != nil {
		fmt.Print(err.Error())
		return
	}

	buf, _ = ioutil.ReadAll(resp.Body)
	fmt.Printf("%v\n", string(buf))

	resp, err = http.Get("http://127.0.0.1:8699/actionchain/" + actionId)
	if err != nil {
		fmt.Print(err.Error())
		return
	}

	buf, _ = ioutil.ReadAll(resp.Body)
	fmt.Printf("%v\n", string(buf))

	req, _ = http.NewRequest("DELETE", "http://127.0.0.1:8699/actionchain/"+actionId, nil)
	resp, err = http.DefaultClient.Do(req)

	if err != nil {
		fmt.Print(err.Error())
		return
	}

	buf, _ = ioutil.ReadAll(resp.Body)
	fmt.Printf("%v\n", string(buf))
}
