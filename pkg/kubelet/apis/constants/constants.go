package constants

import v1 "minik8s.com/minik8s/pkg/api/v1"

const EOF int = 26

func RegistNodeRequest() string {
	return "/node"
}

func WatchNodeRequest(nodeUID string) string {
	return "/watch/node/" + nodeUID
}

func WatchPodsRequest(podUID string) string {
	return "/watch/pods/" + podUID
}

func HeartBeatRequest(nodeUID string, counter string) string {
	return "/heartbeat/" + nodeUID + "/" + counter
}

func RefreshPodRequest(nodeUID string, podUID string) string {
	return "/node/" + nodeUID + "/podstatus/" + podUID
}

const (
	HeartBeatInterval        uint64 = 1  //second
	MaxErrorHeartBeat        int    = 10 // if successively failed over 10 times, close node
	RefreshPodStatusInterval uint64 = 10
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
