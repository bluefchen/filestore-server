package handler

import (
	"context"
	"filestore-server/common"
	"filestore-server/config"
	"filestore-server/service/account/proto"
	"filestore-server/util"
	"github.com/gin-gonic/gin"
	"github.com/micro/go-micro"
	"github.com/micro/go-micro/registry"
	"github.com/micro/go-micro/registry/consul"
	"log"
	"net/http"
)

var (
	userCli proto.UserService
)

func init() {
	reg := consul.NewRegistry(func(op *registry.Options) {
		op.Addrs = []string{
			"127.0.0.1:8500",
		}
	})
	service := micro.NewService(
		micro.Registry(reg),
	)

	service.Init()

	userCli = proto.NewUserService("go.micro.service.user", service.Client())

}

// SignupHandler : 用户注册请求页面跳转
func SignupHandler(c *gin.Context) {
	c.Redirect(http.StatusFound, "/static/view/signup.html")
}

//DoSignUpHandler: 处理注册post请求
func DoSignUpHandler(c *gin.Context) {
	username := c.Request.FormValue("username")
	password := c.Request.FormValue("password")
	respSignUp, err := userCli.SignUp(context.TODO(), &proto.ReqSignUp{
		Username: username,
		Password: password,
	})
	if err != nil {
		c.Status(http.StatusInternalServerError)
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"code": respSignUp.Code,
		"msg":  respSignUp.Message,
	})
}

//SignInHandler：登陆页面跳转
func SignInHandler(c *gin.Context) {
	c.Redirect(http.StatusFound, "/static/view/signin.html")
}

//DoSignInHandler：处理登陆请求
func DoSignInHandler(c *gin.Context) {
	username := c.Request.FormValue("username")
	password := c.Request.FormValue("password")

	respSignIn, err := userCli.SignIn(context.TODO(), &proto.ReqSignIn{
		Username: username,
		Password: password,
	})
	if err != nil {
		log.Println(err.Error())
		c.Status(http.StatusInternalServerError)
		return
	}

	if respSignIn.Code != common.StatusOK {
		c.JSON(200, gin.H{
			"msg":  "登录失败",
			"code": respSignIn.Code,
		})
		return
	}

	resp := util.RespMsg{
		Code: int(common.StatusOK),
		Msg:  "OK",
		Data: struct {
			Location      string
			Username      string
			Token         string
			UploadEntry   string
			DownloadEntry string
		}{
			Location:      "/static/view/home.html",
			Username:      username,
			Token:         respSignIn.Token,
			UploadEntry:   config.UploadLBHost,
			DownloadEntry: config.DownloadLBHost,
		},
	}
	//c.Data(http.StatusOK, "octet-stream", resp.JSONBytes())
	c.Data(http.StatusOK, "application/json", resp.JSONBytes())
}

// UserInfoHandler ： 查询用户信息
func UserInfoHandler(c *gin.Context) {
	username := c.Request.FormValue("username")

	// 3. 查询用户信息
	user, err := dblayer.GetUserInfo(username)
	if err != nil {
		c.JSON(http.StatusForbidden, gin.H{})
		return
	}

	// 4. 组装并且响应用户数据
	resp := util.RespMsg{
		Code: 0,
		Msg:  "OK",
		Data: user,
	}
	c.Data(http.StatusOK, "octet-stream", resp.JSONBytes())
}
