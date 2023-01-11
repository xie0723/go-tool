package redisclient

import (
	"os"

	"github.com/astaxie/beego"
	"github.com/astaxie/beego/logs"
	"github.com/daimall/tools/aes/cbc"
	"github.com/go-redis/redis"
)

var redisCli *redis.Client

func GetInst() *redis.Client {
	if redisCli == nil {
		initRedisClient()
	}
	return redisCli
}

// 初始化
func initRedisClient() {
	//注册 lua 脚本执行器
	var redisOptions = &redis.Options{
		Addr: beego.AppConfig.String("Redis::URL"),
	}
	var err error
	redisPasswd := beego.AppConfig.String("Redis::Password")
	redisOptions.Password = redisPasswd
	// 密码解密
	if pwdEncryptKey := beego.AppConfig.String("Redis::PwdEncryptKey"); pwdEncryptKey != "" {
		// 密码是加密形态，需要解密
		if redisOptions.Password, err = cbc.New(pwdEncryptKey).Decrypt(redisPasswd); err != nil {
			logs.Error("Decrypt redis passwd failed, pwdKey: %s, ciphertext: %s, err:%s",
				pwdEncryptKey, redisPasswd, err.Error())
			os.Exit(-101)
		}
	}
	if redisOptions.DB, err = beego.AppConfig.Int("Redis::DB"); err != nil {
		logs.Error("get redis db failed,", err.Error())
		os.Exit(-102)
	}
	redisCli = redis.NewClient(redisOptions)
	if _, err = redisCli.Ping().Result(); err != nil {
		logs.Error("redis ping failed,", err.Error())
		os.Exit(-103)
	}
}
