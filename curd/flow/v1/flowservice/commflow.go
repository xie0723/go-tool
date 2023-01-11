package flowservice

import (
	"fmt"
	"reflect"
	"sort"
	"strings"
	"time"

	"github.com/astaxie/beego/logs"
	"github.com/daimall/tools/curd/common"
	"github.com/daimall/tools/curd/dbmysql/dbgorm"
	oplog "github.com/daimall/tools/curd/oplog"
	"gorm.io/gorm"
)

// 流程公共属性
type CommFlow struct {
	ID      uint   `gorm:"primary_key" json:"id"`                          // 自增主键
	CurStep string `gorm:"size:20;column:cur_step;index"  json:"cur_step"` // 当前步骤
	State   int    `gorm:"column:state;index"  json:"state"`               // 状态 1: 流程草稿   2: 流程流转中  3： 流程完成  4: 流程完成（被拒绝）， 5: 流程完成（超时）
	Creator string `gorm:"size:100;column:creator;index"  json:"creator"`  // 创建人

	CreatedAt time.Time `json:"created_at"` // 创建时间
	UpdatedAt time.Time `json:"updated_at"` // 最后更新时间

	BaseController common.BaseController `gorm:"-" json:"-"` //
}

// 获取ID
func (f *CommFlow) GetID() uint {
	return f.ID
}

// 获取flowname
func (s *CommFlow) GetFlowName() string {
	return ""
}

// 获取某个属性值（供step跳转使用）
func (f *CommFlow) GetValue(flow FlowService, attr string) reflect.Value {
	v := reflect.ValueOf(flow)
	attrs := strings.Split(attr, ".")
	for _, p := range attrs {
		v = GetRefValue(v)
		v = v.FieldByName(p)
	}
	return v
}

// 更新属性信息（供step使用）
func (f *CommFlow) SetValue(tx *gorm.DB, flow FlowService, attr string, refvalue reflect.Value) (err error) {
	var v = reflect.ValueOf(flow)
	attrs := strings.Split(attr, ".")
	for _, p := range attrs {
		v = GetRefValue(v)
		v = v.FieldByName(p)
	}
	v.Set(refvalue)
	if tx != nil {
		// 为nil时表示只需要更新对象，不需要更新数据库
		if err = tx.Model(flow).Updates(f).Error; err != nil {
			logs.Error("set value failed,", err.Error())
		}
	}
	return
}

// 流程步骤（下一步/上一步）
func (x *CommFlow) GoNext(tx *gorm.DB, flow FlowService, tblName, serviceName string, remainingIds []uint) (curStepHandlers []FlowHandler, err error) {
	var (
		step                FlowStep
		remainCheckHandlers []CheckHandler
	)
	if curapp, ok := flow.(GetCurStepInf); ok {
		if step, err = curapp.GetCurStep(); err != nil {
			logs.Error("get curstep failed,", err.Error())
			return
		}
	} else {
		err = fmt.Errorf("GetCurStepInf is not implement")
		return
	}

	// 加载所有审批记录
	// var handers []FlowHandler
	if curStepHandlers, err = step.LoadHandlers(tx, flow.GetID()); err != nil {
		logs.Error("LoadHandlers failed,", err.Error(), flow.GetID(), step.Key())
		return
	}
	var total = len(curStepHandlers)
	var pass, nopass int
	for _, record := range curStepHandlers {
		if record.IsFinish() && record.GetConclusion() {
			pass++
		}
		if record.IsFinish() && !record.GetConclusion() {
			nopass++
		}
	}
	// 部分责任人完成处理，只保存处理结果，不跳转
	if pass*100/total < step.PassRate() && nopass*100/total <= 100-step.PassRate() {
		logs.Debug("check one but not enough")
		return
	}
	var next FlowStep
	var preStep = false
	// 判断满足通过率
	if pass*100/total >= step.PassRate() {
		if napp, ok := flow.(GetNextStepInf); ok {
			if next, err = napp.GetNextStep(); err != nil {
				logs.Error("GetNextStep failed,", err.Error())
				return
			}
		} else {
			err = fmt.Errorf("GetNextStepInf is not implement")
			return
		}

	}
	// 已经不满足通过率
	if nopass*100/total > 100-step.PassRate() {
		// 退回流程
		if papp, ok := flow.(GetPreStepInf); ok {
			if next, err = papp.GetPreStep(); err != nil {
				logs.Error("GetPreStep failed,", err.Error())
				return
			}
			preStep = true
		} else {
			err = fmt.Errorf("GetPreStepInf is not implement")
			return
		}

	}
	if next == nil {
		// 最后一步
		if err = tx.Model(flow).Updates(map[string]interface{}{"cur_step": "-", "state": FlowStateFinish}).Error; err != nil {
			logs.Error("update supplier flow to finish failed,", err.Error())
			return
		}
		if preStep {
			// 进入草稿状态，清空当前步骤
			if err = step.ClearHandlers(tx, flow.GetID(), flow.GetFlowName(),
				[]string{step.Key()}, remainingIds); err != nil {
				logs.Error("go draft, clear handlers failed,", err.Error())
				return
			}
		}
	} else {
		// 退回，清空当前handlers
		if preStep {
			var curStep FlowStep
			if remainingIds != nil {
				if err = tx.Table(next.Hander().TableName()).Where("id in (?)", remainingIds).Find(&remainCheckHandlers).Error; err != nil {
					logs.Error("get remain check handlers failed,", err.Error())
					return
				}
			}
			if curapp, ok := flow.(GetCurStepInf); ok {
				if curStep, err = curapp.GetCurStep(); err != nil {
					logs.Error("get cur step failed,", err.Error())
					return
				}
			} else {
				err = fmt.Errorf("GetCurStepInf is not implement")
				return
			}

			if err = next.ClearHandlers(tx, flow.GetID(), flow.GetFlowName(),
				[]string{curStep.Key(), next.Key()}, remainingIds); err != nil {
				logs.Error("go prestep, clear handlers failed,", err.Error())
				return
			}
		}
		// 进入下一步
		if err = tx.Model(flow).Update("cur_step", next.Key()).Error; err != nil {
			logs.Error("update supplier flow to next step failed,", err.Error())
			return
		}
		// 添加下一步的handler
		var defaultHandlers []string
		if remainCheckHandlers == nil {
			defaultHandler := next.GetDefaultHandlers(flow)
			defaultHandlers = strings.Split(defaultHandler, ",")
		} else {
			for _, h := range remainCheckHandlers {
				defaultHandlers = append(defaultHandlers, h.User)
			}
		}
		var nextStepHandlers []FlowHandler
		for _, user := range defaultHandlers {
			nextStepHandlers = append(nextStepHandlers, &CheckHandler{
				User:        user,
				Step:        next.Key(),
				ServiceId:   flow.GetID(),
				ServiceName: serviceName,
			})
		}
		if err = next.AddHandlers(tx, nextStepHandlers); err != nil {
			logs.Error("next.AddHandlers failed,", err.Error())
			return
		}
	}
	return
}

// 获取上一步
func (f *CommFlow) GetPreStep(steps map[string]FlowStep) (step FlowStep, err error) {
	if f.CurStep == "" {
		return nil, nil //
	}
	var keys []string
	for key := range steps {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	for i := range keys {
		if f.CurStep == keys[i] {
			if i == 0 {
				return nil, nil // 前面没有步骤了（容错，一般不会触发，第一步不会调用）
			}
			return steps[keys[i-1]], nil
		}
	}
	return
}

// 获取下一步
func (f *CommFlow) GetNextStep(steps map[string]FlowStep, firstStep ...FlowStep) (step FlowStep, err error) {
	var keys []string
	for key := range steps {
		keys = append(keys, key)
	}
	if len(steps) <= 0 {
		err = fmt.Errorf("count of steps is zero")
		return
	}
	sort.Strings(keys)
	if f.CurStep == "" {
		if len(firstStep) == 1 {
			return firstStep[0], nil
		}
		return steps[keys[0]], nil
	}
	for i := range keys {
		if f.CurStep == keys[i] {
			if i+1 == len(keys) {
				// 最后一步
				return nil, nil
			}
			return steps[keys[i+1]], nil
		}
	}
	return nil, fmt.Errorf("step[%s] not found", f.CurStep)
}

// 获取操作状态（含历史）
func (f *CommFlow) OpHistory(offset, limit int, name, tblName string) (ret interface{}, err error) {
	handlers := []CheckHandler{}
	g := dbgorm.GetDBInst()
	g = g.Table(tblName).Where("service_id = ?", f.GetID()).
		Where("service = ?", name)
	if limit != 0 {
		g = g.Offset(offset).Limit(limit)
	}
	if err = g.Find(&handlers).Error; err != nil {
		logs.Error("get op history failed,", err.Error())
		return
	}
	return handlers, nil
}

// 获取下一步
func (f *CommFlow) OpLogHistory(offset, limit int, name, tblName string) (ret interface{}, err error) {
	handlers := []oplog.OpLog{}
	g := dbgorm.GetDBInst()
	g = g.Table(tblName).Where("flow_id = ?", f.GetID()).
		Where("flow = ?", name)
	if limit != 0 {
		g = g.Offset(offset).Limit(limit)
	}
	if err = g.Find(&handlers).Error; err != nil {
		logs.Error("get op log history failed,", err.Error())
		return
	}
	return handlers, nil
}

// 获取上一步的handlerid，给部分退回时使用
func (f *CommFlow) GetStepHandlers(name, tblName, stepName string) (ret interface{}, err error) {
	handlers := []CheckHandler{}
	g := dbgorm.GetDBInst()
	g = g.Table(tblName).Where("service_id = ?", f.GetID()).
		Where("service = ?", name).Where("step = ?", stepName).Where("conclusion <> ?", 0)
	if err = g.Find(&handlers).Error; err != nil {
		logs.Error("get step handlers failed,", err.Error())
		return
	}
	return handlers, nil
}

func (f *CommFlow) SetBaseController(c common.BaseController) {
	f.BaseController = c
}
