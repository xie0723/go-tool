package flowservice

import (
	"io"
	"reflect"

	"github.com/daimall/tools/curd/common"
	"gorm.io/gorm"
)

// 更新字段接口
type Action interface {
	Do(uname string, serviceId uint, actionType string, c common.BaseController) (ret interface{}, flowId uint, oplog string, err error)
}

// flow handler 接口
type FlowHandler interface {
	// 保存用户的处理结果
	Do(uname string, c common.BaseController) (ret interface{}, oplog string, err error)
	// 从数据库中加载数据，返回一个新实例
	LoadInst(flow FlowService, uname string, id uint) (handler FlowHandler, err error)
	LoadStepHandlers(tx *gorm.DB, flowId uint, stepKey string) (handlers []FlowHandler, err error)
	GetConclusion() bool // 是否通过
	IsFinish() bool      // 是否完成审核
	TableName() string   // 表名
}

// flow step 接口
type FlowStep interface {
	// 唯一标识
	Key() (stepKey string)
	// 通过标准，100 表示所有人结论都要是通过
	PassRate() (rate int)
	// 找到注册handle
	Hander() (handler FlowHandler)
	// 流程步骤等操作信息
	LoadHandlers(tx *gorm.DB, flowId uint) (handlers []FlowHandler, err error)
	// handlers 入库
	AddHandlers(tx *gorm.DB, handlers []FlowHandler) (err error)
	// 退回上一步，清理handlers
	ClearHandlers(tx *gorm.DB, flowId uint, flow string, steps []string, remainingIds []uint) (err error)
	// 从flow属性中获取责任人信息，多个逗号分割，作为步骤初始化时创建handers
	GetDefaultHandlers(flow FlowService) (handers string)
	// 获取处理当前步骤所需要对参数对象（供前端使用）
	GetConfigs(flow FlowService) (ret interface{})
}

//FlowService 流程接口，支持各种Service定制
type FlowService interface {
	GetID() uint
	// 获取flowname
	GetFlowName() string
	// 新建流程
	New(uname string, c common.BaseController) (flowId uint, ret interface{}, oplog string, err error)
	// make新实例
	NewInst() (flowService FlowService)
	// 获取新实例（从数据库中加载初始值）
	LoadInst(flowId uint) (flowService FlowService, err error)
}

type ActionInf interface {
	// 获取一个自定义动作
	GetAction(serviceId uint, actionType string) (action Action, err error)
}

type GetOneInf interface {
	// 获取一个对象详情
	GetOne(uname string, id uint) (ret interface{}, oplog string, err error)
}

type GetAllInf interface {
	// 获取所有流程
	GetAll(uname string, query []*common.QueryConditon, fields []string,
		sortby []string, order []string, offset int,
		limit int) (ret interface{}, count int64, oplog string, err error)
}

type UpdateInf interface {
	// 更新对象
	Update(flowid uint, fields []string, c common.BaseController) (ret interface{}, oplog string, err error) // 刷新流程基础信息

}

// 兼容老接口
type DeleteCompatible1Inf interface {
	// 删除一个对象
	Delete(uint) (ret interface{}, oplog string, err error)
}

type DeleteInf interface {
	// 删除一个对象
	Delete(flowId uint, uname string, c common.BaseController) (ret interface{}, oplog string, err error)
}
type MultiDeleteInf interface {
	// 删除多个对象
	MultiDelete([]string) (ret interface{}, oplog string, err error)
}

// 日志表自定义接口
type OplogModelInf interface {
	// 返回操作日志记录对象（主要是确定表名）
	OplogModel(uname, flow string, flowid uint, action, remark string) interface{}
}

// 导入接口
type Import interface {
	// 导入操作
	Import(uname string, importFile io.Reader, c common.BaseController) (ret interface{}, oplog string, err error)
}

// 导出接口
type Export interface {
	// 返回excel文件连接
	Export(uname string, query []*common.QueryConditon, fields []string,
		sortby []string, order []string) (content io.ReadSeeker, oplog string, err error)
}

type OpHistoryInf interface {
	// 获取操作状态
	GetOpHistory() (ret interface{}, oplog string, err error)
}

type OpLogHistoryInf interface {
	// 获取操作日志历表
	GetOpLogHistory() (ret interface{}, oplog string, err error)
}

type PreHandlersInf interface {
	// 获取上一步操作者
	GetPreHandlers() (ret interface{}, oplog string, err error)
}

type SetBaseControllerInf interface {
	// 获取上一步操作者
	SetBaseController(c common.BaseController)
}
type GetConfigsInf interface {
	// 获取处理当前流程所需要对参数对象（供前端使用）
	GetConfigs(uname string, id uint, c common.BaseController) (ret interface{}, oplog string, err error)
}

type GoNextInf interface {
	// 流程步骤（下一步/上一步）， remainingIds：上一步需要重新处理的，默认全部需要
	GoNext(tx *gorm.DB, remainingIds []uint) (handlers []FlowHandler, err error)
}
type GetCurStepInf interface {
	// 获取当前步骤
	GetCurStep() (step FlowStep, err error)
}
type GetNextStepInf interface {
	// 获取下一步
	GetNextStep() (step FlowStep, err error)
}
type GetPreStepInf interface {
	// 获取上一步
	GetPreStep() (step FlowStep, err error)
}
type GetValueInf interface {
	// 获取某个属性值（供step使用）
	GetValue(attr string) reflect.Value
}
type SetValueInf interface {
	// 更新属性信息（供step使用）
	SetValue(tx *gorm.DB, attr string, refvalue reflect.Value) (err error)
}

// flows 各种类型的流程集合
var services = make(map[string]FlowService)

//Register 注册新类型的服务
func Register(serviceType string, service FlowService) {
	if service == nil {
		panic("service: Register error, service is nil")
	}
	if _, ok := services[serviceType]; ok {
		panic("service: Register called twice for flow " + serviceType)
	}
	services[serviceType] = service
}

// GetService 获取Servie 对象
func GetService(serviceType string) FlowService {
	if v, ok := services[serviceType]; ok {
		return v
	}
	panic("service does not exist: " + serviceType)
}
