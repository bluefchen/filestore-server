package main

import (
	"filestore-server/common"
	cfg "filestore-server/service/upload/config"
	"filestore-server/service/upload/route"
	"fmt"
	"github.com/micro/go-micro"
	"time"
	proto "filestore-server/service/upload/proto"
	uprpc "filestore-server/service/upload/rpc"
)

func startAPIService() {
	router := route.Router()
	router.Run(cfg.UploadServiceHost)
}

func startRpcService() {
	service := micro.NewService(
		micro.Name("go.micro.service.upload"), // 服务名称
		micro.RegisterTTL(time.Second*10),     // TTL指定从上一次心跳间隔起，超过这个时间服务会被服务发现移除
		micro.RegisterInterval(time.Second*5), // 让服务在指定时间内重新注册，保持TTL获取的注册时间有效
		micro.Flags(common.CustomFlags...),
	)

	service.Init()

	proto.RegisterUploadServiceHandler(service.Server(), new(uprpc.Upload))
	err := service.Run()
	if err != nil {
		fmt.Println(err.Error())
	}
}
func main() {
	go startAPIService()
	startRpcService()
}
