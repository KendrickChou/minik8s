package commands

import (
	"fmt"
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
		fmt.Println("正在更新对象: ", id)

		switch id[0] {
		case 'P':
			resp := apiclient.Rest(id, "", apiclient.OBJ_POD, apiclient.OP_PUT)
			fmt.Printf("%s\n", resp)
		case 'S':
			resp := apiclient.Rest(id, "", apiclient.OBJ_SERVICE, apiclient.OP_PUT)
			fmt.Printf("%s\n", resp)
		case 'R':
			resp := apiclient.Rest(id, "", apiclient.OBJ_REPLICAS, apiclient.OP_PUT)
			fmt.Printf("%s\n", resp)
		case 'D':
			resp := apiclient.Rest(id, "", apiclient.OBJ_DNS, apiclient.OP_PUT)
			fmt.Printf("%s\n", resp)
		case 'H':
			resp := apiclient.Rest(id, "", apiclient.OBJ_HPA, apiclient.OP_PUT)
			fmt.Printf("%s\n", resp)
		case 'E':
			resp := apiclient.Rest(id, "", apiclient.OBJ_ENDPOINT, apiclient.OP_PUT)
			fmt.Printf("%s\n", resp)
		case 'G':
			resp := apiclient.Rest(id, "", apiclient.OBJ_GPU, apiclient.OP_PUT)
			fmt.Printf("%s\n", resp)
		default:
			fmt.Println("找不到指定的对象！")
		}
	},
}

func init() {
	rootCmd.AddCommand(putCmd)
}
