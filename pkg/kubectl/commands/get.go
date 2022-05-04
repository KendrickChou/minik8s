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
				resp := apiclient.Rest("", []byte{}, apiclient.OBJ_ALL_SERVICES, apiclient.OP_GET)
				fmt.Printf("%s\n", resp)
			case "replica":
				resp := apiclient.Rest("", []byte{}, apiclient.OBJ_ALL_REPLICAS, apiclient.OP_GET)
				fmt.Printf("%s\n", resp)
			case "all":
				resp := apiclient.Rest("", []byte{}, apiclient.OBJ_ALL_REPLICAS, apiclient.OP_GET)
				fmt.Printf("%s\n", resp)
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
	resp := apiclient.Rest("", []byte{}, apiclient.OBJ_ALL_PODS, apiclient.OP_GET)
	var kvs []GetPodResponse
	err := json.Unmarshal(resp, &kvs)
	if err != nil {
		fmt.Println("服务器返回信息无效: ", err)
		return
	}
	fmt.Println("=->Pods")
	fmt.Printf("%v\t\t\t\t%v\t\t\t%v\t\t\t%v\n", "Key", "Name", "Uid", "Status")
	for _, kv := range kvs {
		fmt.Printf("%v\t\t%v\t\t%v\t\t%v\n", kv.Key, kv.Pod.Name, kv.Pod.UID, kv.Pod.Status.Phase)
	}
}
