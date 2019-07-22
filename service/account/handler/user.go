package handler

import (
	"context"
	"filestore-server/common"
	"filestore-server/config"
	dblayer "filestore-server/db"
	"filestore-server/service/account/proto"
	"filestore-server/util"
	"fmt"
	"time"
)

type User struct {
}

func GenToken(username string) string {
	// 40位字符:md5(username+timestamp+token_salt)+timestamp[:8]
	ts := fmt.Sprintf("%x", time.Now().Unix())
	tokenPrefix := util.MD5([]byte(username + ts + "_tokensalt"))
	return tokenPrefix + ts[:8]
}

func (this *User) SignUp(ctx context.Context, req *proto.ReqSignUp, resp *proto.RespSignUp) error {
	fmt.Println("user sign up account")
	username := req.Username
	password := req.Password
	fmt.Println(username + password)
	if len(username) < 3 || len(password) < 5 {
		resp.Code = common.StatusParamInvalid
		resp.Message = "注册参数无效"
		return nil
	}
	// 对密码进行加盐及取Sha1值加密
	encPasswd := util.Sha1([]byte(password + config.PasswordSalt))
	// 将用户信息注册到用户表中
	suc := dblayer.UserSignup(username, encPasswd)
	if suc {
		resp.Code = common.StatusOK
		resp.Message = "注册成功"
	} else {
		resp.Code = common.StatusRegisterFailed
		resp.Message = "注册失败"
	}
	return nil
}

func (this *User) SignIn(ctx context.Context, req *proto.ReqSignIn, resp *proto.RespSignIn) error {
	fmt.Println("user sign in account")

	username := req.Password
	password := req.Password
	encPassword := util.Sha1([]byte(password + config.PasswordSalt))
	suc := dblayer.UserSignin(username, encPassword)
	if !suc {
		resp.Code = common.StatusLoginFailed
		resp.Message = "用户名或密码错误，登陆失败"
		return nil
	}
	token := GenToken(username)

	updateSuc := dblayer.UpdateToken(username, token)
	if !updateSuc {
		resp.Code = common.StatusServerError
		return nil
	}

	resp.Code = common.StatusOK
	resp.Message = "登陆成功"
	resp.Token = token

	return nil
}

func UserInfo(ctx context.Context, req *proto.ReqUserInfo, resp *proto.RespUserInfo) error {
	username := req.Username
	_, err := dblayer.GetUserInfo(username)
	if err != nil {
		resp.Code = common.StatusUserNotExists
		resp.Message = "未查询到" + username
	}





	return nil
}
