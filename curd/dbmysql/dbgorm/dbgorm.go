package dbgorm

import (
	"fmt"
	"time"

	"github.com/astaxie/beego"
	"github.com/astaxie/beego/logs"
	"github.com/daimall/tools/aes/cbc"
	"gorm.io/driver/mysql"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	_ "modernc.org/sqlite"
)

var db *gorm.DB

// 获取 gorm数据库链接实例
func GetDBInst() *gorm.DB {
	if db == nil {
		db = NewDBInst()
	}
	return db
}

func NewDBInst() *gorm.DB {
	return newDBInst("DB")
}

func NewDBInstBySection(section string) *gorm.DB {
	return newDBInst(section)
}

func newDBInst(section string) *gorm.DB {
	var db *gorm.DB
	var err error
	newLogger := logger.New(
		logs.GetLogger(), // io writer（日志输出的目标，前缀和日志包含的内容——译者注）
		logger.Config{
			SlowThreshold:             time.Second, // 慢 SQL 阈值
			LogLevel:                  logger.Info, // 日志级别
			IgnoreRecordNotFoundError: true,        // 忽略ErrRecordNotFound（记录未找到）错误
			Colorful:                  false,       // 禁用彩色打印
		},
	)
	if beego.AppConfig.String(section+"::Driver") == "mysql" {
		passwd := beego.AppConfig.String(section + "::Passwd")
		if pwdEncryptKey := beego.AppConfig.String(section + "::PwdEncryptKey"); pwdEncryptKey != "" {
			// 密码是加密形态，需要解密
			if passwd, err = cbc.New(pwdEncryptKey).Decrypt(passwd); err != nil {
				logs.Error("Decrypt db passwd failed, pwdKey: %s, ciphertext: %s, err:%s",
					pwdEncryptKey, passwd, err.Error())
				panic(err)
			}
		}
		dsn := fmt.Sprintf(beego.AppConfig.String(section+"::SourceName"), passwd)
		if db, err = gorm.Open(mysql.Open(dsn), &gorm.Config{
			Logger: newLogger,
		}); err != nil {
			logs.Error(err.Error())
			panic(err)
		}
		return db
	} else if beego.AppConfig.String(section+"::Driver") == "sqlite3" {
		dbfile := beego.AppConfig.String(section + "::SourceName")
		if db == nil {
			if db, err = gorm.Open(sqlite.Dialector{DSN: dbfile, DriverName: "sqlite"}, &gorm.Config{
				Logger: newLogger,
			}); err != nil {
				logs.Error("open lottery.db failed,", err.Error())
				panic(err.Error())
			}
		}
		return db
	} else {
		logs.Error("unkown Driver")
	}
	return db
}
