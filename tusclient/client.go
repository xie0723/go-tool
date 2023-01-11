package tusclient

import (
	"bytes"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"

	"github.com/astaxie/beego/httplib"
	"github.com/astaxie/beego/logs"
)

// 创建文件
func CreateFile(address, filename string, fileSize int64) (fileUrl string, err error) {
	response, err := httplib.Post(address).Header("Upload-Length", strconv.FormatInt(fileSize, 10)).Header("Tus-Resumable", "1.0.0").
		Header("Upload-Metadata", "filename "+base64.StdEncoding.EncodeToString([]byte(filename))).DoRequest()
	if err != nil {
		logs.Error("Create file[%s] on tusd server[%s] failed, err:%s",
			filename, address, err.Error())
		return
	}
	if response.StatusCode == http.StatusCreated {
		fileUrl = response.Header.Get("Location")
	} else {
		err = fmt.Errorf("create file[%s] on tusd server[%s] failed, response.Status[%s]",
			filename, address, response.Status)
		logs.Error(err.Error())
	}
	return
}

// GetUploadProgress 获取文件上传进度
func GetUploadInfo(fileUrl string) (offset, length int64, fileName string, err error) {
	response, err := httplib.Head(fileUrl).Header("Tus-Resumable", "1.0.0").DoRequest()
	if err != nil {
		logs.Error("get[Head] file[%s] upload info failed, %s", fileUrl, err.Error())
		return
	}
	if offset, err = strconv.ParseInt(response.Header.Get("Upload-Offset"), 10, 64); err != nil {
		logs.Error("get offset of file[%s] from Header failed, %s", fileUrl, err.Error())
		return 0, 0, "", err
	}
	if length, err = strconv.ParseInt(response.Header.Get("Upload-Length"), 10, 64); err != nil {
		logs.Error("get offset of file[%s] from Header failed, %s", fileUrl, err.Error())
		return offset, 0, "", err
	}
	var v []byte
	if v, err = base64.StdEncoding.DecodeString(strings.ReplaceAll(response.Header.Get("Upload-Metadata"), "filename ", "")); err != nil {
		logs.Error("get file name of file[%s] from Header failed, %s", fileUrl, err.Error())
		return offset, length, "", err
	}
	fileName = string(v)
	return
}

func WriteFile(r io.ReadSeeker, fileID string) (error, int64) {
	offset, _, _, err := GetUploadInfo(fileID)
	if err != nil {
		return err, 0
	}
	offset, err = r.Seek(offset, 0)
	if err != nil {
		return err, 0
	}

	buff := make([]byte, 32*1024, 32*1024)
	for {
		n, err := r.Read(buff)
		if err != nil {
			if err == io.EOF {
				err = nil
				break
			}
			return err, 0
		}
		d := buff[:n]

		req, err := http.NewRequest("PATCH", fileID, bytes.NewBuffer(d))
		if err != nil {
			return err, 0
		}
		req.Header.Add("Content-Type", "application/offset+octet-stream")
		req.Header.Add("Upload-Offset", strconv.FormatInt(offset, 10))
		req.Header.Add("Tus-Resumable", "1.0.0")
		response, err := http.DefaultClient.Do(req)
		if err != nil {
			return err, 0
		}
		if response.StatusCode != http.StatusNoContent {
			return errors.New("WriteFileInServer, err:" + response.Status), 0
		}
		offset += int64(n)
	}
	return nil, offset
}

// 删除Tusd文件，可变参数，每一个可以是多个文件（逗号连接）
func DeleteFile(fileUrls ...string) error {
	var response *http.Response
	var err error
	for _, commaUrls := range fileUrls {
		for _, v := range strings.Split(commaUrls, ",") {
			if response, err = httplib.Delete(v).Header("Tus-Resumable", "1.0.0").DoRequest(); err == nil {
				if response.StatusCode == 204 {
					logs.Debug("delete file[%s] succ.", v)
					continue
				}
				err = errors.New(response.Status)
			}
			logs.Error("delete file[%s] failed, %s", v, err.Error())
			return err
		}
	}
	return nil
}
