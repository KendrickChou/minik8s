package constants

import v1 "minik8s.com/minik8s/pkg/api/v1"

const EOF int = 26

const (
	RegistNodeRequest string = "/node"
	WatchNodeRequest  string = "/watch/node/"
	WatchPodsRequest  string = "/watch/pods/"
	HeartBeatRequest  string = "/heartbeat/"
)

const (
	HeartBeatInterval uint64 = 1  //second
	MaxErrorHeartBeat int    = 10 // if successively failed over 10 times, close node
)

const (
	NetworkIDPrefix          string = "container:"
	InitialPauseContainerKey string = "pause"
	NetworkBridgeName        string = "kubenet"
)

var InitialPauseContainer v1.Container = v1.Container{
	Name:            "pause",
	Image:           "kubernetes/pause:latest",
	ImagePullPolicy: "IfNotPresent",
}
