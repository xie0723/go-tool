package attach

import (
	"bytes"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/astaxie/beego"
	"github.com/astaxie/beego/logs"
	"github.com/daimall/tools/curd/common"
	"github.com/daimall/tools/curd/dbmysql/dbgorm"
	"github.com/daimall/tools/tusclient"
	uuid "github.com/satori/go.uuid"
	"gorm.io/gorm"
)

const (
	AttachTypeSupplier = iota + 1
	AttachTypeAVL
)
const (
	FileServerTypeLocal = "local"
	FileServerTypeTUSD  = "tusd"
)

// 附件
type Attach struct {
	ID       int64  `gorm:"primary_key"  json:"id"`
	FileName string `gorm:"size:50;column:file_name;index"  json:"file_name"` //
	URL      string `gorm:"size:255;column:url"  json:"url"`                  // TUSD

	Kind string `gorm:"size:10;column:kind"  json:"kind"` // local /tusd

	Obj   string `gorm:"size:50;column:obj;index"  json:"obj"` // 所属对象
	KeyID uint   `gorm:"column:fid;index"  json:"-"`           // 对象id
	Sub   string `gorm:"size:50;column:sub;index"  json:"sub"` // 对象子分组

	CreatedAt *time.Time `json:"created_at"` // 创建时间
	TblName   string     `gorm:"-"`
	Uuid      string     `gorm:"-"` // uuid路径名
}

// 表名
func (a *Attach) TableName() string {
	if a.TblName != "" {
		return a.TblName
	}
	return "attachs"
}

// 上传附件(文件上传+记录入库)
func UploadAttachs(files []*multipart.FileHeader, obj, sub string, objId uint, tblName string, tx *gorm.DB) (attachs []Attach, err error) {
	if beego.AppConfig.String("FileServer::Kind") == FileServerTypeLocal {
		var dir string
		if attachs, dir, err = saveFiletoLocal(files, beego.AppConfig.String("FileServer::Path"),
			obj, objId, tblName); err != nil {
			return
		}
		// 记录入库
		if err = batchSaveAttach(dbgorm.GetDBInst(), attachs); err != nil {
			logs.Error("batchSaveAttach failed,", err.Error())
			os.RemoveAll(dir)
		}
	} else if beego.AppConfig.String("FileServer::Kind") == FileServerTypeTUSD {
		server := beego.AppConfig.String("FileServer::URL")
		for _, file := range files {
			var srcf multipart.File
			logs.Debug("range deal with file[%s]", file.Filename)
			if srcf, err = file.Open(); err != nil {
				logs.Error("open upload file[%s] failed, %s", file.Filename, err.Error())
				return
			}
			defer srcf.Close()
			logs.Debug("open upload file[%s] success", file.Filename)
			var fileId string
			if fileId, err = tusclient.CreateFile(server, file.Filename, file.Size); err != nil {
				logs.Error(err.Error())
				return
			}
			if err, _ = tusclient.WriteFile(srcf, fileId); err != nil {
				logs.Error(err.Error())
				return
			}
			logs.Debug("upload file[%s] to tus server success", file.Filename)
			attach := Attach{FileName: file.Filename, URL: fileId, Obj: obj, KeyID: objId,
				Sub: sub, TblName: tblName, Kind: beego.AppConfig.String("FileServer::Kind")}
			attachs = append(attachs, attach)
		}
		// 记录入库
		if err = batchSaveAttach(tx, attachs); err != nil {
			logs.Error("batchSaveAttach failed,", err.Error())
			for i := range attachs {
				tusclient.DeleteFile(attachs[i].URL)
			}
		}
	}
	return
}

// 删除附件（删除文件+数据库记录）
func DeleteAttachs(p interface{}) (err error) {
	var ids []string
	switch v := p.(type) { //v表示b1 接口转换成Bag对象的值
	case []Attach:
		for _, a := range v {
			ids = append(ids, strconv.FormatInt(a.ID, 10))
			if beego.AppConfig.String("FileServer::Kind") == FileServerTypeLocal {
				os.Remove(filepath.Join(common.GetPath([]string{a.URL})))
			} else if beego.AppConfig.String("FileServer::Kind") == FileServerTypeTUSD {
				tusclient.DeleteFile(a.URL)
			}
		}
	case []string: // ids列表
		ids = v
	case string: // 路径或文件，直接删除
		if beego.AppConfig.String("FileServer::Kind") == FileServerTypeLocal {
			os.Remove(v)
		} else if beego.AppConfig.String("FileServer::Kind") == FileServerTypeTUSD {
			tusclient.DeleteFile(v)
		}
	default:
		return fmt.Errorf("params type error")
	}
	if len(ids) > 0 {
		attach := Attach{}
		if err = dbgorm.GetDBInst().Where("id in (?)", ids).Delete(&attach).Error; err != nil {
			logs.Error("delete attach%+v failed, %s", ids, err.Error())
			return
		}
	}
	return
}

// 存储文件到本地
// path 相对路径
func saveFiletoLocal(files []*multipart.FileHeader, path string, obj string, objId uint, tblName string) (attachs []Attach, dir string, err error) {
	uuidpath := uuid.NewV4().String()
	paths := strings.FieldsFunc(path, func(r rune) bool {
		if r == '\\' || r == '/' {
			return true
		}
		return false
	})
	paths = append(paths, uuidpath)
	if len(paths) < 1 || len(uuidpath) <= 0 {
		return nil, "", fmt.Errorf("gen data path failed")
	}
	dir = common.GetPath(paths)
	defer func() {
		if err != nil {
			// 删除目录
			//os.RemoveAll(path)
		}
	}()
	if err = os.MkdirAll(dir, os.ModePerm); err != nil {
		return
	}
	for _, file := range files {
		var attach = Attach{TblName: tblName}
		attach.FileName = file.Filename
		attach.Uuid = uuidpath
		attach.Obj = obj
		attach.KeyID = objId
		var oFile multipart.File
		if oFile, err = file.Open(); err != nil {
			logs.Error("open attach file[%s] fail, %s", file.Filename, err.Error())
			return nil, "", err
		}
		defer oFile.Close()
		tofile := filepath.Join(dir, file.Filename)
		f, err := os.OpenFile(tofile, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0666)
		if err != nil {
			return nil, "", err
		}
		defer f.Close()
		if _, err = io.Copy(f, oFile); err != nil {
			return nil, "", err
		}
		attach.URL = filepath.Join(path, uuidpath, file.Filename)
		attachs = append(attachs, attach)
	}
	return
}

func batchSaveAttach(db *gorm.DB, attachs []Attach) error {
	if len(attachs) == 0 {
		logs.Warn("no attach need to save")
		return nil
	}
	tblName := attachs[0].TableName()
	var buffer bytes.Buffer
	sql := fmt.Sprintf("insert into `%s` (`file_name`,`kind`,`sub`,`url`,`obj`,`fid`,`created_at`) values", tblName)
	if _, err := buffer.WriteString(sql); err != nil {
		return err
	}
	for i, a := range attachs {
		if i == len(attachs)-1 {
			buffer.WriteString(fmt.Sprintf("('%s','%s','%s','%s','%s',%d,'%s');", a.FileName, a.Kind, a.Sub, a.URL, a.Obj, a.KeyID, time.Now().Format("2006-01-02 15:04:05")))
		} else {
			buffer.WriteString(fmt.Sprintf("('%s','%s','%s','%s','%s',%d,'%s'),", a.FileName, a.Kind, a.Sub, a.URL, a.Obj, a.KeyID, time.Now().Format("2006-01-02 15:04:05")))
		}
	}
	return db.Exec(buffer.String()).Error
}

// 辅助方法，获取附件URL拼接
func GetAttachUrls(attachs []Attach) (ret string) {
	var urls []string
	for i := range attachs {
		urls = append(urls, attachs[i].URL)
	}
	return strings.Join(urls, ",")
}

func UploadAttach(c common.BaseController, flowName string, objId uint, keys []string, tx *gorm.DB) (attachs []Attach, err error) {
	// 存储附件- 营业执照
	var files []*multipart.FileHeader
	for i := range keys {
		var l []Attach
		if files, err = c.GetFiles(keys[i]); err != nil {
			if err == http.ErrMissingFile {
				logs.Debug("no attatch for %s", keys[i])
				err = nil
				continue // 没有附件
			}
			logs.Error("get %s attach file failed, %s", keys[i], err.Error())
			err = common.ParamsErr
			return
		}
		//business_license
		if l, err = UploadAttachs(files, flowName, keys[i], objId, "", tx); err != nil {
			return
		}
		attachs = append(attachs, l...)
	}
	return
}
