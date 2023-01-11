package aes

import (
	"crypto/sha1"
	"fmt"
	"io"
)

// 获取Aes算法key
func GetPriAesKey(appid, model, timestamp string) string {
	// AppId(10位) + Model + Timestamp（十位）  取AppId的1 3 5 7 9位组成新的AppId，取Timestamp的2 4 6 8 10 位组成新的Timestamp，以新的AppId(5位) + Model + Timestamp（5位） 进行SHR1得到
	srcKey := string(appid[0]) + string(appid[2]) + string(appid[4]) + string(appid[6]) + string(appid[8]) + model + string(timestamp[1]) + string(timestamp[3]) + string(timestamp[5]) + string(timestamp[7]) + string(timestamp[9])
	aesKey, _ := GetSha1(srcKey)
	return aesKey[:32]
}

func GetSha1(src string) (string, error) {
	//TODO GET Secret from redis

	//fmt.Println("Sign str is :", signStr)
	t := sha1.New()
	io.WriteString(t, src)
	out := fmt.Sprintf("%x", t.Sum(nil))
	return out, nil
}
