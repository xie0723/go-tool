package ldapctr

import (
	"fmt"
	"strings"

	"github.com/astaxie/beego/logs"
	"github.com/daimall/tools/curd/common"
	"github.com/daimall/tools/ldapm"
	ldap "github.com/go-ldap/ldap/v3"
)

type LdapController struct {
	common.BaseController
}

func (c *LdapController) URLMapping() {
	c.Mapping("GetUserByFilter", c.GetUserByFilter)
}

type User struct {
	Name      string `json:"label"` // 姓名
	AccountId string `json:"value"` // 域账号ID
}

// ldap查询用户列表
func (c *LdapController) GetUserByFilter() {
	var err error
	var filter string
	var ret interface{}
	defer func() {
		if err == nil {
			c.Data["json"] = c.GetStandRestResult().GetStandRestResult(0, "OK", ret)
		} else {
			c.Data["json"] = c.GetStandRestResult().GetStandRestResult(-1, err.Error(), nil)
		}
		c.ServeJSON()
	}()
	if filter = strings.TrimSpace(c.GetString("filter")); filter == "" {
		ret = []struct{}{}
		return
	}
	var users []*User
	if users, err = GetUserByFilter(filter); err != nil {
		logs.Error(err.Error())
		return
	}
	ret = users
}

// 根据Name或AccountId获取用户信息
func GetUserByFilter(filter string) (users []*User, err error) {
	var _users interface{}
	users = []*User{}
	_users, err = ldapm.New().LoadBeegoConf().GetUsers(filter, func(srs []*ldap.Entry) (ret interface{}, err error) {
		var users = []*User{}
		if len(srs) == 0 {
			err = fmt.Errorf("result of comm.GetUsers userEntrys is nil, filter:%s", filter)
			logs.Error(err.Error())
			return users, nil
		}
		for _, sr := range srs {
			var _user User
			if _user, err = entry2User(sr); err != nil {
				err = fmt.Errorf("entry2User failed, %s", err.Error())
				logs.Error(err.Error())
				return
			}
			_user.Name = fmt.Sprintf("%s(%s)", _user.Name, _user.AccountId)
			users = append(users, &_user)
		}
		return users, nil
	})
	if err != nil {
		logs.Error(err.Error())
		return
	}
	if users, ok := _users.([]*User); ok {
		return users, nil
	}
	err = fmt.Errorf("result of GetUsers is not []*User object")
	logs.Error(err.Error())
	return
}

func entry2User(userEntry *ldap.Entry) (ret User, err error) {
	if userEntry == nil {
		err = fmt.Errorf("result of comm.ValidateUser userEntry is nil")
		logs.Error(err.Error())
		return
	}
	__user := User{}
	__user.AccountId = userEntry.GetAttributeValue("sAMAccountName")
	__user.Name = userEntry.GetAttributeValue("cn")
	return __user, nil
}
