package flowcontroller

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"runtime"
	"strings"
	"time"

	"github.com/astaxie/beego"
	"github.com/astaxie/beego/logs"
	"github.com/daimall/tools/curd/common"
	"github.com/daimall/tools/curd/customerror"
	"github.com/daimall/tools/curd/flow/v1/flowservice"
)

// FlowController operations for model
type FlowController struct {
	BaseController
}

// URLMapping ...
func (c *FlowController) URLMapping() {
	c.Mapping("Post", c.Post)
	c.Mapping("Action", c.Action)
	c.Mapping("GetOne", c.GetOne)
	c.Mapping("GetAll", c.GetAll)
	c.Mapping("Put", c.Put)
	c.Mapping("Delete", c.Delete)
	c.Mapping("DeleteList", c.DeleteList)
}

// Post ...
// @Title 创建一条流程
// @Description create Service
// @Param	service	string 	Service	true		"body for Service content"
// @Success 201 {int} Service
// @Failure 403 body is empty
// @router /:service [post]
func (c *FlowController) Post() {
	var err error
	var ret interface{}
	var serviceId uint
	var oplog string
	defer func() {
		c.ResponseJSON(err, ret, serviceId, ServiceActionCreate, oplog)
	}()
	serviceId, ret, oplog, err = c.Service.New(c.uname, c.BaseController.BaseController)
}

// Action ...
// @Title 独立动作
// @Description handle a action
// @Param	body		body 	params	true		"body for Service content"
// @Success 201 {int} OK(step info)
// @Failure 403 body is empty
// @router /:service/:id/:action [post]
func (c *FlowController) Action() {
	var err error
	var ret interface{}
	var serviceId uint
	var oplog string
	var action flowservice.Action
	var actionType = c.Ctx.Input.Param(":action")
	defer func() {
		c.ResponseJSON(err, ret, serviceId, actionType, oplog)
	}()

	idStr := c.Ctx.Input.Param(":id")
	if serviceId, err = getUintID(idStr); err != nil {
		logs.Error("get serviceId failed,", err.Error())
		return
	}

	if actionType == "" {
		err = common.ActionNotFound
		return
	}
	if serviceId != 0 {
		if c.Service, err = c.Service.LoadInst(serviceId); err != nil {
			logs.Error("service.LoadInst failed,", err.Error())
			return
		}
	}
	if actionApp, ok := c.Service.(flowservice.ActionInf); ok {
		if action, err = actionApp.GetAction(serviceId, actionType); err != nil {
			logs.Error("GetAction[%s] failed, %s", actionType, err.Error())
			return
		}
	} else {
		err = fmt.Errorf("GetAction method is not implement")
		return
	}
	ret, serviceId, oplog, err = action.Do(c.uname, serviceId, actionType, c.BaseController.BaseController)
}

// GetOne ...
// @Title Get One
// @Description get Service by id
// @Param	id		path 	string	true		"The key for staticblock"
// @Success 200 {object} Service
// @Failure 403 :id is empty
// @router /:service/:id [get]
func (c *FlowController) GetOne() {
	var err error
	var oplog string
	var serviceId uint
	var ret interface{}
	defer func() {
		c.ResponseJSON(err, ret, serviceId, ServiceActionGetOne, oplog)
	}()
	idStr := c.Ctx.Input.Param(":id")
	if serviceId, err = getUintID(idStr); err != nil {
		logs.Error("get serviceId failed,", err.Error())
		return
	}
	if getOneApp, ok := c.Service.(flowservice.GetOneInf); ok {
		ret, oplog, err = getOneApp.GetOne(c.uname, serviceId)
		return
	}
	err = fmt.Errorf("GetOne interface not implement")
}

// GetAll ...
// @Title Get All
// @Description get Service
// @Param	query	query	string	false	"Filter. e.g. col1:v1,col2:v2 ..."
// @Param	fields	query	string	false	"Fields returned. e.g. col1,col2 ..."
// @Param	sortby	query	string	false	"Sorted-by fields. e.g. col1,col2 ..."
// @Param	order	query	string	false	"Order corresponding to each sortby field, if single value, apply to all sortby fields. e.g. desc,asc ..."
// @Param	limit	query	string	false	"Limit the size of result set. Must be an integer"
// @Param	offset	query	string	false	"Start position of result set. Must be an integer"
// @Success 200 {object} Service
// @Failure 403
// @router /:service [get]
func (c *FlowController) GetAll() {
	var err error
	var oplog string
	var serviceId uint
	var ret struct {
		Items interface{} `json:"items"`
		Total int64       `json:"total"`
	}
	var l interface{}
	var count int64
	defer func() {
		ret.Items = l
		ret.Total = count
		c.ResponseJSON(err, ret, serviceId, ServiceActionGetAll, oplog)
	}()
	//amis orderBy=id&orderDir=desc
	// orderBy = sortBy
	// orderDir = order
	// limit = perPage
	// offset = (page-1) * perPage
	var fields []string
	var sortby []string
	var order []string
	var query []*common.QueryConditon
	var limit int = 10
	var offset int
	// fields: col1,col2,entity.col3
	if v := c.GetString("fields"); v != "" {
		fields = strings.Split(v, ",")
	}
	// limit: 10 (default is 10)
	if v, err := c.GetInt("limit"); err == nil {
		limit = v
	}
	// offset: 0 (default is 0)
	if v, err := c.GetInt("offset"); err == nil {
		offset = v
	}
	// 适配amis
	if v, err := c.GetInt("perPage"); err == nil {
		limit = v
		if v, err := c.GetInt("page"); err == nil {
			offset = (v - 1) * limit
		}
	}

	// sortby: col1,col2
	if v := c.GetString("sortby"); v != "" {
		v = strings.Replace(v, ".", "__", -1)
		sortby = strings.Split(v, ",")
	}

	// 适配amis
	if v := c.GetString("orderBy"); v != "" {
		sortby = []string{v}
	}

	// order: desc,asc
	if v := c.GetString("order"); v != "" {
		order = strings.Split(v, ",")
	}

	// 适配amis
	if v := c.GetString("orderDir"); v != "" {
		order = []string{v}
	}

	var keepMap = map[string]struct{}{
		"orderDir": {},
		"orderBy":  {},
		"page":     {},
		"perPage":  {},
	}
	if beego.AppConfig.DefaultString("webKind", "BS") == "AMIS" {
		// query: k|type=v,v,v  k|type:v|v|v  其中Type可以没有,默认值是 MultiText
		kv := c.Ctx.Request.URL.Query()
		for kInit, v1 := range kv {
			if _, ok := keepMap[kInit]; ok {
				continue
			}
			vInit := v1[0]
			if len(strings.TrimSpace(vInit)) == 0 {
				continue
			}
			qcondtion := new(common.QueryConditon)
			key_type := strings.Split(kInit, "|") // 解析key中的type信息
			if len(key_type) == 2 {
				qcondtion.QueryKey = key_type[0]
				qcondtion.QueryType = key_type[1]
			} else if len(key_type) == 1 {
				qcondtion.QueryKey = key_type[0]
				qcondtion.QueryType = common.MultiText
			} else {
				logs.Error("Error: invalid query key|type format," + kInit)
				c.JSONResponse(common.QueryCondErr)
				return
			}
			qcondtion.QueryValues = strings.Split(vInit, ",") // 解析出values信息
			//qcondtion.QueryKey = strings.Replace(qcondtion.QueryKey, ".", "__", -1)
			query = append(query, qcondtion)
		}
	} else {
		// query: k|type:v|v|v,k|type:v|v|v  其中Type可以没有,默认值是 MultiText
		if v := c.GetString("query"); v != "" {
			for _, cond := range strings.Split(v, ",") { // 分割多个查询key
				qcondtion := new(common.QueryConditon)
				kv := strings.SplitN(cond, ":", 2)
				if len(kv) != 2 {
					logs.Error("query condtion format error:%s, need key:value", kv)
					c.JSONResponse(common.QueryCondErr)
					return
				}
				kInit, vInit := kv[0], kv[1]          // 初始分割查询key和value（备注，value是多个用|分割）
				key_type := strings.Split(kInit, "|") // 解析key中的type信息
				if len(key_type) == 2 {
					qcondtion.QueryKey = key_type[0]
					qcondtion.QueryType = key_type[1]
				} else if len(key_type) == 1 {
					qcondtion.QueryKey = key_type[0]
					qcondtion.QueryType = common.MultiText
				} else {
					logs.Error("Error: invalid query key|type format," + kInit)
					c.JSONResponse(common.QueryCondErr)
					return
				}
				qcondtion.QueryValues = strings.Split(vInit, "|") // 解析出values信息
				//qcondtion.QueryKey = strings.Replace(qcondtion.QueryKey, ".", "__", -1)
				query = append(query, qcondtion)
			}
		}
	}

	if getAllApp, ok := c.Service.(flowservice.GetAllInf); ok {
		l, count, oplog, err = getAllApp.GetAll(c.uname, query, fields, sortby, order, offset, limit)
		return
	}
	err = fmt.Errorf("getall interface is not implement")
}

// Put ...
// @Title Put
// @Description update the Service
// @Param	id		path 	string	true		"The id you want to update"
// @Param	body		body 	Service	true		"body for Service content"
// @Success 200 {object} Service
// @Failure 403 :id is not int
// @router /:service/:id [put]
func (c *FlowController) Put() {
	var err error
	var ret interface{}
	var oplog string
	var serviceId uint
	defer func() {
		c.ResponseJSON(err, ret, serviceId, ServiceActionPut, oplog)
	}()

	// 指定更新字段
	var fields []string
	// fields: col1,col2,entity.col3
	if v := c.GetString("fields"); v != "" {
		fields = strings.Split(v, ",")
	}
	idStr := c.Ctx.Input.Param(":id")
	if serviceId, err = getUintID(idStr); err != nil {
		logs.Error("get serviceId failed,", err.Error())
		return
	}
	if serviceId != 0 {
		if c.Service, err = c.Service.LoadInst(serviceId); err != nil {
			logs.Error("service.LoadInst failed,", err.Error())
			return
		}
	}
	if updateApp, ok := c.Service.(flowservice.UpdateInf); ok {
		ret, oplog, err = updateApp.Update(serviceId, fields, c.BaseController.BaseController)
		return
	}
	err = fmt.Errorf("update interface is not implement")
}

// Delete ...
// @Title Delete
// @Description delete the Service
// @Param	id		path 	string	true		"The id you want to delete"
// @Success 200 {string} delete success!
// @Failure 403 id is empty
// @router /:service/:id [delete]
func (c *FlowController) Delete() {
	var err error
	var ret interface{}
	var oplog string
	var serviceId uint
	defer func() {
		c.ResponseJSON(err, ret, serviceId, ServiceActionDelete, oplog)
	}()
	idStr := c.Ctx.Input.Param(":id")
	if serviceId, err = getUintID(idStr); err != nil {
		logs.Error("get serviceId failed,", err.Error())
		return
	}
	if deleteApp, ok := c.Service.(flowservice.DeleteInf); ok {
		ret, oplog, err = deleteApp.Delete(serviceId, c.uname, c.BaseController.BaseController)
		return
	}
	// 兼容老delete接口
	if deleteApp, ok := c.Service.(flowservice.DeleteCompatible1Inf); ok {
		ret, oplog, err = deleteApp.Delete(serviceId)
		return
	}
	err = fmt.Errorf("delete interface is not implement")
}

// DeleteList ...
// @Title multi-Delete
// @Description delete multi Services
// @Param	ids	 	string	true		"The ids you want to delete"
// @Success 200 {string} delete success!
// @Failure 403 id is empty
// @router /:service/deletelist [delete]
func (c *FlowController) DeleteList() {
	var err error
	var ret interface{}
	var oplog string
	var serviceId uint
	defer func() {
		c.ResponseJSON(err, ret, serviceId, ServiceActionDeleteList, oplog)
	}()
	if dlApp, ok := c.Service.(flowservice.MultiDeleteInf); ok {
		ret, oplog, err = dlApp.MultiDelete(strings.Split(c.GetString("ids"), ","))
		return
	}
	err = fmt.Errorf("DeleteList interface is not implement")
}

// Import ...
// @Title 导出excel，批量创建对象
// @Description batch create Service
// @Param	service	string 	Service	true		"body for Service content"
// @Success 201 {int} Service
// @Failure 403 body is empty
// @router /:service/import [post]
func (c *FlowController) Import() {
	var err error
	var ret interface{}
	var oplog string
	defer func() {
		c.ResponseJSON(err, ret, 0, ServiceActionDeleteList, oplog)
	}()
	if importApp, ok := c.Service.(flowservice.Import); ok {
		var importFile io.Reader
		if importFile, _, err = c.GetFile("importFile"); err != nil {
			logs.Error("get GetFile:importFile failed", err.Error())
			return
		}
		ret, oplog, err = importApp.Import(c.uname, importFile, c.BaseController.BaseController)
		return
	}
	err = fmt.Errorf("import interface not implement")
}

// Export ...
// @Title export
// @Description get Service
// @Param	query	query	string	false	"Filter. e.g. col1:v1,col2:v2 ..."
// @Param	fields	query	string	false	"Fields returned. e.g. col1,col2 ..."
// @Param	sortby	query	string	false	"Sorted-by fields. e.g. col1,col2 ..."
// @Param	order	query	string	false	"Order corresponding to each sortby field, if single value, apply to all sortby fields. e.g. desc,asc ..."
// @Param	limit	query	string	false	"Limit the size of result set. Must be an integer"
// @Param	offset	query	string	false	"Start position of result set. Must be an integer"
// @Success 200 {object} Service
// @Failure 403
// @router /:service/export [get]
func (c *FlowController) Export() {
	var err error
	var oplog string

	var fields []string
	var sortby []string
	var order []string
	var query []*common.QueryConditon

	defer func() {
		var method string
		pc, _, _, _ := runtime.Caller(1)
		method = runtime.FuncForPC(pc).Name()
		if err == nil {
			// 记录操作日志
			c.LogFunc(0, "export", oplog)
		} else {
			if customErr, ok := err.(customerror.CustomError); ok {
				logs.Error("FlowController[%s]%s(customErr)", method, err.Error())
				c.JSONResponse(customErr)
			} else {
				logs.Error("FlowController[%s]%s", method, err.Error())
				c.JSONResponse(customerror.New(-1, err.Error()))
			}
		}
	}()
	// fields: col1,col2,entity.col3
	if v := c.GetString("fields"); v != "" {
		fields = strings.Split(v, ",")
	}
	// sortby: col1,col2
	if v := c.GetString("sortby"); v != "" {
		v = strings.Replace(v, ".", "__", -1)
		sortby = strings.Split(v, ",")
	}

	// 适配amis
	if v := c.GetString("orderBy"); v != "" {
		sortby = []string{v}
	}

	// order: desc,asc
	if v := c.GetString("order"); v != "" {
		order = strings.Split(v, ",")
	}

	// 适配amis
	if v := c.GetString("orderDir"); v != "" {
		order = []string{v}
	}

	var keepMap = map[string]struct{}{
		"orderDir": {},
		"orderBy":  {},
		"page":     {},
		"perPage":  {},
	}
	if beego.AppConfig.DefaultString("webKind", "BS") == "AMIS" {
		// query: k|type=v,v,v  k|type:v|v|v  其中Type可以没有,默认值是 MultiText
		kv := c.Ctx.Request.URL.Query()
		for kInit, v1 := range kv {
			if _, ok := keepMap[kInit]; ok {
				continue
			}
			vInit := v1[0]
			qcondtion := new(common.QueryConditon)
			key_type := strings.Split(kInit, "|") // 解析key中的type信息
			if len(key_type) == 2 {
				qcondtion.QueryKey = key_type[0]
				qcondtion.QueryType = key_type[1]
			} else if len(key_type) == 1 {
				qcondtion.QueryKey = key_type[0]
				qcondtion.QueryType = common.MultiText
			} else {
				logs.Error("Error: invalid query key|type format," + kInit)
				c.JSONResponse(common.QueryCondErr)
				return
			}
			qcondtion.QueryValues = strings.Split(vInit, ",") // 解析出values信息
			//qcondtion.QueryKey = strings.Replace(qcondtion.QueryKey, ".", "__", -1)
			query = append(query, qcondtion)
		}
	} else {
		// query: k|type:v|v|v,k|type:v|v|v  其中Type可以没有,默认值是 MultiText
		if v := c.GetString("query"); v != "" {
			for _, cond := range strings.Split(v, ",") { // 分割多个查询key
				qcondtion := new(common.QueryConditon)
				kv := strings.SplitN(cond, ":", 2)
				if len(kv) != 2 {
					logs.Error("query condtion format error:%s, need key:value", kv)
					c.JSONResponse(common.QueryCondErr)
					return
				}
				kInit, vInit := kv[0], kv[1]          // 初始分割查询key和value（备注，value是多个用|分割）
				key_type := strings.Split(kInit, "|") // 解析key中的type信息
				if len(key_type) == 2 {
					qcondtion.QueryKey = key_type[0]
					qcondtion.QueryType = key_type[1]
				} else if len(key_type) == 1 {
					qcondtion.QueryKey = key_type[0]
					qcondtion.QueryType = common.MultiText
				} else {
					logs.Error("Error: invalid query key|type format," + kInit)
					c.JSONResponse(common.QueryCondErr)
					return
				}
				qcondtion.QueryValues = strings.Split(vInit, "|") // 解析出values信息
				//qcondtion.QueryKey = strings.Replace(qcondtion.QueryKey, ".", "__", -1)
				query = append(query, qcondtion)
			}
		}
	}

	if exportApp, ok := c.Service.(flowservice.Export); ok {
		var content io.ReadSeeker
		if content, oplog, err = exportApp.Export(c.uname, query, fields, sortby, order); err != nil {
			logs.Error("export failed,", err.Error())
			return
		}
		c.Ctx.ResponseWriter.Header().Add("Content-Disposition", "attachment")
		c.Ctx.ResponseWriter.Header().Add("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
		http.ServeContent(c.Ctx.ResponseWriter, c.Ctx.Request, "export", time.Now(), content)
		return
	}
	err = fmt.Errorf("export interface not implement")
}

// Configs ...
// @Title Get 获取参数对象结构
// @Description get configs for next
// @Param id path string true "The key for staticblock"
// @Success 200 {object} Service
// @Failure 403 :id is empty
// @router /:service/:id/config [get]
func (c *FlowController) Configs() {
	var err error
	var ret interface{}
	var serviceId uint
	var oplog string
	defer func() {
		c.ResponseJSON(err, ret, serviceId, ServiceActionGetConfigs, oplog)
	}()
	idStr := c.Ctx.Input.Param(":id")
	if serviceId, err = getUintID(idStr); err != nil {
		logs.Error("get serviceId failed,", err.Error())
		return
	}
	if configApp, ok := c.Service.(flowservice.GetConfigsInf); ok {
		ret, oplog, err = configApp.GetConfigs(c.uname, serviceId, c.BaseController.BaseController)
		return
	}
	err = fmt.Errorf("GetConfigs interface is not implement")
}

// Next ...
// @Title 进入流程下一步
// @Description handle Service to next step
// @Param	body		body 	params	true		"body for Service content"
// @Success 201 {int} OK(step info)
// @Failure 403 body is empty
// @router /:service/:id/:handlerId [post]
func (c *FlowController) Next() {
	var err error
	var ret interface{}
	var serviceId uint
	var oplog string
	var handler flowservice.FlowHandler
	var handlerId uint
	var action string
	defer func() {
		c.ResponseJSON(err, ret, serviceId, action, oplog)
	}()

	idStr := c.Ctx.Input.Param(":id")
	if serviceId, err = getUintID(idStr); err != nil {
		logs.Error("get serviceId failed,", err.Error())
		return
	}
	idStr = c.Ctx.Input.Param(":handlerId")
	if handlerId, err = getUintID(idStr); err != nil {
		logs.Error("get handlerId failed,", err.Error())
		return
	}
	if serviceId != 0 {
		if c.Service, err = c.Service.LoadInst(serviceId); err != nil {
			logs.Error("service.LoadInst failed,", err.Error())
			return
		}
	}
	var step flowservice.FlowStep
	if curStepApp, ok := c.Service.(flowservice.GetCurStepInf); ok {
		if step, err = curStepApp.GetCurStep(); err != nil {
			logs.Error("GetCurStep failed, %s", err.Error())
			return
		}
	} else {
		err = fmt.Errorf("GetCurStep interface is not implement")
		return
	}
	action = step.Key()
	if handler, err = step.Hander().LoadInst(c.Service, c.uname, handlerId); err != nil {
		logs.Error("LoadInst failed, %s", err.Error())
		return
	}
	ret, oplog, err = handler.Do(c.uname, c.BaseController.BaseController)
}

// GetHistory ...
// @Title Get 获取流程的操作记录（历史）
// @Description get get operation list history
// @Param id path string true "The key for staticblock"
// @Success 200 {object} Service
// @Failure 403 :id is empty
// @router /:service/:id/oplist [get]
func (c *FlowController) GetHistory() {
	var err error
	var ret interface{}
	var serviceId uint
	var oplog string
	defer func() {
		c.ResponseJSON(err, ret, serviceId, ServiceOpList, oplog)
	}()
	idStr := c.Ctx.Input.Param(":id")
	if serviceId, err = getUintID(idStr); err != nil {
		logs.Error("get serviceId failed,", err.Error())
		return
	}
	if c.Service, err = c.Service.LoadInst(serviceId); err != nil {
		logs.Error("service.LoadInst failed,", err.Error())
		return
	}
	if historyApp, ok := c.Service.(flowservice.OpHistoryInf); ok {
		ret, oplog, err = historyApp.GetOpHistory()
	} else {
		err = errors.New("GetOpHistory interface not impl")
	}
}

// GetOpLogHistory ...
// @Title Get 获取流程等操作日志记录
// @Description get get operation list history
// @Param id path string true "The key for staticblock"
// @Success 200 {object} Service
// @Failure 403 :id is empty
// @router /:service/:id/oploglist [get]
func (c *FlowController) GetOpLogHistory() {
	var err error
	var ret interface{}
	var serviceId uint
	var oplog string
	defer func() {
		c.ResponseJSON(err, ret, serviceId, ServiceOpList, oplog)
	}()
	idStr := c.Ctx.Input.Param(":id")
	if serviceId, err = getUintID(idStr); err != nil {
		logs.Error("get serviceId failed,", err.Error())
		return
	}
	if c.Service, err = c.Service.LoadInst(serviceId); err != nil {
		logs.Error("service.LoadInst failed,", err.Error())
		return
	}
	if historyApp, ok := c.Service.(flowservice.OpLogHistoryInf); ok {
		ret, oplog, err = historyApp.GetOpLogHistory()
	} else {
		err = errors.New("GetOpLogHistory interface not impl")
	}
}

// GetPreHandlers ...
// @Title Get 获取上一步处理人（退回流程使用）
// @Description get get operation list history
// @Param id path string true "The key for staticblock"
// @Success 200 {object} Service
// @Failure 403 :id is empty
// @router /:service/:id/prehandlers [get]
func (c *FlowController) GetPreHandlers() {
	var err error
	var ret interface{}
	var serviceId uint
	var oplog string
	defer func() {
		c.ResponseJSON(err, ret, serviceId, ServiceOpList, oplog)
	}()
	idStr := c.Ctx.Input.Param(":id")
	if serviceId, err = getUintID(idStr); err != nil {
		logs.Error("get serviceId failed,", err.Error())
		return
	}
	if c.Service, err = c.Service.LoadInst(serviceId); err != nil {
		logs.Error("service.LoadInst failed,", err.Error())
		return
	}
	if preHandlersApp, ok := c.Service.(flowservice.PreHandlersInf); ok {
		ret, oplog, err = preHandlersApp.GetPreHandlers()
	} else {
		err = errors.New("GetOpHistory interface not impl")
	}
}
