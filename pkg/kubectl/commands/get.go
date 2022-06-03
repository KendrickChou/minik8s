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
			case "dns":
				getDNSs()
			case "hpa":
				getHPAs()
			case "function":
				getFuntions()
			case "AC":
				getACs()
			case "all":
				getPods()
				getNodes()
				getServices()
				getDNSs()
				getEndpoints()
				getReplicaSet()
				getHPAs()
			default:
				fmt.Println("未知的对象类型！")
			}
		} else {
			fmt.Println("正在查询指定对象: ", kind)
			switch id[0] {
			case 'P':
				resp := apiclient.Rest(id, "", apiclient.OBJ_POD, apiclient.OP_GET)
				fmt.Printf("%s\n", resp)
			case 'S':
				resp := apiclient.Rest(id, "", apiclient.OBJ_SERVICE, apiclient.OP_GET)
				fmt.Printf("%s\n", resp)
			case 'N':
				resp := apiclient.Rest(id, "", apiclient.OBJ_NODE, apiclient.OP_GET)
				fmt.Printf("%s\n", resp)
			case 'R':
				resp := apiclient.Rest(id, "", apiclient.OBJ_REPLICAS, apiclient.OP_GET)
				fmt.Printf("%s\n", resp)
			case 'D':
				resp := apiclient.Rest(id, "", apiclient.OBJ_DNS, apiclient.OP_GET)
				fmt.Printf("%s\n", resp)
			case 'H':
				resp := apiclient.Rest(id, "", apiclient.OBJ_HPA, apiclient.OP_GET)
				fmt.Printf("%s\n", resp)
			case 'E':
				resp := apiclient.Rest(id, "", apiclient.OBJ_ENDPOINT, apiclient.OP_GET)
				fmt.Printf("%s\n", resp)
			case 'G':
				resp := apiclient.Rest(id, "", apiclient.OBJ_GPU, apiclient.OP_GET)
				fmt.Printf("%s\n", resp)
			default:
				fmt.Println("找不到指定的对象！")
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
	fmt.Printf("\n============\n")
	fmt.Printf("=->%v Pods<-=", len(kvs))
	fmt.Printf("\n============\n")
	fmt.Printf("%v\t\t\t\t%v\t\t\t%v\t\t\t%v\t\t\t%v\t%v\t\t%v\n", "Key", "Name", "Uid", "Node", "PodStatus", "podIP", "OwnerReferences")
	for _, kv := range kvs {
		fmt.Printf("%v\t\t%v\t%v\t\t%v\t\t%v\t\t%v\t\t", kv.Key, kv.Pod.Name, kv.Pod.UID, kv.Pod.Spec.NodeName, kv.Pod.Status.Phase, kv.Pod.Status.PodIP)
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
	fmt.Printf("\n=============\n")
	fmt.Printf("=->%v Nodes<-=", len(kvs))
	fmt.Printf("\n=============\n")
	fmt.Printf("%v\t\t\t\t%v\t\t\t%v\t\t\t%v\n", "Key", "Name", "Uid", "Status")
	for _, kv := range kvs {
		fmt.Printf("%v\t\t%v\t\t%v\t\t%v\n", kv.Key, kv.Node.Name, kv.Node.UID, kv.Node.Status.Phase)
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
	fmt.Printf("\n================\n")
	fmt.Printf("=->%v Services<-=", len(kvs))
	fmt.Printf("\n================\n")
	fmt.Printf("%v\t\t\t\t%v\t\t\t%v\t\t\t%v\n", "Key", "Name", "Uid", "Cluster IP")
	for _, kv := range kvs {
		fmt.Printf("%v\t\t%v\t\t%v\t\t%v\n", kv.Key, kv.Service.Name, kv.Service.UID, kv.Service.Spec.ClusterIP)
	}
	fmt.Printf("\n")
}
func getDNSs() {
	resp := apiclient.Rest("", "", apiclient.OBJ_ALL_DNSS, apiclient.OP_GET)
	var kvs []GetDNSResponse
	err := json.Unmarshal(resp, &kvs)
	if err != nil {
		fmt.Println("服务器返回信息无效: ", err)
		return
	}
	fmt.Printf("\n===================\n")
	fmt.Printf("=->%v DNS Configs<-=", len(kvs))
	fmt.Printf("\n===================\n")
	fmt.Printf("%v\t\t\t\t%v\t\t\t%v\t\t\t%v\n", "Key", "Name", "Uid", "Paths")
	for _, kv := range kvs {
		fmt.Printf("%v\t\t%v\t\t%v\t\t%v\n", kv.Key, kv.DNS.Name, kv.DNS.UID, kv.DNS.Paths)
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
	fmt.Printf("\n=================\n")
	fmt.Printf("=->%v Endpoints<-=", len(kvs))
	fmt.Printf("\n=================\n")
	fmt.Printf("%v\t\t\t\t\t%v\t\t\t%v\t\t\t%v\n", "Key", "Name", "Uid", "OwnerReferences")
	for _, kv := range kvs {
		fmt.Printf("%v\t\t%v\t\t%v\t\t", kv.Key, kv.Endpoint.Name, kv.Endpoint.UID)
		for _, owner := range kv.Endpoint.OwnerReferences {
			fmt.Printf("%v: %v, ", owner.Kind, owner.UID)
		}
		fmt.Printf("\n")

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
	fmt.Printf("\n===================\n")
	fmt.Printf("=->%v ReplicaSets<-=", len(kvs))
	fmt.Printf("\n===================\n")
	fmt.Printf("%v\t\t\t\t\t%v\t\t\t%v\t\t\t%v\n", "Key", "Name", "Uid", "MatchedApp")
	for _, kv := range kvs {
		fmt.Printf("%v\t\t%v\t\t%v\t\t%v\n", kv.Key, kv.ReplicaSet.Name, kv.ReplicaSet.UID, kv.ReplicaSet.Spec.Selector.MatchLabels["app"])
	}
	fmt.Printf("\n")
}

func getHPAs() {
	resp := apiclient.Rest("", "", apiclient.OBJ_ALL_HPAS, apiclient.OP_GET)
	var kvs []GetHpaResponse
	err := json.Unmarshal(resp, &kvs)
	if err != nil {
		fmt.Println("服务器返回信息无效: ", err)
		return
	}
	fmt.Printf("\n============\n")
	fmt.Printf("=->%v HPAs<-=", len(kvs))
	fmt.Printf("\n============\n")
	fmt.Printf("%v\t\t\t\t\t%v\t\t\t%v\t\t\t%v\n", "Key", "Name", "Uid", "TargetReplicaSet")
	for _, kv := range kvs {
		fmt.Printf("%v\t\t%v\t\t%v\t\t%v\n", kv.Key, kv.HPA.Name, kv.HPA.UID, kv.HPA.Spec.ScaleTargetRef.Name)
	}
	fmt.Printf("\n")
}

func getFuntions() {
	resp := apiclient.Rest("", "", apiclient.OBJ_ALL_FUNCTIONS, apiclient.OP_GET)
	var kvs GetFunctionResponse
	err := json.Unmarshal(resp, &kvs)
	if err != nil {
		fmt.Println("服务器返回信息无效: ", err)
		return
	}
	fmt.Printf("\n=================\n")
	fmt.Printf("=->%v Funcitons<-=", len(kvs.Funcitons))
	fmt.Printf("\n=================\n")
	fmt.Printf("%v\n", "Name")
	for _, kv := range kvs.Funcitons {
		fmt.Printf("%v\n", kv)
	}
	fmt.Printf("\n")
}

func getACs() {
	resp := apiclient.Rest("", "", apiclient.OBJ_ALL_ACTCHAINS, apiclient.OP_GET)
	var kvs GetFunctionResponse
	err := json.Unmarshal(resp, &kvs)
	if err != nil {
		fmt.Println("服务器返回信息无效: ", err)
		return
	}
	fmt.Printf("\n=================\n")
	fmt.Printf("=->%v Funcitons<-=", len(kvs.Funcitons))
	fmt.Printf("\n=================\n")
	fmt.Printf("%v\n", "Name")
	for _, kv := range kvs.Funcitons {
		fmt.Printf("%v\n", kv)
	}
	fmt.Printf("\n")
}
