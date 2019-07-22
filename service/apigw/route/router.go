package route

import (
	"filestore-server/service/apigw/handler"
	"github.com/gin-gonic/gin"
)

func Router() *gin.Engine {
	router := gin.Default()

	router.Static("/static/", "/Users/darius/Documents/WorkSpace/GoProject/src/filestore-server/static")

	//用户注册
	router.GET("/user/signup", handler.SignupHandler)
	router.POST("/user/signup", handler.DoSignUpHandler)

	//用户登陆
	router.GET("/user/signin",handler.SignInHandler)
	router.POST("/user/signin",handler.DoSignInHandler)



	//获取用户信息
	router.POST("/user/info",handler.UserInfoHandler)
	return router
}
