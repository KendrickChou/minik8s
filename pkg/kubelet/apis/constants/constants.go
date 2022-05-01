package constants

import v1 "minik8s.com/minik8s/pkg/api/v1"

const EOF int = 26

const (
	ApiServerAddress string = "addr"
)

const (
	RegistNodeRequest string = ApiServerAddress + "/node"
	WatchPodsRequest  string = ApiServerAddress + "/watch/pods/"
	HeartBeatRequest  string = ApiServerAddress + "/heartbeat/"
)

const (
	HeartBeatInterval uint64 = 1  //second
	MaxErrorHeartBeat int    = 10 // if successively failed over 10 times, close node
)

const (
	NetworkIDPrefix          string = "container:"
	InitialPauseContainerKey string = "pause"
)

var InitialPauseContainer v1.Container = v1.Container{
	Name:            "pause",
	Image:           "kubernetes/pause:latest",
	ImagePullPolicy: "IfNotPresent",
}
