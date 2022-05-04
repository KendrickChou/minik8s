package commands

import (
	"fmt"
	"github.com/spf13/cobra"
	v1 "minik8s.com/minik8s/pkg/api/v1"
)

type StatusResponse struct {
	Status string `json:"status"`
	Id     string `json:"id,omitempty"`
	Error  string `json:"error,omitempty"`
}
type GetPodResponse struct {
	Key  string `json:"key"`
	Pod  v1.Pod `json:"value"`
	Type string `json:"type"`
}

var rootCmd = &cobra.Command{
	Use:   "newApp",
	Short: "minik8s的命令行工具",
	Long:  `欢迎使用minik8s的命令行工具！`,

	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("111")
	},
}

func Execute() {
	cobra.CheckErr(rootCmd.Execute())
}
