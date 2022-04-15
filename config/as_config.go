package config

//--------------- API SERVER CONFIG -------------------
const (
	AS_EtcdAddr = "127.0.0.1"
	AS_EtcdPort = 2379

	AS_HttpListenPort   = 8080
	AS_OP_PUT_String    = "PUT"
	AS_OP_GET_String    = "GET"
	AS_OP_DELETE_String = "DELETE"
	AS_OP_WATCH_String  = "POST"
	AS_OP_ERROR_String  = "ERROR"
)
