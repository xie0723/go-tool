package tusclient

import (
	"encoding/json"
	"fmt"

	"github.com/astaxie/beego"
	"github.com/astaxie/beego/logs"
	"github.com/daimall/tools/curd/common"
	"github.com/daimall/tools/sign"
)

//上传文件（创建）
type TusdController struct {
	common.BaseController
}

type FileCreateParam struct {
	FileName string `json:"fileName"` // 文件名称
	Length   int64  `json:"length"`   // 文件长度
	Sign     string `json:"sign"`     // 签名结果
}

// API  创建文件
// 返回  tusd文件的url
func (c *TusdController) CreateTusdFile() {
	var err error
	var v FileCreateParam
	var ret = make(map[string]string)
	defer func() {
		if err != nil {
			c.Data["json"] = common.StandRestResult{Code: -1, Message: err.Error()}
		} else {
			c.Data["json"] = common.StandRestResult{Code: 0, Message: "OK", Data: ret}
		}
		c.ServeJSON()
	}()
	if err = json.Unmarshal(c.Ctx.Input.RequestBody, &v); err != nil {
		logs.Error("json.Unmarshal err: %s", err.Error())
		return
	}
	if v.FileName == "" {
		err = fmt.Errorf("file name is nil")
		logs.Error(err.Error())
		return
	}
	if v.Length <= 0 {
		err = fmt.Errorf("file length <= 0")
		logs.Error(err.Error())
		return
	}
	//校验签名
	if !sign.New().ToLower(true).VerifyParamsSign(v) {
		err = fmt.Errorf("valid create file parameters sign failed")
		logs.Error(err.Error())
		return
	}
	address := beego.AppConfig.String("FileServer::URL")
	ret["url"], err = CreateFile(address, v.FileName, v.Length)
}
