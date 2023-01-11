package jira

import (
	"strconv"

	"github.com/astaxie/beego/logs"
	"github.com/daimall/tools/curd/common"
)

// JiraBugController operations for JiraBug
type JiraBugController struct {
	common.BaseController
}

// 从jira上获取项目列表
// @router /v1/jira/project [get]
func (c *JiraBugController) GetProjects() {
	if ret, err := getProjectList(); err != nil {
		c.Data["json"] = common.RestResult{Code: -1, Message: err.Error()}
	} else {
		c.Data["json"] = common.RestResult{Code: 0, Data: ret}
	}
	c.ServeJSON()
}

// 从jira上获取项目信息
// @router /v1/jira/project/:id [get]
func (c *JiraBugController) GetProject() {
	idStr := c.Ctx.Input.Param(":id")
	if ret, err := getProject(idStr); err != nil {
		c.Data["json"] = common.RestResult{Code: -1, Message: err.Error()}
	} else {
		c.Data["json"] = common.RestResult{Code: 0, Data: ret}
	}
	c.ServeJSON()
}

// 获得field可选属性
// @router /v1/jira/project/field/:id/option [get]
func (c *JiraBugController) GetProjectAttrs() {
	var err error
	var id float64
	var ret interface{}
	defer func() {
		if err != nil {
			c.Data["json"] = common.RestResult{Code: -1, Message: err.Error()}
		} else {
			c.Data["json"] = common.RestResult{Code: 0, Data: ret}
		}
		c.ServeJSON()
	}()
	idStr := c.Ctx.Input.Param(":id")
	if id, err = strconv.ParseFloat(idStr, 64); err != nil {
		logs.Error(err.Error())
		return
	}
	ret, err = getFieldOptions(id)
}

// 判断JiraKey是否存在
// @router /v1/jira/keyexist [get]
func (c *JiraBugController) JiraKeyIsExist() {
	jiraKey := c.GetString("jirakey")
	if ret, err := jiraKeyIsExist(jiraKey); err != nil {
		c.Data["json"] = common.RestResult{Code: -1, Message: err.Error()}
	} else {
		c.Data["json"] = common.RestResult{Code: 0, Data: ret}
	}
	c.ServeJSON()
}
