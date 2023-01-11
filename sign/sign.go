package sign

import (
	"github.com/astaxie/beego"
	"github.com/astaxie/beego/logs"
)

//获取签名sgin
func (s *Sign) GenSign(params interface{}) string {
	var err error
	if s.Type == FROM_CONFIG { //需要从配置文件读取是否签名
		if s.IsSign, err = beego.AppConfig.Bool("SECURITY::NEED_SIGN"); err != nil {
			s.IsSign = false
		}
		s.Key = beego.AppConfig.String("SECURITY::SIGN_KEY")
		s.AppId = beego.AppConfig.String("SECURITY::APP_ID")
	}
	_, signmap := struct2Map(params)
	return s.genSign(signmap)
}

// 和server保持一致的签名方案
// body + appId+appkey + timestamp
func (s *Sign) VerifyMapSign(sign string, signmap map[string]string) bool {
	if !s.IsSign {
		logs.Debug("need not sign")
		return true
	}
	return s.verifyMapSign(sign, signmap)
}
