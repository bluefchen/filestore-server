package main

import (
	"filestore-server/common"
	"filestore-server/service/account/handler"
	"filestore-servser/service/account/proto"
	"github.com/micro/go-micro"
	"github.com/micro/go-micro/registry"
	"github.com/micro/go-micro/registry/consul"
	"log"
	"time"
)

func main() {

	reg := consul.NewRegistry(func(op *registry.Options) {
		op.Addrs = []string{
			"127.0.0.1:8500",
		}
	})

	//	创建服务
	service := micro.NewService(
		micro.Registry(reg),
		//service := k8s.NewService(
		micro.Name("go.micro.service.user"),
		micro.RegisterTTL(time.Second*10),
		micro.RegisterInterval(time.Second*5),
		micro.Flags(common.CustomFlags...),
	)
	service.Init()
	proto.RegisterUserServiceHandler(service.Server(),new(handler.User))
	err:= service.Run()
	if err!= nil {
		log.Println("service run failed,err:",err.Error())
	}
}
