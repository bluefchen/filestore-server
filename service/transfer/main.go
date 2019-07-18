package main

import (
	"bufio"
	"encoding/json"
	"filestore-server/config"
	dblayer "filestore-server/db"
	"filestore-server/mq"
	"filestore-server/store/oss"
	"filestore-server/util"
	"log"
	"os"
	"strings"
)

// ProcessTransfer : 处理文件转移
func ProcessTransfer(msg []byte) bool {
	log.Println(string(msg))

	pubData := mq.TransferData{}
	err := json.Unmarshal(msg, &pubData)
	if err != nil {
		log.Println(err.Error())
		return false
	}

	path := util.GetCurrentPath() + strings.Split(pubData.CurLocation, "..")[2]
	fin, err := os.Open(path)
	if err != nil {
		log.Println(err.Error())
		return false
	}

	//将文件上传到oss
	err = oss.Bucket().PutObject(
		pubData.DestLocation,
		bufio.NewReader(fin))
	if err != nil {
		log.Println(err.Error())
		return false
	}
	//
	_ = dblayer.UpdateFileLocation(
		pubData.FileHash,
		pubData.DestLocation)
	return true
}

func main() {
	if !config.AsyncTransferEnable {
		log.Println("异步转移文件功能目前被禁用，请检查相关配置")
		return
	}
	log.Println("文件转移服务启动中，开始监听转移任务队列...")
	//开启消费
	mq.StartConsume(
		config.TransOSSQueueName,
		"transfer_oss",
		ProcessTransfer,
	)
}
