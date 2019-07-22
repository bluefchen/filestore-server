package route

import (
	"github.com/gin-gonic/gin"
	"filestore-server/service/upload/api"
	"github.com/gin-contrib/cors"
)

func Router() *gin.Engine {
	//gin framework ,包括logger,Recovery
	router := gin.Default()

	//处理静态资源
	router.Static("/static/", "../../../static")

	// 使用gin插件支持跨域请求
	router.Use(cors.New(cors.Config{
		AllowOrigins:  []string{"*"}, // []string{"http://localhost:8080"},
		AllowMethods:  []string{"GET", "POST", "OPTIONS"},
		AllowHeaders:  []string{"Origin", "Range", "x-requested-with", "content-Type"},
		ExposeHeaders: []string{"Content-Length", "Accept-Ranges", "Content-Range", "Content-Disposition"},
		// AllowCredentials: true,
	}))

	//上传文件
	router.GET("/file/upload", api.UploadHandler)
	router.POST("/file/upload", api.DoUploadHandler)

	router.OPTIONS("/file/upload", func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "POST,OPTIONS")
		c.Status(204)
	})

	return router
}
