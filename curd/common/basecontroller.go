package common

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/daimall/tools/aes"
	"github.com/daimall/tools/aes/cbc"
	"io"
	"mime/multipart"
	"os"

	"github.com/astaxie/beego"
	"github.com/astaxie/beego/logs"
	"github.com/daimall/tools/curd/customerror"

	"gopkg.in/go-playground/validator.v9"
)

var validate *validator.Validate

func init() {
	validate = validator.New()
}

//BaseController ...
//所有控制controller的基础struct
type BaseController struct {
	REST             StandRestResultInf
	beego.Controller // beego 基础控制器
}

func (b *BaseController) GetStandRestResult() StandRestResultInf {
	if b.REST == nil {
		return StandRestResult{}
	}
	return b.REST
}

// JSONResponse 返回JSON格式结果
func (c *BaseController) JSONResponse(err error, data ...interface{}) {
	if err != nil {
		if cuserr, ok := err.(customerror.CustomError); ok {
			c.Data["json"] = c.GetStandRestResult().GetStandRestResult(cuserr.GetCode(), cuserr.GetMessage(), nil)
		} else {
			c.Data["json"] = c.GetStandRestResult().GetStandRestResult(-1, err.Error(), nil)
		}
		logs.Error("JSONResponse:", err.Error())
	} else {
		if len(data) == 1 {
			c.Data["json"] = c.GetStandRestResult().GetStandRestResult(0, "OK", data[0])
		} else {
			c.Data["json"] = c.GetStandRestResult().GetStandRestResult(0, "OK", data)
		}
	}
	c.ServeDecryptJSON()
}

//ValidateParameters obj must pointer, json Unmarshal object and require parameter validate
func (c *BaseController) ValidateParameters(obj interface{}) customerror.CustomError {
	if err := json.Unmarshal(c.Ctx.Input.RequestBody, obj); err != nil {
		logs.Error("UnmarshalRequestBody Err", err.Error())
		return ParamsError
	}
	if err := validate.Struct(obj); err != nil {
		logs.Error("validate params err:", err.Error())
		return ParamsValidateError
	}
	return nil

}

// SaveToFileCustomName saves uploaded file to new path with custom name.
// it only operates the first one of mutil-upload form file field.
func (c *BaseController) SaveToFileCustomName(fromfile string, fc func(*multipart.FileHeader) string) error {
	file, h, err := c.Ctx.Request.FormFile(fromfile)
	if err != nil {
		return err
	}
	defer file.Close()
	f, err := os.OpenFile(fc(h), os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0666)
	if err != nil {
		return err
	}
	defer f.Close()
	io.Copy(f, file)
	return nil
}

const (
	ENCRYPT_TYPE                   = "EncryptType"       // 加密类型KEY
	ENCRYPT_TYPE_AES_PRIV_REQ      = "AES_PRIV_REQ"      // 请求加密
	ENCRYPT_TYPE_AES_PRIV_RESP     = "AES_PRIV_RESP"     // 响应加密
	ENCRYPT_TYPE_AES_PRIV_REQ_RESP = "AES_PRIV_REQ_RESP" // 请求和响应都加密
	ERR_CODE_ENCRYPT_FAILED        = 9001                // 加密响应body失败
)

// ServeSelfJSON ...
// Controller 自定义方法，用来处理加密返回
func (c *BaseController) ServeDecryptJSON() {
	if c.Ctx.Input.Header(ENCRYPT_TYPE) == ENCRYPT_TYPE_AES_PRIV_RESP ||
		c.Ctx.Input.Header(ENCRYPT_TYPE) == ENCRYPT_TYPE_AES_PRIV_REQ_RESP {
		var err error
		defer func() {
			if err != nil {
				c.Data["json"] = RestResult{Code: ERR_CODE_ENCRYPT_FAILED,
					Message: "Encrypt response failed," + err.Error()}
				c.ServeJSON()
			}
		}()
		var origData []byte
		var decryptedData []byte
		if origData, err = json.Marshal(c.Data["json"]); err != nil {
			logs.Error("marshal c.Data failed,", err.Error())
			return
		}
		appid := c.Ctx.Input.Header("BsAppID")
		model := c.Ctx.Input.Header("Model")
		timestamp := c.Ctx.Input.Header("Timestamp")
		if len(appid) < 10 || len(timestamp) < 10 {
			err = fmt.Errorf("appid[%s] or timestamp[%s] invalid", appid, timestamp)
			logs.Error(err.Error())
			return
		}
		aesKey := aes.GetPriAesKey(appid, model, timestamp)
		if decryptedData, err = cbc.NewPri(aesKey).EncryptBytes(origData); err != nil {
			logs.Error("encrypt response body failed,", err.Error())
			return
		}
		decryptedBytes := make([]byte, base64.StdEncoding.EncodedLen(len(decryptedData)))
		base64.StdEncoding.Encode(decryptedBytes, decryptedData)
		c.Ctx.Output.Header("Content-Type", "application/json; charset=utf-8")
		if err = c.Ctx.Output.Body(decryptedBytes); err != nil {
			logs.Error("set response body failed", err.Error())
			return
		}
	} else {
		c.ServeJSON()
	}
}
