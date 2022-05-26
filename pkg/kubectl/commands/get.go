package commands

import (
	"encoding/json"
	"fmt"
	"github.com/spf13/cobra"
	"minik8s.com/minik8s/pkg/apiclient"
)

var getCmd = &cobra.Command{
	Use:   "get",
	Short: `查询资源`,
	Long:  `用于查询资源`,
	Run: func(cmd *cobra.Command, args []string) {

		kind, err := cmd.Flags().GetString("kind")
		if err != nil {
			fmt.Println("getString err: ", err)
			return
		}
		id, err := cmd.Flags().GetString("id")
		if err != nil {
			fmt.Println("getString err: ", err)
			return
		}

		if id == "X" {
			fmt.Println("正在查询指定资源类型的所有对象: ", kind)
			switch kind {
			case "pod":
				getPods()
			case "service":
				getServices()
			case "replica":
				getReplicaSet()
			case "node":
				getNodes()
			case "endpoint":
				getEndpoints()
			case "all":
				getPods()
				getNodes()
				getServices()
				getEndpoints()
				getReplicaSet()
			default:
				fmt.Println("未知的对象类型！")
			}
		}

	},
}

func init() {
	getCmd.Flags().StringP("kind", "k", "all", "指定访问的对象类型")
	getCmd.Flags().StringP("id", "i", "X", "指定访问的对象id")

	rootCmd.AddCommand(getCmd)
}

func getPods() {
	resp := apiclient.Rest("", "", apiclient.OBJ_ALL_PODS, apiclient.OP_GET)
	var kvs []GetPodResponse
	err := json.Unmarshal(resp, &kvs)
	if err != nil {
		fmt.Println("服务器返回信息无效: ", err)
		return
	}
	fmt.Printf("=->%v Pods\n", len(kvs))
	fmt.Printf("%v\t\t\t\t%v\t\t\t%v\t\t\t%v\n", "Key", "Name", "Uid", "OwnerReferences")
	for _, kv := range kvs {
		fmt.Printf("%v\t\t%v\t\t%v\t\t", kv.Key, kv.Pod.Name, kv.Pod.UID)
		for _, owner := range kv.Pod.OwnerReferences {
			fmt.Printf("%v: %v, ", owner.Kind, owner.UID)
		}
		fmt.Printf("\n")
	}
	fmt.Printf("\n")
}

func getNodes() {
	resp := apiclient.Rest("", "", apiclient.OBJ_ALL_NODES, apiclient.OP_GET)
	var kvs []GetNodeResponse
	err := json.Unmarshal(resp, &kvs)
	if err != nil {
		fmt.Println("服务器返回信息无效: ", err)
		return
	}
	fmt.Printf("=->%v Nodes\n", len(kvs))
	fmt.Printf("%v\t\t\t\t%v\t\t\t%v\t\t\t%v\n", "Key", "Name", "Uid", "IP")
	for _, kv := range kvs {
		fmt.Printf("%v\t\t%v\t\t%v\t\t%v\n", kv.Key, kv.Node.Name, kv.Node.UID, kv.Node.Spec.IP)
	}
	fmt.Printf("\n")
}

func getServices() {
	resp := apiclient.Rest("", "", apiclient.OBJ_ALL_SERVICES, apiclient.OP_GET)
	var kvs []GetServiceResponse
	err := json.Unmarshal(resp, &kvs)
	if err != nil {
		fmt.Println("服务器返回信息无效: ", err)
		return
	}
	fmt.Printf("=->%v Services\n", len(kvs))
	fmt.Printf("%v\t\t\t\t%v\t\t\t%v\t\t\t%v\n", "Key", "Name", "Uid", "Cluster IP")
	for _, kv := range kvs {
		fmt.Printf("%v\t\t%v\t\t%v\t\t%v\n", kv.Key, kv.Service.Name, kv.Service.UID, kv.Service.Spec.ClusterIP)
	}
	fmt.Printf("\n")
}

func getEndpoints() {
	resp := apiclient.Rest("", "", apiclient.OBJ_ALL_ENDPOINTS, apiclient.OP_GET)
	var kvs []GetEndpointResponse
	err := json.Unmarshal(resp, &kvs)
	if err != nil {
		fmt.Println("服务器返回信息无效: ", err)
		return
	}
	fmt.Printf("=->%v Endpoints\n", len(kvs))
	fmt.Printf("%v\t\t\t\t\t%v\t\t\t%v\t\t\t%v\n", "Key", "Name", "Uid", "")
	for _, kv := range kvs {
		fmt.Printf("%v\t\t%v\t\t%v\t\t%v\n", kv.Key, kv.Endpoint.Name, kv.Endpoint.UID)
	}
	fmt.Printf("\n")
}

func getReplicaSet() {
	resp := apiclient.Rest("", "", apiclient.OBJ_ALL_REPLICAS, apiclient.OP_GET)
	var kvs []GetReplicaResponse
	err := json.Unmarshal(resp, &kvs)
	if err != nil {
		fmt.Println("服务器返回信息无效: ", err)
		return
	}
	fmt.Printf("=->%v ReplicaSets\n", len(kvs))
	fmt.Printf("%v\t\t\t\t\t%v\t\t\t%v\t\t\t%v\n", "Key", "Name", "Uid", "")
	for _, kv := range kvs {
		fmt.Printf("%v\t\t%v\t\t%v\t\t%v\n", kv.Key, kv.ReplicaSet.Name, kv.ReplicaSet.UID, "")
	}
	fmt.Printf("\n")
}
