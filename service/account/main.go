package main

import (
	"filestore-server/service/account/handler"
	"filestore-server/service/account/proto"
	"github.com/micro/go-micro"
	"log"
)

func main() {

	/*reg := consul.NewRegistry(func(op *registry.Options) {
		op.Addrs = []string{
			"127.0.0.1:8500",
		}
	})*/

	service := micro.NewService(
		//micro.Registry(reg),
		micro.Name("go.micro.service.user"),
	)

	service.Init()
	proto.RegisterUserServiceHandler(service.Server(), new(handler.User))

	err := service.Run()
	if err != nil {
		log.Println("account service error:", err.Error())
	}
}
