package gpu

import (
	"bufio"
	"bytes"
	"encoding/json"
	"io"
	"k8s.io/klog/v2"
	"minik8s.com/minik8s/config"
	v1 "minik8s.com/minik8s/pkg/api/v1"
	"net/http"
	"os"
	"strconv"
	"time"
)

var job *v1.GPUJob

type GetGPUResponse struct {
	Key  string    `json:"key"`
	Job  v1.GPUJob `json:"value"`
	Type string    `json:"type"`
}

func Run() {
	getJob()
	runJob()
	for {
		time.Sleep(time.Second * 3)
		getResult()
		if !isJobRunning() {
			break
		}
	}
}

func getJob() {
	url := os.Args[1] + "/gpu/" + os.Args[2]
	resp, err := http.Get(url)
	if err != nil {
		return
	}

	buf, err := io.ReadAll(resp.Body)
	if err != nil {
		klog.Error(err)
		return
	}

	job = new(v1.GPUJob)
	var gpuResp GetGPUResponse
	err = json.Unmarshal(buf, &gpuResp)
	if err != nil {
		klog.Error(err)
		return
	}
	job = &gpuResp.Job
}

func downloadFile(filename string, jobUID string) {
	resp, err := http.Get(os.Args[1] + "/public/" + jobUID + "-" + filename)
	if err != nil {
		klog.Error(err)
		return
	}
	defer resp.Body.Close()
	reader := bufio.NewReader(resp.Body) // 获得reader对象
	file, err := os.Create("./" + filename)
	if err != nil {
		klog.Error(err)
		return
	}

	writer := bufio.NewWriter(file) // 获得writer对象
	_, err = io.Copy(writer, reader)
	if err != nil {
		klog.Error(err)
		return
	}
}

func runJob() {
	downloadFile(job.Script, job.UID)
	for _, file := range job.Files {
		downloadFile(file.Filename, job.UID)
	}

	sshClient := NewSshClient(config.AS_GPU_DATA_ADDR)
	sshClient.UploadFile("./"+job.Script, config.AS_GPU_HOMEPATH+job.Script)
	for _, up_file := range job.Files {
		sshClient.UploadFile("./"+up_file.Filename, config.AS_GPU_HOMEPATH+up_file.Filename)
	}
	sshClient.Close()

	sshClient = NewSshClient(config.AS_GPU_LOGIN_ADDR)
	res := sshClient.RunCmd("sbatch " + config.AS_GPU_HOMEPATH + job.Script)
	klog.Infof("sbatch result: %v", res)
	job.JobNum = string(res[len(res)-9 : len(res)-1])

	sshClient.Close()
}

func getResult() {
	sshClient := NewSshClient(config.AS_GPU_LOGIN_ADDR)
	job.Output = string(sshClient.RunCmd("cat " + config.AS_GPU_HOMEPATH + job.JobNum + ".out"))
	job.Error = string(sshClient.RunCmd("cat " + config.AS_GPU_HOMEPATH + job.JobNum + ".err"))
	klog.Infof("get result: \nOutput: %s\n Error: %s", job.Output, job.Error)

	cli := http.Client{}
	buf, _ := json.Marshal(job)
	url := config.AC_ServerAddr + ":" + strconv.Itoa(config.AC_ServerPort) + config.AC_RestGpu_Path
	req, _ := http.NewRequest(http.MethodPut, url+"/"+job.UID, bytes.NewReader(buf))
	_, _ = cli.Do(req)

	sshClient.Close()
}

func isJobRunning() bool {
	sshClient := NewSshClient(config.AS_GPU_LOGIN_ADDR)
	res := sshClient.RunCmd("squeue | grep " + job.JobNum)
	return len(res) > 5
}
