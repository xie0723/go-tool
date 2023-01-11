package jira

import (
	"fmt"
	"time"

	"github.com/astaxie/beego"
	"github.com/astaxie/beego/httplib"
	"github.com/astaxie/beego/logs"
	"github.com/daimall/tools/aes/cbc"
	_ "github.com/go-sql-driver/mysql"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var db *gorm.DB

type JiraData struct {
	SyncKey         string                   `json:"customfield_20400,omitempty"`  // 对方的KEY
	Summary         string                   `json:"summary,omitempty"`            // 概要
	Level           map[string]interface{}   `json:"customfield_15121,omitempty"`  // 严重程度
	Versions        []map[string]interface{} `json:"versions,omitempty"`           // 影响版本
	Project         map[string]interface{}   `json:"project,omitempty"`            // 项目
	Components      []map[string]interface{} `json:"components,omitempty"`         // 模块
	RecurRate       map[string]interface{}   `json:"customfield_17400,omitempty"`  // 重现概率
	Description     string                   `json:"description,omitempty"`        // 描述
	Occurdate       string                   `json:"customfield_17200,omitempty"`  // 发生时间
	IssueType       map[string]interface{}   `json:"issuetype,omitempty"`          // 问题类型
	Areas           map[string]interface{}   `json:"customfield_17302,omitempty"`  // 领域
	VersionType     map[string]interface{}   `json:"customfield_17101,omitempty"`  // 版本类型
	Priority        map[string]interface{}   `json:"priority,omitempty"`           // 优先级
	Milestone       map[string]interface{}   `json:"customfield_16800,omitempty"'` // 里程碑版本
	CurrentOwner    map[string]interface{}   `json:"assignee,omitempty"`           // 处理人
	ExpectStart     string                   `json:"customfield_16729,omitempty"`  // 预计开始
	ExpectEnd       string                   `json:"customfield_16730,omitempty"`  // 预计结束
	Source          *CommonParam             `json:"customfield_16100,omitempty"`  // 来源
	Product         []CommonParam            `json:"customfield_16107,omitempty"`  // 产品
	DemandAttr      *CommonParam             `json:"customfield_19400,omitempty"`  // 需求属性
	TargetMarket    []CommonParam            `json:"customfield_19700,omitempty"`  // 目标市场
	OtherFields     []CommonParam            `json:"customfield_16104,omitempty"`  // 是否涉及其它领域
	DesignDocuments string                   `json:"customfield_16617,omitempty"`  // 设否需要设计文档
	Risk            string                   `json:"customfield_17300,omitempty"`  // 风险
	Duedate         string                   `json:"duedate,omitempty"`            // 到期日
	Comment         *struct {
		Comments []Comments `json:"comments"`
	} `json:"comment,omitempty"`
	Attachment []Attachment `json:"attachment,omitempty"` // 附件信息
	//Label           []string                 `json:"label"`              // 标签
}

type Comments struct {
	Id     string `json:"id"`
	Author Author `json:"author"`
	Body   string `json:"body"`
}
type Author struct {
	Name string `json:"name"`
}
type CommonParam struct {
	ID string `json:"id""`
}

type FieldOption struct {
	Id          string  `gorm:"column:ID"`
	CUSTOMFIELD float64 `gorm:"column:CUSTOMFIELD"`
	Value       string  `gorm:"column:customvalue"`
}

type Attachment struct {
	Id       string `json:"id"`
	Self     string `json:"self"`
	FileName string `json:"fileName"`
	Size     int64  `json:"size"`
	Content  string `json:"content"`
}

// 添加jira问题单
func AddJiraIssue(jiraBug *JiraData, auth string) (ret interface{}, err error) {
	if jiraBug.Summary == "" {
		err = fmt.Errorf("summary is null")
		logs.Error(err.Error())
		return
	}
	if auth == "" {
		err = fmt.Errorf("auth is null")
		logs.Error(err.Error())
		return
	}
	// 通过rest接口，通知slave存储任务信息
	url := beego.AppConfig.String("JIRA::ADDRESS") + "/rest/api/2/issue/"
	_jiraBug := struct {
		Fields *JiraData `json:"fields"`
	}{Fields: jiraBug}
	//_jiraBug.Fields.Occurdate = "2019-11-04T15:08:00.400+0800"
	var req *httplib.BeegoHTTPRequest
	req, err = httplib.Post(url).Header("Authorization", "Basic "+auth).
		Header("Content-Type", "application/json").JSONBody(_jiraBug)
	if err != nil {
		err = fmt.Errorf("JIRA系统异常, JSONBody:add issue to jira failed, url:%s, err:%s", url, err.Error())
		logs.Error(err.Error())
		return
	}
	// 执行post接口，获取rest返回值
	var resp interface{}
	if err = req.ToJSON(&resp); err != nil {
		err = fmt.Errorf("JIRA系统异常, ToJSON:add issue to jira failed, url:%s, err:%s", url, err.Error())
		logs.Error(err.Error())
		return
	}
	logs.Debug(fmt.Sprintf("ADD JIRA ISSUSE RSULT:%+v", resp))
	// 判断返回值，入库
	if _r, ok := resp.(map[string]interface{}); ok {
		if _r["key"] == nil {
			err = fmt.Errorf("jira result does not have key info:%+v", resp)
			logs.Error(err.Error())
			return
		}
		return _r["key"], nil
	} else {
		err = fmt.Errorf("parse key from result failed, result:%s", resp)
		logs.Error(err.Error())
		return
	}
}

// AddJiraRemark 添加jira问题备注
func AddJiraRemark(remarks []string, key string, auth string) (err error) {
	var result interface{}
	if auth == "" {
		return fmt.Errorf("auth is null")
	}

	//通过rest接口，通知slave存储任务信息
	url := fmt.Sprintf("%s/rest/api/2/issue/%s/comment", beego.AppConfig.String("JIRA::ADDRESS"), key)

	for i := range remarks {
		data := struct {
			Body string `json:"body"`
		}{Body: remarks[i]}

		req, err := httplib.Post(url).Header("Authorization", "Basic "+auth).
			Header("Content-Type", "application/json").JSONBody(data)
		if err != nil {
			logs.Error(fmt.Sprintf("JIRA系统异常，JSONBody:add remark to jira failed,url:%s,err:%s", url, err.Error()))
			return err
		}

		logs.Debug(req.Response())
		//执行post接口，获取rest返回值
		if err = req.ToJSON(&result); err != nil {
			err = fmt.Errorf("JIRA系统异常, ToJSON:add remark to jira failed,url:%s,err:%s", url, err.Error())
			logs.Error(err.Error())
			return err
		}
		logs.Debug(fmt.Sprintf("ADD REMARK ISSUSE RSULT:%+v", result))
		//判断返回值，入库
		if _r, ok := result.(map[string]interface{}); ok {
			if _r["author"] == nil {
				logs.Error(fmt.Sprintf("JIRA系统异常，jira result does not have key info:%+v", result))
				return fmt.Errorf("JIRA系统异常，jira result does not have key info:%+v", result)
			}
		} else {
			logs.Error("parse key from result failed, result:", result)
			return fmt.Errorf("parse key from result failed, result:%s", result)
		}
	}
	return nil
}

// 从jira上获取项目列表
func getProjectList() (l interface{}, err error) {
	auth := beego.AppConfig.String("JIRA::Auth")
	url := beego.AppConfig.String("JIRA::ADDRESS") + "/rest/api/2/project"
	req := httplib.Get(url).Header("Authorization", "Basic "+auth).Header("Content-Type", "application/json")
	// 执行post接口，获取rest返回值
	if err = req.ToJSON(&l); err != nil {
		err = fmt.Errorf("JIRA系统异常, ToJSON:get project list from jira failed,url:%s,err:%s", url, err.Error())
		logs.Error(err.Error())
		return nil, err
	}
	return l, nil
}

// 从jira上获取项目信息
func getProject(projectid string) (p interface{}, err error) {
	auth := beego.AppConfig.String("JIRA::Auth")
	if projectid == "" {
		err = fmt.Errorf("projectid is null")
		logs.Error(err.Error())
		return
	}
	url := beego.AppConfig.String("JIRA::ADDRESS") + "/rest/api/2/project/" + projectid
	req := httplib.Get(url).Header("Authorization", "Basic "+auth).Header("Content-Type", "application/json")
	// 执行post接口，获取rest返回值
	if err = req.ToJSON(&p); err != nil {
		err = fmt.Errorf("JIRA系统异常, ToJSON:get project from jira failed, url:%s, err:%s", url, err.Error())
		logs.Error(err.Error())
		return nil, err
	}
	return p, nil
}

// 获得field可选属性
func getFieldOptions(fieldID float64) (l []FieldOption, err error) {
	if err = db.Table("customfieldoption").Where("CUSTOMFIELD = ?", fieldID).Find(&l).Error; err != nil {
		logs.Error("query options failed, CUSTOMFIELD:", fieldID)
		return nil, err
	}
	return l, nil
}

//  判断JiraKey在jira系统中是否存在
func jiraKeyIsExist(jiraKey string) (bool, error) {
	var err error
	var issue struct {
		JIRAKey string `json:"key"`
	}
	url := beego.AppConfig.String("JIRA::ADDRESS") + "/rest/api/2/issue/" + jiraKey
	auth := beego.AppConfig.String("JIRA::Auth")
	req := httplib.Get(url).Header("Authorization", "Basic "+auth).
		Header("Content-Type", "application/json")
	if err = req.ToJSON(&issue); err != nil {
		err = fmt.Errorf("JIRA系统异常, ToJSON:get issue detail failed, url:%s, err:%s", url, err.Error())
		logs.Error(err.Error())
		return false, err
	}
	if issue.JIRAKey == "" {
		return false, nil
	}
	return true, nil
}

// 判断项目在jira系统中是否存在
func ProjectIsExist(projectid string) (bool, error) {
	var err error
	var project struct {
		JIRAKey string `json:"key"`
	}
	url := beego.AppConfig.String("JIRA::ADDRESS") + "/rest/api/2/project/" + projectid
	auth := beego.AppConfig.String("JIRA::Auth")
	req := httplib.Get(url).Header("Authorization", "Basic "+auth).
		Header("Content-Type", "application/json")
	// 执行post接口，获取rest返回值
	if err = req.ToJSON(&project); err != nil {
		err = fmt.Errorf("JIRA系统异常, ToJSON:get project from jira failed, url:%s, err:%s", url, err.Error())
		logs.Error(err.Error())
		return false, err
	}
	if project.JIRAKey == "" {
		return false, nil
	}
	return true, nil
}

func init() {
	var err error
	var jiraDBSourceName = beego.AppConfig.String("JIRA::DBSourceName")
	if len(jiraDBSourceName) == 0 {
		// 无需实例化JIRA数据库
		return
	}
	newLogger := logger.New(
		logs.GetLogger(), // io writer（日志输出的目标，前缀和日志包含的内容——译者注）
		logger.Config{
			SlowThreshold:             time.Second, // 慢 SQL 阈值
			LogLevel:                  logger.Info, // 日志级别
			IgnoreRecordNotFoundError: true,        // 忽略ErrRecordNotFound（记录未找到）错误
			Colorful:                  false,       // 禁用彩色打印
		},
	)

	passwd := beego.AppConfig.String("JIRA::DBPasswd")
	if pwdEncryptKey := beego.AppConfig.String("JIRA::PwdEncryptKey"); pwdEncryptKey != "" {
		// 密码是加密形态，需要解密
		if passwd, err = cbc.New(pwdEncryptKey).Decrypt(passwd); err != nil {
			logs.Error("Decrypt JiraDB passwd failed, pwdKey: %s, ciphertext: %s, err:%s",
				pwdEncryptKey, passwd, err.Error())
			panic(err)
		}
	}

	dsn := fmt.Sprintf(jiraDBSourceName, passwd)
	if db, err = gorm.Open(mysql.Open(dsn), &gorm.Config{
		Logger: newLogger,
	}); err != nil {
		logs.Error(err.Error())
		panic(err)
	}
}
