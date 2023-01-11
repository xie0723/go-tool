package sign

import (
	"crypto/sha1"
	"encoding/json"
	"fmt"
	"io"
	"reflect"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/astaxie/beego"
	"github.com/astaxie/beego/logs"
)

type Sign struct {
	Key       string //签名KEY
	KeyName   string //key名称，默认 AppKey
	AppId     string //应用id
	AppIdName string //应用id名称，默认AppId，Server侧的规则是AppID
	IsSign    bool   //是否需要签名验证
	Type      int    //0表示 Key和IsSign从配置文件读取，1表示调用者赋值
	IsToLower bool   //签名原字符串是否全部转化成小写字符
}

const (
	FROM_CONFIG = iota
	FROM_AUTHOR
)

// 存储Sign的map
var _insts map[int]*Sign

const defaultSignIndex = 1

//实例化单例
func New() *Sign {
	if _insts == nil {
		_insts = make(map[int]*Sign)
	}
	if v, ok := _insts[defaultSignIndex]; !ok || v == nil {
		_insts[defaultSignIndex] = new(Sign)
		_insts[defaultSignIndex].KeyName = "AppKey"
		_insts[defaultSignIndex].AppIdName = "AppId"
		_insts[defaultSignIndex].IsToLower = false
	}
	return _insts[defaultSignIndex]
}

// 新对象，避免覆盖
func NewInst(index int) *Sign {
	if _insts == nil {
		_insts = make(map[int]*Sign)
	}
	if v, ok := _insts[index]; !ok || v == nil {
		_insts[index] = new(Sign)
		_insts[index].KeyName = "AppKey"
		_insts[index].AppIdName = "AppId"
		_insts[index].IsToLower = false
	}
	return _insts[index]
}

//设置KEY值
func (s *Sign) SetKey(key string) *Sign {
	s.Key = key
	return s
}

//设置签名key的key name
func (s *Sign) SetKeyName(name string) *Sign {
	s.KeyName = name
	return s
}

//设置签名Appid的 name
func (s *Sign) SetAppIdName(name string) *Sign {
	s.AppIdName = name
	return s
}

//设置AppId值
func (s *Sign) SetAppId(appId string) *Sign {
	s.AppId = appId
	return s
}

//设置是否需要签名
func (s *Sign) SetIsSgin(isSign bool) *Sign {
	s.IsSign = isSign
	return s
}

//设置是否从配置文件读取检验KEY(0表示从配置文件读取参数)
func (s *Sign) SetType(_type int) *Sign {
	s.Type = _type
	return s
}

// 设置签名原字符串是否需要全部转换成小写字符
func (s *Sign) ToLower(lower bool) *Sign {
	s.IsToLower = lower
	return s
}

//实施签名验证
func (s *Sign) VerifyParamsSign(params interface{}) bool {
	var err error
	if s.Type == FROM_CONFIG { //需要从配置文件读取是否签名
		if s.IsSign, err = beego.AppConfig.Bool("SECURITY::NEED_SIGN"); err != nil {
			s.IsSign = false
		}
		s.Key = beego.AppConfig.String("SECURITY::SIGN_KEY")
		s.AppId = beego.AppConfig.String("SECURITY::APP_ID")
	}
	if !s.IsSign {
		logs.Debug("need not sign")
		return true
	}
	/**
	 *	校验参数和appid的签名, 如签名有误, 则返回false
	 */
	sign, signmap := struct2Map(params)
	logs.Debug("Sign:%s, Signmap:%+v", sign, signmap)
	return s.verifyMapSign(sign, signmap)
}

func (s *Sign) verifyMapSign(sign string, signmap map[string]string) bool {
	/**
	 *	校验AppId和Sign签名, 如签名有误, 则返回false
	 */
	expectedSign := s.genSign(signmap)
	if expectedSign != sign {
		logs.Error("Sign failed, expected sign is :", expectedSign, "post sign is :", sign)
	}
	return expectedSign == sign
}

//计算签名值
func (s *Sign) genSign(signmap map[string]string) string {
	//TODO GET Secret from redis
	signmap[s.KeyName] = s.Key
	signmap[s.AppIdName] = s.AppId
	keys := make([]string, len(signmap))
	i := 0
	for k := range signmap {
		keys[i] = k
		i++
	}
	sort.Strings(keys)
	signStr := ""
	var v string
	for _, k := range keys {
		v = signmap[k]
		signStr = signStr + k + v
	}
	if s.IsToLower {
		signStr = strings.ToLower(signStr)
	}
	logs.Debug("SIGNSrc:%s", signStr)
	t := sha1.New()
	if _, err := io.WriteString(t, signStr); err != nil {
		logs.Error("io.WriteString(t, signStr) failed,", err.Error())
	}
	out := fmt.Sprintf("%x", t.Sum(nil))
	return out
}

//将struct 对象转换成map，方便验证签名(并返回请求端的签名结果，以便校验)
func struct2Map(params interface{}) (string, map[string]string) {
	var sign string // 签名结果字符串
	signmap := map[string]string{}
	paramValRef := reflect.ValueOf(params)
	paramTypRef := reflect.TypeOf(params)
	for i := 0; i < paramTypRef.NumField(); i++ {
		var v string
		fieldTyp := paramTypRef.Field(i)
		fieldVal := paramValRef.Field(i)
		//判断是否是需要签名的字段
		if strings.ToLower(fieldTyp.Tag.Get("sign")) == "no" {
			continue
		}
		knd := fieldTyp.Type.Kind()
		switch knd {
		case reflect.Bool:
			v = strconv.FormatBool(fieldVal.Bool())
		case reflect.Uint:
			v = strconv.FormatUint(fieldVal.Uint(), 10)
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32:
			v = strconv.Itoa(int(fieldVal.Int()))
		case reflect.Int64:
			v = strconv.FormatInt(fieldVal.Int(), 10)
		case reflect.Float32, reflect.Float64:
			v = strconv.FormatFloat(fieldVal.Float(), 'f', -1, 64)
		case reflect.String:
			v = fieldVal.String()
		case reflect.Struct:
			if fieldTyp.Type.Name() == "Time" {
				if vv, ok := fieldVal.Interface().(time.Time); ok {
					v = vv.Format("2006-01-02 15:04:05")
				}
			} else if fieldTyp.Tag.Get("sign") == "body" {
				b, _ := json.Marshal(fieldVal.Interface())
				v = string(b)
			} else {
				_tmpSign, _tmpMap := struct2Map(fieldVal.Interface())
				for k, v := range _tmpMap {
					signmap[k] = v
				}
				if _tmpSign != "" {
					sign = _tmpSign
				}
				continue
			}
		default:
			continue
		}
		if fieldTyp.Name == "Sign" {
			sign = v
			continue
		}
		signmap[fieldTyp.Name] = v
	}
	return sign, signmap
}
