package common

import "github.com/daimall/tools/curd/customerror"

var (
	//NotFound not found
	NotFound = customerror.New(404, "NOT_FOUND")
	//ErrServer 服务器错误
	ErrServer = customerror.New(500, "服务器错误")
	//SignError sign error
	SignError = customerror.New(1103, "Sign签名错误")

	//ParamsError json error
	ParamsError = customerror.New(1101, "参数格式错误")
	// Validate error
	ParamsValidateError = customerror.New(1102, "参数校验错误")
	//InternalServerError 500
	InternalServerError = customerror.New(500, "INTERNAL_SERVER_ERROR")

	//ParamsFormatError 参数格式错误
	ParamsFormatError = customerror.New(1102, "PARAMS_FORMAT_ERROR")

	//LoginTokenInvalid  登录token失效
	LoginTokenInvalid = customerror.New(1104, "LOGIN_TOKEN_INVALID")

	//AppIDError AppIDError
	AppIDError = customerror.New(1105, "AppID Error")
	//TimestampError TimestampError
	TimestampError = customerror.New(1106, "Timestamp Error")

	//AccountError 账号或密码错误
	AccountError = customerror.New(2102, "ACCOUNT_ERROR")

	// TOKEN 不存在
	TokenNotFound = customerror.New(3000, "Token信息不存在")

	//AccessTokenExpire 帐号app的token过期
	AccessTokenExpire = customerror.New(3001, "AccessToken Error or Expire")
	//RefreshTokenExpire 帐号app的刷新令牌过期
	RefreshTokenExpire = customerror.New(3002, "RefreshToken Error or Expire")
	TokenInvalid       = customerror.New(3003, "非法的token")
	// 用户认证信息不存在
	UnameNotFound = customerror.New(3005, "用户认证信息不存在")
	//OauthCodeExpire accesstoken过期
	OauthCodeExpire = customerror.New(3006, "Oauth Code Expire")
	//
	AuthStateInvalid = customerror.New(3007, "Auth state 不正确")
	//
	AuthCodeInvalid = customerror.New(3008, "Auth code 不正确")
	//
	UserNameOrPasswordInvalid = customerror.New(5009, "用户名或者密码错误")
	// LdapErr 域账号登陆异常
	LdapErr = customerror.New(5005, "域账号登陆异常")
	// DisabledUser 禁用的用户
	DisabledUser       = customerror.New(5006, "禁用的用户")
	AnotherClientLogin = customerror.New(50008, "其他客户端登录了")
	//GetCodeFrequently 请求验证码次数过多
	GetCodeFrequently = customerror.New(3009, "请求验证码次数过多")
	//VerificationCodeeError 验证码错误
	VerificationCodeeError = customerror.New(3010, "验证码错误")

	//ImageError 图片格式不支持
	ImageError = customerror.New(4001, "该图片格式不支持")
	//NicknameError 昵称错误
	NicknameError = customerror.New(4002, "昵称包含禁用词")
	// UserAlreadyExist 创建的用户已经存在
	UserAlreadyExist = customerror.New(5001, "用户已存在")
	// UserDoesNotExist 用户不存在
	UserDoesNotExist = customerror.New(5002, "用户不存在")
	// WrongPassword 密码错误
	WrongPassword = customerror.New(5003, "密码错误")
)

var (
	ServerErr            = customerror.New(500, "服务器错误")
	QueryCondErr         = customerror.New(1001, "查询条件不正确")
	StepTypeNotFound     = customerror.New(1002, "Flow Step 不存在")
	ServiceIdNotInt      = customerror.New(1003, "Service Id格式不正确")
	ParamsErr            = customerror.New(1004, "参数错误")
	UploadErr            = customerror.New(1005, "上传文件失败")
	UpdateActionNotFound = customerror.New(1006, "Update Acton 不存在")
	ActionNotFound       = customerror.New(1007, "Action 不存在")
)
