package commands

import (
	"bytes"
	"fmt"
	"github.com/spf13/cobra"
	"io"
	"minik8s.com/minik8s/config"
	"minik8s.com/minik8s/pkg/apiclient"
	"net/http"
	"strconv"
)

var delCmd = &cobra.Command{
	Use:   "del",
	Short: `销毁资源`,
	Long:  `用于销毁指定id的资源`,
	Run: func(cmd *cobra.Command, args []string) {

		id, err := cmd.Flags().GetString("id")
		if err != nil {
			fmt.Println("getString err: ", err)
			return
		}
		fmt.Println("正在删除对象: ", id)

		switch id[0] {
		case 'P':
			resp := apiclient.Rest(id, "", apiclient.OBJ_POD, apiclient.OP_DELETE)
			fmt.Printf("%s\n", resp)
		case 'N':
			resp := apiclient.Rest(id, "", apiclient.OBJ_NODE, apiclient.OP_DELETE)
			fmt.Printf("%s\n", resp)
		case 'S':
			resp := apiclient.Rest(id, "", apiclient.OBJ_SERVICE, apiclient.OP_DELETE)
			fmt.Printf("%s\n", resp)
		case 'R':
			resp := apiclient.Rest(id, "", apiclient.OBJ_REPLICAS, apiclient.OP_DELETE)
			fmt.Printf("%s\n", resp)
		case 'D':
			resp := apiclient.Rest(id, "", apiclient.OBJ_DNS, apiclient.OP_DELETE)
			fmt.Printf("%s\n", resp)
		case 'H':
			resp := apiclient.Rest(id, "", apiclient.OBJ_HPA, apiclient.OP_DELETE)
			fmt.Printf("%s\n", resp)
		case 'E':
			resp := apiclient.Rest(id, "", apiclient.OBJ_ENDPOINT, apiclient.OP_DELETE)
			fmt.Printf("%s\n", resp)
		case 'G':
			resp := apiclient.Rest(id, "", apiclient.OBJ_GPU, apiclient.OP_DELETE)
			fmt.Printf("%s\n", resp)
		case 'A':
			cli := http.Client{}
			url := config.AC_ServerAddr + ":" + strconv.Itoa(config.AC_ServerPort)
			req, _ := http.NewRequest(http.MethodDelete, url+"/", bytes.NewReader([]byte{}))
			resp, _ := cli.Do(req)
			buf, _ := io.ReadAll(resp.Body)
			fmt.Printf("%s\n", buf)
		default:
			fmt.Println("找不到指定的对象！")
		}
	},
}

func init() {
	delCmd.Flags().StringP("id", "i", "X", "指定对象id")

	rootCmd.AddCommand(delCmd)
}
