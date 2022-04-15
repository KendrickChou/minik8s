package apiserver

type ApiServer struct {
}

func CreateNewApiServer() *ApiServer {
	var as = new(ApiServer)
	return as
}

func Init() {
	initEtcd()
	runHttpServer()
}

//------------------------ API SERVER TEST -----------------------------

//
//func TestEtcd() {
//	initEtcd()
//	defer closeEtcd()
//	wch := etcdWatch("hello")
//	go handleWatchResult(wch)
//	etcdPut("hello", "world")
//	etcdPut("hello1", "world1")
//	etcdPut("hello2", "world2")
//	etcdPut("hello3", "world3")
//	etcdGetPrefix("hello")
//}
//
//func handleWatchResult(wch chan KV) {
//	for kv := range wch {
//		fmt.Printf("user watch key: %v, val: %v\n", kv.key, kv.value)
//	}
//}
//
//func TestHttp() {
//	runHttpServer()
//}
