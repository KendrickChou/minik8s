package gpu

import (
	"fmt"
	"github.com/pkg/sftp"
	"golang.org/x/crypto/ssh"
	"io"
	"k8s.io/klog"
	"log"
	"minik8s.com/minik8s/config"
	"net"
	"os"
	"time"
)

type SshClient struct {
	username string
	password string
	host     string
	Client   *ssh.Client
}

func NewSshClient(host string) *SshClient {
	sshClient := new(SshClient)
	sshClient.username = config.AS_GPU_USERNAME
	sshClient.password = config.AS_GPU_PASSWD
	sshClient.host = host

	auth := make([]ssh.AuthMethod, 0)
	auth = append(auth, ssh.Password(sshClient.password))

	clientConfig := &ssh.ClientConfig{
		User:    sshClient.username,
		Auth:    auth,
		Timeout: 30 * time.Second,
		HostKeyCallback: func(hostname string, remote net.Addr, key ssh.PublicKey) error {
			return nil
		},
	}

	addr := sshClient.host + ":22"
	client, err := ssh.Dial("tcp", addr, clientConfig) //连接ssh
	if err != nil {
		log.Fatal("连接ssh失败", err)
	}
	sshClient.Client = client

	return sshClient
}

func (sshClient *SshClient) RunCmd(cmd string) []byte {
	session, err := sshClient.Client.NewSession()
	if err != nil {
		panic(err)
	}
	defer session.Close()

	runResult, err := session.CombinedOutput(cmd)
	if err != nil {
		return []byte{}
	}
	return runResult
}

func (sshClient *SshClient) UploadFile(localPath string, remoteDir string) {
	ftpClient, err := sftp.NewClient(sshClient.Client)
	if err != nil {
		klog.Error("创建ftp客户端失败", err)
		panic(err)
	}

	defer ftpClient.Close()

	fmt.Println(localPath, remoteDir)
	srcFile, err := os.Open(localPath)
	if err != nil {
		klog.Error("打开文件失败", err)
		panic(err)
	}
	defer srcFile.Close()

	dstFile, e := ftpClient.Create(remoteDir)
	if e != nil {
		klog.Error("创建文件失败", e)
		panic(e)
	}
	defer dstFile.Close()

	buffer := make([]byte, 1024000)
	for {
		n, err := srcFile.Read(buffer)
		dstFile.Write(buffer[:n])
		if err != nil {
			if err == io.EOF {
				klog.Info("文件传输结束")
				break
			} else {
				klog.Error("文件传输出错", err)
				panic(err)
			}
		}
	}
}

func (sshClient *SshClient) Close() {
	err := sshClient.Client.Close()
	if err != nil {
		klog.Errorf("close err: %v", err)
		return
	}
}
