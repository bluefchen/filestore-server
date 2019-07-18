package handler

import (
	"filestore-server/common"
	"filestore-server/util"
	"github.com/gin-gonic/gin"
	"net/http"
)

// HTTPInterceptor : http请求拦截器
func HTTPInterceptor() gin.HandlerFunc {
	return func(c *gin.Context) {
		username := c.Request.FormValue("username")
		token := c.Request.FormValue("token")

		//验证登录token是否有效
		if len(username) < 3 || !IsTokenValid(username, token) {
			// token校验失败则跳转到直接返回失败提示
			// 通知后面的方法不再执行
			c.Abort()
			//TODO
			resp := util.NewRespMsg(
				int(common.StatusTokenInvalid),
				"token无效",
				nil,
			)
			c.JSON(http.StatusOK,resp)
			return
		}
		// 继续执行
		c.Next()
	}
}
