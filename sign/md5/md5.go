package md5

import (
	"crypto/md5"
	"encoding/json"
	"fmt"
	"io"
	"reflect"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/astaxie/beego/logs"
)

//生成 MD5值

func (s *MD5) GenMD5(params interface{}) string {
	md5map := struct2Map(params)
	return s.genMD5(md5map)
}

type MD5 struct {
	IsToLower bool //签名原字符串是否全部转化成小写字符
}

var ins *MD5
var once sync.Once

//实例化单例
func GetIns() *MD5 {
	once.Do(func() {
		ins = &MD5{IsToLower: true}
	})
	return ins
}

// 设置签名原字符串是否需要全部转换成小写字符
func (s *MD5) ToLower(lower bool) *MD5 {
	s.IsToLower = lower
	return s
}

//计算MD5 值
func (s *MD5) genMD5(md5map map[string]string) string {
	//TODO GET Secret from redis
	keys := make([]string, len(md5map))
	i := 0
	for k := range md5map {
		keys[i] = k
		i++
	}
	sort.Strings(keys)
	md5str := ""
	var v string
	for _, k := range keys {
		v = md5map[k]
		md5str = md5str + k + v
	}
	if s.IsToLower {
		md5str = strings.ToLower(md5str)
	}
	logs.Debug("MD5 src:%s", md5str)
	w := md5.New()
	if _, err := io.WriteString(w, md5str); err != nil {
		logs.Error("io.WriteString(t, md5str) failed,", err.Error())
	}
	out := fmt.Sprintf("%x", w.Sum(nil))
	return out
}

//将struct 对象转换成map，方便获取需要MD5的属性
func struct2Map(params interface{}) map[string]string {
	md5map := map[string]string{}
	var paramValRef reflect.Value
	var paramTypRef reflect.Type
	v := reflect.ValueOf(params)
	if v.CanInterface() {
		switch v.Kind() {
		case reflect.Ptr:
			paramValRef = v.Elem()
			paramTypRef = reflect.TypeOf(params).Elem()
		case reflect.Struct:
			paramValRef = v
			paramTypRef = reflect.TypeOf(params)
		default:
			logs.Error("params type not support1,%s", v.Kind())
		}
	} else {
		logs.Error("params type not support2,%s", v.CanInterface())
	}

	for i := 0; i < paramTypRef.NumField(); i++ {
		var v string
		fieldTyp := paramTypRef.Field(i)
		fieldVal := paramValRef.Field(i)
		//判断是否是需要MD5的字段
		if strings.ToLower(fieldTyp.Tag.Get("md5")) == "" {
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
			} else if fieldTyp.Tag.Get("md5") == "body" {
				b, _ := json.Marshal(fieldVal.Interface())
				v = string(b)
			} else {
				_tmpMap := struct2Map(fieldVal.Interface())
				for k, v := range _tmpMap {
					md5map[k] = v
				}
				continue
			}
		default:
			continue
		}
		md5map[fieldTyp.Name] = v
	}
	return md5map
}
