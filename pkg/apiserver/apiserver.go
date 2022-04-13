package apiserver

type ApiServer struct {
}

func CreateNewApiServer() *ApiServer {
	var as = new(ApiServer)
	return as
}

func TestEtcd() {
	initEtcd()
	etcdPut("hello", "world")
	etcdGet("hello")
}
