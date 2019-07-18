package route

import (
	"filestore-server/handler"
	"github.com/gin-gonic/gin"
)

func Router() *gin.Engine {
	//gin framework ,包括logger,Recovery
	router := gin.Default()

	//处理静态资源
	router.Static("/static/", "../../static")

	//不需要验证就能访问的接口
	//用户注册
	router.GET("/user/signup", handler.SignupHandler)
	router.POST("/user/signup", handler.DoSignUpHandler)
	//用户登陆
	router.GET("/user/signin", handler.SignInHandler)
	router.POST("/user/signin", handler.DoSignInHandler)

	// 加入中间件，用于校验token的拦截器，下方的代码将全部走拦截器
	router.Use(handler.HTTPInterceptor())

	//上传文件
	router.GET("/file/upload", handler.UploadHandler)
	router.POST("/file/upload", handler.DoUploadHandler)
	//获取文件源信息
	router.GET("/file/meta",handler.GetFileMetaHandler)
	// 根据用户名查询文件批量信息
	router.POST("/file/query",handler.FileQueryHandler)
	//下载文件
	router.POST("/file/download",handler.DownloadHandler)
	//修改文件信息
	router.POST("/file/update",handler.FileMetaUpdateHandler)
	//删除文件
	router.GET("/file/delete",handler.FileDeleteHandler)
	//秒传接口
	router.POST("/file/fastupload",handler.TryFastUploadHandler)
	//获取下载链接
	router.POST("/file/downloadurl",handler.DownloadURLHandler)

	// 获取用户信息
	router.POST("/user/info",handler.UserInfoHandler)

	return router
}
