package handler

import (
	"context"
	"filestore-server/common"
	cfg "filestore-server/config"
	dblayer "filestore-server/db"
	"filestore-server/service/account/proto"
	"filestore-server/util"
)

type User struct {
}

// Signup：处理用户注册请求
func (user *User)Signup(ctx context.Context, req *proto.ReqSignup, resp *proto.RespSignup) error {
	username := req.Username
	password := req.Password
	if len(username) < 3 || len(password) < 5 {
		resp.Code = common.StatusParamInvalid
		resp.Message = "注册参数无效"
		return nil
	}
	// 对密码进行加盐及取Sha1值加密
	encPasswd := util.Sha1([]byte(password + cfg.PasswordSalt))
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
