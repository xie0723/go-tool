package ldapm

import (
	"errors"
	"fmt"
	"sync"

	"github.com/astaxie/beego"
	"github.com/astaxie/beego/logs"
	"github.com/daimall/tools/aes/cbc"
	ldap "github.com/go-ldap/ldap/v3"
)

type ldapAdapter struct {
	IP     string //ldap ip address
	Port   int    //ldap port
	Region string //区域
	RootDN string //根域
	User   string //管理员用户名
	Passwd string //管理员用户密码
}

var once sync.Once
var ins *ldapAdapter

//初始化实例
func New() *ldapAdapter {
	once.Do(func() {
		ins = new(ldapAdapter)
	})
	return ins
}

//设置ldap ip和端口
func (l *ldapAdapter) SetAddress(ip string, port int) *ldapAdapter {
	l.IP = ip
	l.Port = port
	return l
}

//设置ldap rootdn
func (la *ldapAdapter) SetRootDN(rootdn string) *ldapAdapter {
	la.RootDN = rootdn
	return la
}

//设置ldap 登录账号
func (la *ldapAdapter) SetAccount(user, passwd string) *ldapAdapter {
	la.User = user
	la.Passwd = passwd
	return la
}

//从配置文件中读取参数
func (la *ldapAdapter) LoadBeegoConf() *ldapAdapter {
	la.IP = beego.AppConfig.String("LDAP::ADDR")
	la.Port, _ = beego.AppConfig.Int("LDAP::PORT")
	la.Region = beego.AppConfig.String("LDAP::REGION")
	la.RootDN = beego.AppConfig.String("LDAP::ROOTDN")
	la.User = beego.AppConfig.String("LDAP::BIND_USER")
	la.Passwd = beego.AppConfig.String("LDAP::BIND_PWD")
	if pwdEncryptKey := beego.AppConfig.String("LDAP::PwdEncryptKey"); pwdEncryptKey != "" {
		// 密码是加密形态，需要解密
		var err error
		if la.Passwd, err = cbc.New(pwdEncryptKey).Decrypt(la.Passwd); err != nil {
			logs.Error("Decrypt ladp passwd failed, pwdKey: %s, ciphertext: %s, err:%s",
				pwdEncryptKey, la.Passwd, err.Error())
		}
	}
	return la
}

func (la *ldapAdapter) GetUsers(name string, f func([]*ldap.Entry) (interface{}, error)) (ret interface{}, err error) {
	var l *ldap.Conn
	var sr *ldap.SearchResult
	if l, sr, err = la.ldapUserSearch("(&(objectcategory=person)(CN=*" + name + "*))"); err != nil {
		return nil, fmt.Errorf("GetUsers(CN)>ldapUserSearch failed, %s", err.Error())
	}
	defer l.Close()
	if len(sr.Entries) == 0 {
		var l2 *ldap.Conn
		var sr2 *ldap.SearchResult
		if l2, sr2, err = la.ldapUserSearch("(&(objectcategory=person)(sAMAccountName=*" + name + "*))"); err != nil {
			return nil, fmt.Errorf("GetUsers(sAMAccountName)>ldapUserSearch failed, %s", err.Error())
		}
		defer l2.Close()
		return f(sr2.Entries)
	}
	return f(sr.Entries)
}

func (la *ldapAdapter) GetUserInfo(account string, f func(*ldap.Entry) (interface{}, error)) (ret interface{}, err error) {
	var l *ldap.Conn
	var sr *ldap.SearchResult
	if l, sr, err = la.ldapUserSearch("(sAMAccountName=" + account + ")"); err != nil {
		if len(sr.Entries) == 0 {
			return nil, errors.New("no user found by account: " + account)
		}
	}
	defer l.Close()
	return f(sr.Entries[0])
}

//ValidateUser 校验用户
func (la *ldapAdapter) ValidateUser(username, userpwd string, f func(*ldap.Entry) (interface{}, error)) (ret interface{}, err error) {
	var entry *ldap.Entry
	if entry, err = la.ladpAuth(username, userpwd); err != nil {
		return nil, err
	}
	return f(entry)
}

//ladpAuth ....
//ladp 认证
func (la *ldapAdapter) ladpAuth(username, userpwd string) (ret *ldap.Entry, err error) {
	l, sr, err := la.ldapUserSearch("(sAMAccountName=" + username + ")")
	defer l.Close()
	if err != nil {
		return nil, fmt.Errorf("ldapUserSearch failed, %s", err.Error())
	}
	if err != nil || len(sr.Entries) < 1 {
		err = errors.New("user:" + username + " does not exist")
		return nil, err
	}
	if err = l.Bind(username+"@"+la.Region, userpwd); err != nil {
		logs.Error("Passwd of user:", username, "error", err.Error())
		return nil, fmt.Errorf("Passwd of user:%s error, %s", username, err.Error())
	}
	return sr.Entries[0], nil
}

func (la *ldapAdapter) ldapUserSearch(filter string) (l *ldap.Conn, sr *ldap.SearchResult, err error) {
	if err != nil {
		logs.Error("ldap port is not int: ", la.Port)
		return nil, nil, err
	}
	l, err = ldap.Dial("tcp", fmt.Sprintf("%s:%d", la.IP, la.Port))
	if err != nil {
		beego.Error("Test LDAP Connect Error: ", err.Error())
		return
	}
	err = l.Bind(la.User, la.Passwd)
	if err != nil {
		logs.Error("login ldap server failed, username:", la.User, err.Error())
		return
	}
	attributes := []string{}
	searchRequest := ldap.NewSearchRequest(
		la.RootDN,
		2, 3, 0, 0, false,
		filter,
		attributes,
		nil)
	sr, err = l.Search(searchRequest)
	return
}
