package config

//--------------- API SERVER CONFIG -------------------
const (
	AS_EtcdAddr = "127.0.0.1"
	AS_EtcdPort = 2379

	AS_GPU_HOMEPATH   = "/lustre/home/acct-stu/stu643/"
	AS_GPU_USERNAME   = "stu643"
	AS_GPU_LOGIN_ADDR = "login.hpc.sjtu.edu.cn"
	AS_GPU_DATA_ADDR  = "data.hpc.sjtu.edu.cn"
	AS_GPU_PASSWD     = "z8TVKO^n"

	AS_HttpListenPort   = 8080
	AS_OP_PUT_String    = "PUT"
	AS_OP_GET_String    = "GET"
	AS_OP_DELETE_String = "DELETE"
	AS_OP_POST_String   = "POST"
	AS_OP_ERROR_String  = "ERROR"
)

const (
	AC_ServerAddr = "http://10.119.11.209"
	AC_ServerPort = 8080

	AC_ServerlessAddr = "http://10.119.11.209"
	AC_ServerlessPort = 8699

	AC_WatchServices_Path  = "/watch/services"
	AC_WatchPods_Path      = "/watch/pods"
	AC_WatchNodes_Path     = "/watch/nodes"
	AC_WatchReplicas_Path  = "/watch/replicas"
	AC_WatchHPAs_Path      = "/watch/hpas"
	AC_WatchEndpoints_Path = "/watch/endpoints"
	AC_WatchDnss_Path      = "/watch/dnss"
	AC_WatchGpus_Path      = "/watch/gpus"
	AC_WatchService_Path   = "/watch/service"
	AC_WatchPod_Path       = "/watch/pod"
	AC_WatchReplica_Path   = "/watch/replica"
	AC_WatchHPA_Path       = "/watch/hpa"
	AC_WatchEndpoint_Path  = "/watch/endpoint"

	AC_RestServices_Path  = "/services"
	AC_RestEndpoints_Path = "/endpoints"
	AC_RestDnss_Path      = "/dnss"
	AC_RestGpus_Path      = "/gpus"
	AC_RestPods_Path      = "/pods"
	AC_RestReplicas_Path  = "/replicas"
	AC_RestHPAs_Path      = "/hpas"
	AC_RestNodes_Path     = "/nodes"
	AC_RestService_Path   = "/service"
	AC_RestPod_Path       = "/pod"
	AC_RestReplica_Path   = "/replica"
	AC_RestHPA_Path       = "/hpa"
	AC_RestNode_Path      = "/node"
	AC_RestEndpoint_Path  = "/endpoint"
	AC_RestDns_Path       = "/dns"
	AC_RestGpu_Path       = "/gpu"
	AC_RestFunction_Path  = "/function"
	AC_RestActChain_Path  = "/actionchain"
	AC_RestTrigger_Path  = "/trigger"

	AC_Root_Path = "/"
)
