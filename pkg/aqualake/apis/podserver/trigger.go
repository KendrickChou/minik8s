package podserver

type TriggerReq struct {
	Args interface{} `json:"Args"`
}

type TriggerResp struct {
	//Ret json.RawMessage `json:"Ret"`
	Ret interface{} `json:"Ret"`

	Err string `json:"Err"`
}
