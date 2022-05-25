package commands

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/spf13/cobra"
	"io"
	"io/ioutil"
	"mime/multipart"
	"minik8s.com/minik8s/config"
	v1 "minik8s.com/minik8s/pkg/api/v1"
	"minik8s.com/minik8s/pkg/apiclient"
	"net/http"
	"os"
	"strconv"
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
		case "gpu":
			var gpuJob v1.GPUJob
			err := json.Unmarshal(buf, &gpuJob)
			if err != nil {
				fmt.Println("输入文件解析失败: ", err)
				return
			}
			resp = apiclient.Rest("", string(buf), apiclient.OBJ_GPU, apiclient.OP_POST)
			var stat StatusResponse
			err = json.Unmarshal(resp, &stat)
			if err != nil {
				fmt.Println("服务器返回信息无效: ", err)
			} else if stat.Status != "OK" {
				fmt.Println("创建对象失败：", stat.Error)
			} else {
				fmt.Println("成功创建对象，id：", stat.Id)
				uploadFile(gpuJob.Script, stat.Id+"-"+gpuJob.Script)
				for _, up_file := range gpuJob.Files {
					uploadFile(up_file.Filename, stat.Id+"-"+up_file.Filename)
				}
			}
			return
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

func uploadFile(path string, newFilename string) {
	upfile, err := os.OpenFile(path, os.O_RDONLY, 0666)
	if err != nil {
		fmt.Println("文件打开失败：", err)
	}
	defer upfile.Close()

	url := config.AC_ServerAddr + ":" + strconv.Itoa(config.AC_ServerPort) + "/upload"

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, err := writer.CreateFormFile("file", newFilename)
	if err != nil {
		fmt.Println("文件上传失败：", err)
		return
	}
	_, err = io.Copy(part, upfile)

	err = writer.Close()
	if err != nil {
		fmt.Println("文件上传失败：", err)
		return
	}
	request, err := http.NewRequest("POST", url, body)
	request.Header.Set("Content-Type", writer.FormDataContentType())

	client := &http.Client{}
	resp, err := client.Do(request)
	defer resp.Body.Close()
	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("文件上传失败：", err)
		return
	}
	fmt.Println(string(respBody))
}
