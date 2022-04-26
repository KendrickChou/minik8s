package config

//--------------- API SERVER CONFIG -------------------
const (
	AS_EtcdAddr = "127.0.0.1"
	AS_EtcdPort = 2379

	AS_HttpListenPort   = 8080
	AS_OP_PUT_String    = "PUT"
	AS_OP_GET_String    = "GET"
	AS_OP_DELETE_String = "DELETE"
	AS_OP_POST_String   = "POST"
	AS_OP_ERROR_String  = "ERROR"
)

const (
	AC_ServerAddr = "http://127.0.0.1"
	AC_ServerPort = 8080

	AC_WatchServices_Path = "/watch/services"
	AC_WatchPods_Path     = "/watch/pods"
	AC_WatchReplicas_Path = "/watch/replicas"
	AC_WatchService_Path  = "/watch/service"
	AC_WatchPod_Path      = "/watch/pod"
	AC_WatchReplica_Path  = "/watch/replica"

	AC_RestServices_Path = "/minik8s/services"
	AC_RestPods_Path     = "/minik8s/pods"
	AC_RestReplicas_Path = "/minik8s/replicas"
	AC_RestService_Path  = "/minik8s/service"
	AC_RestPod_Path      = "/minik8s/pod"
	AC_RestReplica_Path  = "/minik8s/replica"
)
