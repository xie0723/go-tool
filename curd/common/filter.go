package common

import (
	"github.com/astaxie/beego/context"
)

// 解析token，存储username
func TokenDealFilter(ctx *context.Context) {
	// var err error
	// var username string
	// if username, err = funcs.GetAccountIdFromToken(ctx.Input.Header(TokenKey), []byte("test_group")); err != nil {
	// 	logs.Error("GetAccountIdFromToken failed, err is ", err.Error())
	// 	return
	// }
	// ctx.Input.CruSession.Set(UserNameSessionKey, username)
}

// 通过黑鲨账号认证
func BSAuth(ctx *context.Context) {
	// 获取TOKEN，比较和session中的token是否一样
	// 如果不一样或者session中token为空，向黑鲨账号认证，认证通过就通过，不通过就跳转到认证失败页面
	// 通过token获取用户信息
	// https://account.blackshark.com/user/profile?access_token=55dbc109072942afb6c53ad1ec46f37d&client_id=2000011111
}
