package constants

const EOF int = 26

const (
	ApiServerAddress string = "addr"
)

const (
	RegistNodeRequest string = ApiServerAddress + "/node"
	WatchPodsRequest string = ApiServerAddress + "/watch/pods/"
	HeartBeatRequest string = ApiServerAddress + "/heartbeat/"
)

const (
	HeartBeatInterval uint64 = 1 //second
	MaxErrorHeartBeat int = 10 // if successively failed over 10 times, close node
)