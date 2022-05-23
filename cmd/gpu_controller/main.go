package main

import (
	"fmt"
	"minik8s.com/minik8s/pkg/gpu"
)

func main() {
	cli := gpu.Cli{
		Addr: "127.0.0.1:22",
		User: "root",
		Pwd:  "123456",
	}
	// 建立连接对象
	c, _ := cli.Connect()
	// 退出时关闭连接
	defer c.Client.Close()
	res, _ := c.Run("ls")
	res1, _ := c.Run("pwd")
	fmt.Println(res)
	fmt.Println(res1)

}
