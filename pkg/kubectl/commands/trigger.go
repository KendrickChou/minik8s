package commands

import (
	"fmt"
	"github.com/spf13/cobra"
	"minik8s.com/minik8s/pkg/apiclient"
)

var triggerCmd = &cobra.Command{
	Use:   "trigger",
	Short: `触发动作链`,
	Long:  `用于触发动作链，需要指定动作链名字`,
	Run: func(cmd *cobra.Command, args []string) {
		id, err := cmd.Flags().GetString("id")
		if err != nil {
			fmt.Println("getString err: ", err)
			return
		}
		arg, err := cmd.Flags().GetString("arg")
		if err != nil {
			fmt.Println("getString err: ", err)
			return
		}

		resp := apiclient.Rest(id, arg, apiclient.OBJ_TRIGGER, apiclient.OP_GET)
		fmt.Printf("%s\n", resp)

	},
}

func init() {
	triggerCmd.Flags().StringP("id", "i", "X", "指定动作链名字")
	triggerCmd.Flags().StringP("arg", "a", "", "指定入口参数")
	

	rootCmd.AddCommand(triggerCmd)
}

