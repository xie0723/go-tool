package customerror

import "fmt"

type CustomError interface {
	Error() string
	GetCode() int
	GetMessage() string
}

// 自定义错误对象
type customErr struct {
	Code    int    // 错误码
	Message string // 错误消息
}

// 实例化一个错误对象
func New(code int, message string) CustomError {
	return &customErr{code, message}
}

// 获取错误消息，也是实现error接口
func (e *customErr) Error() string {
	return fmt.Sprintf("%d::%s", e.Code, e.Message)
}

// 获取错误消息
func (e *customErr) GetCode() int {
	return e.Code
}

// 获取错误码
func (e *customErr) GetMessage() string {
	return e.Message
}
