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
type GetNodeResponse struct {
	Key  string  `json:"key"`
	Node v1.Node `json:"value"`
	Type string  `json:"type"`
}
type GetServiceResponse struct {
	Key     string     `json:"key"`
	Service v1.Service `json:"value"`
	Type    string     `json:"type"`
}
type GetDNSResponse struct {
	Key  string `json:"key"`
	DNS  v1.DNS `json:"value"`
	Type string `json:"type"`
}
type GetEndpointResponse struct {
	Key      string      `json:"key"`
	Endpoint v1.Endpoint `json:"value"`
	Type     string      `json:"type"`
}
type GetReplicaResponse struct {
	Key        string        `json:"key"`
	ReplicaSet v1.ReplicaSet `json:"value"`
	Type       string        `json:"type"`
}
type GetHpaResponse struct {
	Key        string        `json:"key"`
	HPA v1.HorizontalPodAutoscaler `json:"value"`
	Type       string        `json:"type"`
}
type GetFunctionResponse struct {
	Funcitons []string `json:"functions"`
	Error       string        `json:"error"`
}
type GetActionChainResponse struct {
	Funcitons []string `json:"functions"`
	Error       string        `json:"error"`
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
