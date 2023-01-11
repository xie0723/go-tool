package txcloud

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"

	"github.com/astaxie/beego"
	"github.com/astaxie/beego/logs"
	"github.com/tencentyun/cos-go-sdk-v5"
)

// 获取腾讯云client实例
func getTxCloudClient() (client *cos.Client, err error) {
	var u *url.URL

	txUrl := beego.AppConfig.String("TXCloud::URL")
	secretID := beego.AppConfig.String("TXCloud::SecretID")
	secretKey := beego.AppConfig.String("TXCloud::SecretKey")

	u, err = url.Parse(txUrl)
	if err != nil {
		return
	}
	b := &cos.BaseURL{BucketURL: u}
	client = cos.NewClient(b, &http.Client{
		Transport: &cos.AuthorizationTransport{
			SecretID:  secretID,
			SecretKey: secretKey,
		},
	})

	return
}

// 上传文件到腾讯云(图片，合同等持久化保存等文件)
func UploadFile(f io.Reader, fpath string) (downloadURL string, err error) {
	var client *cos.Client

	client, err = getTxCloudClient()
	if err != nil {
		logs.Error("get txCloud client failed,", err.Error())
		return
	}

	// 获取文件类型
	contentType, err := getContentType(fpath)
	if err != nil {
		logs.Error("get contentType failed,", err.Error())
		return
	}

	opt := &cos.ObjectPutOptions{
		ObjectPutHeaderOptions: &cos.ObjectPutHeaderOptions{
			ContentType: contentType,
		},
	}

	_, err = client.Object.Put(context.Background(), fpath, f, opt)
	if err != nil {
		return
	}
	downloadURL = beego.AppConfig.String("TXCloud::URL") + fpath
	return
}

// 通过匹配文件后缀获取contentType
func getContentType(fpath string) (contentType string, err error) {
	reg, err := regexp.Compile("\\.\\w+$")
	if err != nil {
		return
	}
	fileType := reg.FindString(fpath)
	if fileType == "" {
		err = errors.New("get file type failed")
		return
	}
	switch fileType {
	case ".jpg":
		contentType = "image/jpeg"
	case ".jpeg":
		contentType = "image/jpeg"
	case ".png":
		contentType = "image/png"
	case ".pdf":
		contentType = "application/pdf"
	case ".xml":
		contentType = "text/xml"
	}
	return
}

// 删除单个文件
// key:对象键（Key）是对象在存储桶中的唯一标识。例如，在对象的访问域名examplebucket-1250000000.cos.ap-guangzhou.myqcloud.com/doc/pic.jpg中，对象键为 doc/pic.jpg
func DeleteFile(key string) (err error) {
	var (
		client *cos.Client
		resp   *cos.Response
	)

	client, err = getTxCloudClient()
	if err != nil {
		logs.Error("get txCloud client failed,", err.Error())
		return
	}
	resp, err = client.Object.Delete(context.Background(), key)
	if err != nil {
		logs.Error("delete file failed,", err.Error())
		return
	}
	if resp.StatusCode != 200 {
		var body []byte
		if body, err = io.ReadAll(resp.Body); err != nil {
			logs.Error("read response body failed,", err.Error())
			return
		}
		resp.Body.Close()
		err = fmt.Errorf(string(body))
		return
	}

	return
}

// 删除多个文件
func DeleteFiles(keys []string) (err error) {
	var (
		client *cos.Client
		resp   *cos.Response
		obs    []cos.Object
	)

	client, err = getTxCloudClient()
	if err != nil {
		logs.Error("get txCloud client failed,", err.Error())
		return
	}
	for _, v := range keys {
		obs = append(obs, cos.Object{Key: v})
	}
	opt := &cos.ObjectDeleteMultiOptions{
		Objects: obs,
		// 布尔值，这个值决定了是否启动 Quiet 模式
		// 值为 true 启动 Quiet 模式，值为 false 则启动 Verbose 模式，默认值为 false
		// Quiet: true,
	}

	_, resp, err = client.Object.DeleteMulti(context.Background(), opt)
	if err != nil {
		var body []byte
		if body, err = io.ReadAll(resp.Body); err != nil {
			logs.Error("read response body failed,", err.Error())
			return
		}
		resp.Body.Close()
		err = fmt.Errorf(string(body))
		return
	}

	return
}
