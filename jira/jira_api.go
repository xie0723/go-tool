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

// Field中的属性值
type JiraBug struct {
	Project     map[string]string        `json:"project"`                     // 项目
	IssueType   map[string]string        `json:"issuetype"`                   // 问题类型
	Summary     string                   `json:"summary"`                     // 主题
	Description string                   `json:"description"`                 // 描述
	VersionType map[string]string        `json:"customfield_17101,omitempty"` // 版本类型
	Areas       map[string]string        `json:"customfield_17302"`           // 领域
	Components  []map[string]interface{} `json:"components"`                  // 模块
	Versions    []map[string]interface{} `json:"versions"`                    // 影响版本
	Level       map[string]string        `json:"customfield_15121"`           // 严重程度
	RecurRate   map[string]string        `json:"customfield_17400"`           // 重现概率
	Occurdate   string                   `json:"customfield_17200"`           // 发生时间
}

type FieldOption struct {
	Id          string  `gorm:"column:ID"`
	CUSTOMFIELD float64 `gorm:"column:CUSTOMFIELD"`
	Value       string  `gorm:"column:customvalue"`
}

// 添加jira问题单
func AddJiraIssue(jiraBug *JiraBug, auth string) (ret interface{}, err error) {
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
		Fields *JiraBug `json:"fields"`
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

//AddJiraRemark 添加jira问题备注
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

// 判断JiraKey在jira系统中是否存在
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
