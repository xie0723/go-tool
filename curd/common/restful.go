package common

// 标准的 rest 返回接口，字符小写化
type StandRestResult struct {
	Code    int         `json:"code"`    // 0 表示成功，其他失败
	Message string      `json:"message"` // 错误信息
	Data    interface{} `json:"data"`    // 数据体
}

// RestResult Rest接口返回值(兼容特性使用)
type RestResult struct {
	Code    int         // 0 表示成功，其他失败
	Message string      // 错误信息
	Data    interface{} // 数据体
}

func (rest StandRestResult) GetStandRestResult(code int, msg string, data interface{}) interface{} {
	return StandRestResult{Code: code, Message: msg, Data: data}
}

// 返回接口
type StandRestResultInf interface {
	GetStandRestResult(code int, msg string, data interface{}) interface{}
}
