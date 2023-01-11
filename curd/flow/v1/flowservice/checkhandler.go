package flowservice

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/astaxie/beego"
	"github.com/astaxie/beego/logs"
	"github.com/daimall/tools/curd/attach"
	"github.com/daimall/tools/curd/common"
	"github.com/daimall/tools/curd/dbmysql/dbgorm"
	"gorm.io/gorm"
)

type CheckHandler struct {
	ID          uint   `gorm:"primary_key"  json:"id"`                                        // 自增主键
	User        string `gorm:"size:100;column:user;unique_index:onerecord"  json:"user"`      // 操作者的账户id
	Step        string `gorm:"size:20;column:step;unique_index:onerecord"  json:"step"`       // 步骤
	ServiceId   uint   `gorm:"column:service_id;unique_index:onerecord"  json:"serviceId"`    // FlowId
	ServiceName string `gorm:"size:50;column:service;unique_index:onerecord"  json:"service"` // FlowType
	Conclusion  int    `gorm:"column:conclusion"  json:"conclusion"`                          // 评审结论,0未评审，1 通过，2 风险通过 3 拒绝, 100 转他人处理
	Remark      string `gorm:"column:remark"  json:"remark"`                                  // 操作详情
	//Attach      string      `gorm:"size:1000;column:attach;index"  json:"attach"` // 附件
	CreatedAt     time.Time   `json:"created_at"` // 创建时间
	UpdatedAt     time.Time   `json:"updated_at"` // 最后更新时间
	TblName       string      `gorm:"-" json:"-"` // 表名
	Flow          FlowService `gorm:"-" json:"-"` // 关联的流程
	AttachKeys    []string    `gorm:"-" json:"-"` // 附件key列表
	IsPublicCheck bool        `gorm:"-" json:"-"` // 是否为公共审核，如果是不需要鉴权

	Pretreatment   func(uname string, tx *gorm.DB, c common.BaseController, h *CheckHandler) (oplog string, err error)                        `gorm:"-" json:"-"` // 预处理
	Aftertreatment func(uname string, tx *gorm.DB, c common.BaseController, h *CheckHandler, handers []FlowHandler) (oplog string, err error) `gorm:"-" json:"-"` // 后处理
	// Attachs []attach.Attach `gorm:"-" json:"attachs"` // 附件列表

	PreStepHandlerIds []uint `gorm:"-" json:"preStepHandlerIds"` // 上一步需要重新审批的用户
}

func (h *CheckHandler) Do(uname string, c common.BaseController) (ret interface{}, oplog string, err error) {
	if beego.AppConfig.DefaultBool("needlogin", true) {
		if h.IsPublicCheck == false && h.User != uname {
			err = fmt.Errorf("you[%s] have no permission", uname)
			logs.Error(err.Error())
			return
		}
	}
	dbInst := dbgorm.GetDBInst()
	tx := dbInst.Begin()
	defer func() {
		if err == nil {
			tx.Commit()
			return
		}
		tx.Rollback()
	}()
	var checkHandler = CheckHandler{}
	if h.Conclusion == ConclusionGo {
		// 表示不需要界面传回审核结论，非审核步骤
		checkHandler.Remark = "submit ok"
		checkHandler.Conclusion = h.Conclusion
	} else {
		var jsonBody []byte
		params := c.GetString("params")
		if len(params) > 0 {
			logs.Debug("no params, maybe amis scene")
			jsonBody = []byte(params)
		} else {
			// 兼容amis 界面传参的场景
			jsonBody = c.Ctx.Input.RequestBody
		}
		if err = json.Unmarshal(jsonBody, &checkHandler); err != nil {
			logs.Error("Unmarshal check handler data error,", err.Error())
			err = common.ParamsErr
			return
		}
	}

	if checkHandler.Conclusion == ConCLusionTransferOther {
		// 转给他人处理
		h.User = checkHandler.User
		if err = tx.Model(h).Table(h.TableName()).Updates(h).Error; err != nil {
			logs.Error("update step user error %s", err.Error())
			return
		}
		ret = fmt.Sprintf("%s transfer to %s", h.Step, checkHandler.User)
		return
	}
	var attachs []attach.Attach
	if attachs, err = attach.UploadAttach(c, h.Flow.GetFlowName(), h.Flow.GetID(),
		h.AttachKeys, tx); err != nil {
		err = common.UploadErr
		return
	}
	defer func() {
		if err != nil {
			if verr := attach.DeleteAttachs(attachs); verr != nil {
				logs.Error(verr.Error())
			}
		}
	}()
	h.Conclusion = checkHandler.Conclusion
	h.Remark = checkHandler.Remark
	if h.Pretreatment != nil && (checkHandler.Conclusion == ConclusionGo || checkHandler.Conclusion == ConclusionGoWithRisk) {
		if oplog, err = h.Pretreatment(uname, tx, c, h); err != nil {
			logs.Error("Pretreatment exec failed,", err.Error())
			return
		}
	}
	if err = tx.Table(h.TableName()).Where("id = ?", h.ID).Updates(h).Error; err != nil {
		logs.Error("update step conclusion error %s", err.Error())
		return
	}
	var curStepHandlers []FlowHandler
	if napp, ok := h.Flow.(GoNextInf); ok {
		if curStepHandlers, err = napp.GoNext(tx, checkHandler.PreStepHandlerIds); err != nil {
			return
		}
	} else {
		err = fmt.Errorf("GoNextInf is not implement")
		return
	}

	if h.Aftertreatment != nil {
		if oplog, err = h.Aftertreatment(uname, tx, c, h, curStepHandlers); err != nil {
			logs.Error("Aftertreatment exec failed", err.Error())
			return
		}
	}
	if oplog == "" {
		oplog = fmt.Sprintf("%d|%s", h.Conclusion, h.Remark)
	}
	return
}

// 从数据库中加载数据，返回一个新实例
func (h *CheckHandler) LoadInst(flow FlowService, uname string, id uint) (ret FlowHandler, err error) {
	handler := &CheckHandler{ID: id}
	if id == 0 {
		// 查询默认handle
		handler.ServiceId = flow.GetID()
		handler.ServiceName = flow.GetFlowName()
		var step FlowStep
		if curApp, ok := flow.(GetCurStepInf); ok {
			if step, err = curApp.GetCurStep(); err != nil {
				logs.Error("get cur step failed,", err.Error(), flow)
				return
			}
		} else {
			err = fmt.Errorf("GetCurStepInf is not implement")
			return
		}
		handler.Step = step.Key()
	}
	dbInst := dbgorm.GetDBInst()
	if err = dbInst.Table(h.TableName()).Where(handler).First(handler).Error; err != nil {
		logs.Error("get handler failed,", err.Error(), handler)
		return
	}
	handler.TblName = h.TblName
	handler.Flow = flow
	handler.Pretreatment = h.Pretreatment
	handler.AttachKeys = h.AttachKeys
	handler.Aftertreatment = h.Aftertreatment
	handler.Conclusion = h.Conclusion
	handler.PreStepHandlerIds = h.PreStepHandlerIds
	handler.IsPublicCheck = h.IsPublicCheck

	return handler, nil
}

// 是否通过
func (h *CheckHandler) GetConclusion() bool {
	// 通过或者风险通过视为通过
	return h.Conclusion == ConclusionGo || h.Conclusion == ConclusionGoWithRisk
}

// 是否完成审核
func (h *CheckHandler) IsFinish() bool {
	return h.Conclusion != 0
}

// 表名
func (h *CheckHandler) TableName() string {
	if h.TblName != "" {
		return h.TblName
	}
	return "step_handlers"
}

// 加载handler
func (h *CheckHandler) LoadStepHandlers(tx *gorm.DB, flowId uint, stepKey string) (handlers []FlowHandler, err error) {
	l := []CheckHandler{}
	err = tx.Table(h.TableName()).
		Where(&CheckHandler{ServiceId: flowId, TblName: h.TableName(), Step: stepKey}).
		Find(&l).Error
	if len(l) == 0 {
		return nil, fmt.Errorf("no hander found for step[%s],flowId:[%d]", stepKey, flowId)
	}
	handlers = make([]FlowHandler, len(l))
	for i := range l {
		handlers[i] = &l[i]
	}
	return
}

// 非接口方法
// 设置附件keys，用于保存附件使用
func (h *CheckHandler) SetAttachKeys(keys []string) {
	h.AttachKeys = keys
}
