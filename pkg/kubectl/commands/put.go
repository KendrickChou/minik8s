package commands

import (
	"encoding/json"
	"fmt"
	"io"
	"os"

	"github.com/spf13/cobra"
	"minik8s.com/minik8s/pkg/apiclient"
)

var putCmd = &cobra.Command{
	Use:   "put",
	Short: `更新资源`,
	Long:  `用于更新指定资源`,
	Run: func(cmd *cobra.Command, args []string) {

		id, err := cmd.Flags().GetString("id")
		if err != nil {
			fmt.Println("getString err: ", err)
			return
		}
		kind, err := cmd.Flags().GetString("kind")
		if err != nil {
			fmt.Println("getString err: ", err)
			return
		}
		filePath, err := cmd.Flags().GetString("file")
		if err != nil {
			fmt.Println("getString err: ", err)
			return
		}
		fmt.Println("正在打开配置文件: ", filePath)
		file, err := os.OpenFile(filePath, os.O_RDONLY, 0666)
		if err != nil {
			fmt.Println("文件打开失败：", err)
		}
		defer file.Close()

		buf, err := io.ReadAll(file)
		if err != nil {
			return
		}

		fmt.Println("正在更新对象: ", id)
		var resp []byte
		switch kind {
		case "pod":
			resp = apiclient.Rest(id, string(buf), apiclient.OBJ_POD, apiclient.OP_PUT)
			fmt.Printf("%s\n", resp)
		case "node":
			resp = apiclient.Rest(id, string(buf), apiclient.OBJ_NODE, apiclient.OP_PUT)
			fmt.Printf("%s\n", resp)
		case "service":
			resp = apiclient.Rest(id, string(buf), apiclient.OBJ_SERVICE, apiclient.OP_PUT)
			fmt.Printf("%s\n", resp)
		case "replica":
			resp = apiclient.Rest(id, string(buf), apiclient.OBJ_REPLICAS, apiclient.OP_PUT)
			fmt.Printf("%s\n", resp)
		case "dns":
			resp = apiclient.Rest(id, string(buf), apiclient.OBJ_DNS, apiclient.OP_PUT)
			fmt.Printf("%s\n", resp)
		case "hpa":
			resp = apiclient.Rest(id, string(buf), apiclient.OBJ_HPA, apiclient.OP_PUT)
			fmt.Printf("%s\n", resp)
		case "endpoint":
			resp = apiclient.Rest(id, string(buf), apiclient.OBJ_ENDPOINT, apiclient.OP_PUT)
			fmt.Printf("%s\n", resp)
		case "gpu":
			resp = apiclient.Rest(id, string(buf), apiclient.OBJ_GPU, apiclient.OP_PUT)
			fmt.Printf("%s\n", resp)
		case "function":
			resp = apiclient.Rest(id, string(buf), apiclient.OBJ_FUNCTION, apiclient.OP_PUT)
			fmt.Println("服务器返回: ", resp)
			return
		case "AC":
			resp = apiclient.Rest(id, string(buf), apiclient.OBJ_ACTCHAIN, apiclient.OP_PUT)
			fmt.Println("服务器返回: ", resp)
			return
		default:
			fmt.Println("找不到指定的对象！")
			return
		}

		var stat StatusResponse
		err = json.Unmarshal(resp, &stat)
		if err != nil {
			fmt.Println("服务器返回信息无效: ", err)
		} else if stat.Status != "OK" {
			fmt.Println("更新对象失败：", stat.Error)
		} else {
			fmt.Println("成功更新对象，id：", id)
		}
	},
}

func init() {
	putCmd.Flags().StringP("file", "f", "nginx_pod.json", "指定json配置文件")
	putCmd.Flags().StringP("id", "i", "X", "指定对象id")
	putCmd.Flags().StringP("kind", "k", "X", "指定对象id")

	rootCmd.AddCommand(putCmd)
}
