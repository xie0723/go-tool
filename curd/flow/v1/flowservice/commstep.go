package flowservice

import (
	"fmt"
	"strings"

	"github.com/astaxie/beego/logs"
	"gorm.io/gorm"
)

type CommStep struct {
	KeyId              string       // 标识
	Rate               int          // 通过率，默认 100
	Handler            FlowHandler  // 处理人
	HandlersInFlowAttr string       // handers 默认存储到flow到那个属性中
	Configs            []ConfigPage // 处理该步骤需要对参数对象
	LoadConfig         func(flow FlowService) []ConfigPage
	LoadJSONConfig     func(flow FlowService) interface{}                      // json生成的map嵌套对象
	CancelDo           func(tx *gorm.DB, flowId uint, flow string) (err error) // 返回当前部署时，清空数据
}

func (s *CommStep) Key() (stepKey string) {
	return s.KeyId
}

// 通过标准，100 表示所有人结论都要是通过
func (s *CommStep) PassRate() (rate int) {
	if s.Rate == 0 {
		s.Rate = 100
	}
	return s.Rate
}

// 找到注册handle
func (s *CommStep) Hander() (handler FlowHandler) {
	return s.Handler
}

// 流程步骤等操作信息
func (s *CommStep) LoadHandlers(tx *gorm.DB, flowId uint) (handlers []FlowHandler, err error) {
	if s.Handler == nil {
		return nil, fmt.Errorf("hander is nil of step:%s", s.Key())
	}
	return s.Handler.LoadStepHandlers(tx, flowId, s.Key())
}

// handlers 入库
func (s *CommStep) AddHandlers(tx *gorm.DB, handlers []FlowHandler) (err error) {
	for _, handler := range handlers {
		if err = tx.Table(s.Hander().TableName()).Create(handler).Error; err != nil {
			logs.Error("create handler failed,", err.Error())
			return
		}
	}
	return
}

// 删除当前步骤和上一步的处理记录
func (s *CommStep) ClearHandlers(tx *gorm.DB, flowId uint, flow string, steps []string, remainingIds []uint) (err error) {
	if len(remainingIds) > 0 {
		tx = tx.Where("(step = ? and service_id = ? and service = ?) or id in (?)", steps[0], flowId, flow, remainingIds)
	} else {
		tx = tx.Where("step in (?) and service_id = ? and service = ? ", steps, flowId, flow)
	}
	if err = tx.Table(s.Hander().TableName()).Delete(s.Handler).Error; err != nil {
		logs.Error("clear handlers failed,", err.Error())
		return
	}
	if s.CancelDo != nil {
		err = s.CancelDo(tx, flowId, flow)
	}
	return
}

// 从flow属性中获取责任人信息，多个逗号分割，作为步骤初始化时创建handers
func (s *CommStep) GetDefaultHandlers(flow FlowService) (handers string) {
	var ret = []string{}
	keys := strings.Split(s.HandlersInFlowAttr, ",")
	for _, key := range keys {
		if gvapp, ok := flow.(GetValueInf); ok {
			refv := gvapp.GetValue(key)
			if refv.IsValid() {
				v := refv.String()
				ret = append(ret, v)
			}
		} else {
			logs.Error("GetValueInf is not implement")
		}
	}
	handers = strings.Join(ret, ",")
	return
}

// 获取处理当前步骤所需要对参数对象（供前端使用）
func (s *CommStep) GetConfigs(flow FlowService) (ret interface{}) {
	if s.Configs != nil {
		return s.Configs
	}
	if s.LoadConfig != nil {
		return s.LoadConfig(flow)
	}
	if s.LoadJSONConfig != nil {
		return s.LoadJSONConfig(flow)
	}
	return nil
}
