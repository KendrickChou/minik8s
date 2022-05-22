package commands

import (
	"encoding/json"
	"fmt"
	"github.com/spf13/cobra"
	"io"
	"minik8s.com/minik8s/pkg/apiclient"
	"os"
)

var addCmd = &cobra.Command{
	Use:   "add",
	Short: `创建资源`,
	Long:  `用于创建资源`,
	Run: func(cmd *cobra.Command, args []string) {
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

		kind, err := cmd.Flags().GetString("kind")

		var resp []byte
		switch kind {
		case "pod":
			resp = apiclient.Rest("", string(buf), apiclient.OBJ_POD, apiclient.OP_POST)
		case "service":
			resp = apiclient.Rest("", string(buf), apiclient.OBJ_SERVICE, apiclient.OP_POST)
		case "dns":
			resp = apiclient.Rest("", string(buf), apiclient.OBJ_DNS, apiclient.OP_POST)
		case "replica":
			resp = apiclient.Rest("", string(buf), apiclient.OBJ_REPLICAS, apiclient.OP_POST)
		}

		var stat StatusResponse
		err = json.Unmarshal(resp, &stat)
		if err != nil {
			fmt.Println("服务器返回信息无效: ", err)
		} else if stat.Status != "OK" {
			fmt.Println("创建对象失败：", stat.Error)
		} else {
			fmt.Println("成功创建对象，id：", stat.Id)
		}

	},
}

func init() {
	addCmd.Flags().StringP("file", "f", "default.json", "指定json配置文件")
	addCmd.Flags().StringP("kind", "k", "pod", "指定创建对象类型")

	rootCmd.AddCommand(addCmd)
}
