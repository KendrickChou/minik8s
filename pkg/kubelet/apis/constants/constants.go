package constants

import v1 "minik8s.com/minik8s/pkg/api/v1"

const EOF int = 26

func RegistNodeRequest() string {
	return "/node"
}

func WatchNodeRequest(nodeUID string) string {
	return "/watch/innode/" + nodeUID
}

func WatchPodsRequest(nodeUID string) string {
	return "/watch/innode/" + nodeUID + "/pods"
}

func HeartBeatRequest(nodeUID string, counter string) string {
	return "/heartbeat/" + nodeUID + "/" + counter
}

func RefreshPodStatusRequest(nodeUID string, podUID string) string {
	return "/innode/" + nodeUID + "/podstatus/" + podUID
}

func RefreshPodRequest(nodeUID string, podUID string) string {
	return "/innode/" + nodeUID + "/pod/" + podUID
}

func WatchEndpointsRequest() string {
	return "/watch/endpoints"
}

func GetAllEndpointsRequest() string {
	return "/endpoints"
}

func GetAllPodsRequest(nodeID string) string {
	return "/innode/" + nodeID + "/pods"
}

const (
	HeartBeatInterval        uint64 = 1  //second
	MaxErrorHeartBeat        int    = 10 // if successively failed over 10 times, close node
	RefreshPodStatusInterval uint64 = 10
	ReconnectInterval        uint64 = 20
)

const (
	CacheFilePath string = "./cache"
	NodeCacheID   string = "node"

	NetworkIDPrefix          string = "container:"
	InitialPauseContainerKey string = "pause"
	WeaveNetworkName         string = "weave"
	NetworkBridgeName        string = "kubenet"

	NATTableName string = "nat"

	ServiceChainName   string = "K8S-SERVICE"
	ServiceChainPrefix string = "K8S-SVC-"
	SepChainPrefix     string = "K8S-SEP-"

	DNS       string = "172.17.0.1"
	DNSSearch string = "weave.local.root"
)

var InitialPauseContainer v1.Container = v1.Container{
	Name:            "pause",
	Image:           "kubernetes/pause:latest",
	ImagePullPolicy: "IfNotPresent",
	DNS:             DNS,
	DNSSearch:       DNSSearch,
}

var NodeUID string = ""
