package main

import (
	"filestore-server/service/apigw/route"
)

func main() {
	router := route.Router()
	router.Run(":8080")
}
