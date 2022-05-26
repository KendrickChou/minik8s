package podserver

type TriggerReq struct {
	Args interface{}
}

type TriggerResp struct {
	Ret interface{}

	Err string
}
