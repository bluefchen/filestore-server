package handler

import (
	"filestore-server/common"
	"fmt"
	"github.com/gin-gonic/gin"

	// "io/ioutil"
	"net/http"
	"time"

	dblayer "filestore-server/db"
	"filestore-server/util"
)

const (
	// 用于加密的盐值(自定义)
	pwdSalt = "*#890"
)

// SignupHandler : 用户注册请求页面跳转
func SignupHandler(c *gin.Context) {
	c.Redirect(http.StatusFound, "/static/view/signup.html")
}

//DoSignUpHandler: 处理注册post请求
func DoSignUpHandler(c *gin.Context) {
	username := c.Request.FormValue("username")
	password := c.Request.FormValue("password")
	if len(username) < 3 || len(password) < 5 {
		c.JSON(http.StatusOK, gin.H{
			"msg":  "Invalid parameter",
			"code": -1,
		})
		return
	}
	// 对密码进行加盐及取Sha1值加密
	encPasswd := util.Sha1([]byte(password + pwdSalt))
	// 将用户信息注册到用户表中
	suc := dblayer.UserSignup(username, encPasswd)
	if suc {
		c.JSON(http.StatusOK, gin.H{
			"msg":  "SignUp success",
			"code": common.StatusOK,
		})
	} else {
		c.JSON(http.StatusOK, gin.H{
			"msg":  "SignUp failed",
			"code": common.StatusRegisterFailed,
		})
	}
	return
}

// SignInHandler : 登录接口
func SignInHandler(c *gin.Context) {
	c.Redirect(http.StatusFound, "/static/view/signin.html")
}

// DoSignInHandler：处理用户登陆请求
func DoSignInHandler(c *gin.Context) {
	username := c.Request.FormValue("username")
	password := c.Request.FormValue("password")

	encPasswd := util.Sha1([]byte(password + pwdSalt))

	// 1. 校验用户名及密码
	pwdChecked := dblayer.UserSignin(username, encPasswd)
	if !pwdChecked {
		c.JSON(http.StatusOK, gin.H{
			"msg":  "failed login",
			"code": common.StatusLoginFailed,
		})
		return
	}

	// 2. 生成访问凭证(token)
	token := GenToken(username)
	upRes := dblayer.UpdateToken(username, token)
	if !upRes {
		c.JSON(http.StatusOK, gin.H{
			"msg":  "failed login",
			"code": common.StatusLoginFailed,
		})
		return
	}

	// 3. 登录成功后重定向到首页
	resp := util.RespMsg{
		Code: int(common.StatusOK),
		Msg:  "OK",
		Data: struct {
			Location string
			Username string
			Token    string
		}{
			Location: "/static/view/home.html",
			Username: username,
			Token:    token,
		},
	}
	c.Data(http.StatusOK, "octet-stream", resp.JSONBytes())
	//c.Data(http.StatusOK, "application/json", resp.JSONBytes())
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

// GenToken : 生成token
func GenToken(username string) string {
	// 40位字符:md5(username+timestamp+token_salt)+timestamp[:8]
	ts := fmt.Sprintf("%x", time.Now().Unix())
	tokenPrefix := util.MD5([]byte(username + ts + "_tokensalt"))
	return tokenPrefix + ts[:8]
}

// IsTokenValid : token是否有效
func IsTokenValid(username string, token string) bool {
	if len(token) != 40 {
		return false
	}
	// TODO: 判断token的时效性，是否过期
	queryToken := dblayer.QueryUserToken(username)
	if queryToken != token {
		return false
	}
	return true
}
