package commands

import (
	"fmt"
	"github.com/spf13/cobra"
	"minik8s.com/minik8s/pkg/apiclient"
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
			resp := apiclient.Rest(id, []byte{}, apiclient.OBJ_POD, apiclient.OP_DELETE)
			fmt.Printf("%s\n", resp)
		case 'S':
			resp := apiclient.Rest(id, []byte{}, apiclient.OBJ_SERVICE, apiclient.OP_DELETE)
			fmt.Printf("%s\n", resp)
		case 'R':
			resp := apiclient.Rest(id, []byte{}, apiclient.OBJ_REPLICA, apiclient.OP_DELETE)
			fmt.Printf("%s\n", resp)
		default:
			fmt.Println("找不到指定的对象！")
		}
	},
}

func init() {
	delCmd.Flags().StringP("id", "i", "X", "指定对象id")

	rootCmd.AddCommand(delCmd)
}
