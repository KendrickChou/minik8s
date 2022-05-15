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
	AC_WatchNodes_Path    = "/watch/nodes"
	AC_WatchReplicas_Path = "/watch/replicas"
	AC_WatchEndpoints_Path = "/watch/endpoints"
	AC_WatchService_Path  = "/watch/service"
	AC_WatchPod_Path      = "/watch/pod"
	AC_WatchReplica_Path  = "/watch/replica"
	AC_WatchEndpoint_Path  = "/watch/endpoint"

	AC_RestServices_Path = "/services"
	AC_RestEndpoints_Path = "/endpoints"
	AC_RestPods_Path     = "/pods"
	AC_RestReplicas_Path = "/replicas"
	AC_RestNodes_Path    = "/nodes"
	AC_RestService_Path  = "/service"
	AC_RestPod_Path      = "/pod"
	AC_RestReplica_Path  = "/replica"
	AC_RestNode_Path     = "/node"
	AC_RestEndpoint_Path  = "/endpoint"

	AC_Root_Path = "/"
)
