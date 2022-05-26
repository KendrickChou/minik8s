package podserver

type InstallFuncReq struct {
	Name string

	Url string `json:"Url"`
}

type InstallFuncResp struct {
	Ok bool

	Err string
}
