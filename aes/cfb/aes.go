package cfb

import (
	"crypto/aes"
	"crypto/cipher"
	"encoding/base64"
	"errors"
)

type encryptAes struct {
	aesKey      string
	base64Table string
}

func New(aesKey, base64Table string) *encryptAes {
	instance := new(encryptAes)
	instance.aesKey = aesKey
	instance.base64Table = base64Table
	return instance
}

//Encrypt 加密方法
func (e *encryptAes) Encrypt(plainstr string) (encrypted string, err error) {
	var aesEncryptBytes []byte
	if aesEncryptBytes, err = aesEncrypt(plainstr, e.aesKey); err != nil {
		return
	}
	return base64Encode(aesEncryptBytes, e.base64Table), nil
}

//Decrypt 解密方法
func (e *encryptAes) Decrypt(cipherstr string) (claimed string, err error) {
	var aesEncryptBytes []byte
	if aesEncryptBytes, err = base64Decode(cipherstr, e.base64Table); err != nil {
		return
	}
	return aesDecrypt(aesEncryptBytes, e.aesKey)
}

/* 私有方法 */
func base64Encode(encodeByte []byte, base64Table string) string {
	var coder = base64.NewEncoding(base64Table)
	return coder.EncodeToString(encodeByte)
}
func base64Decode(decodeStr string, base64Table string) ([]byte, error) {
	var coder = base64.NewEncoding(base64Table)
	return coder.DecodeString(decodeStr)
}

func aesEncrypt(strMesg, aesStrKey string) (encrypted []byte, err error) {
	var key []byte
	if key, err = getAesKey(aesStrKey); err != nil {
		return
	}
	var iv = key[:aes.BlockSize]
	encrypted = make([]byte, len(strMesg))
	aesBlockEncrypter, err := aes.NewCipher(key)
	if err != nil {
		return
	}
	aesEncrypter := cipher.NewCFBEncrypter(aesBlockEncrypter, iv)
	aesEncrypter.XORKeyStream(encrypted, []byte(strMesg))
	return encrypted, nil
}

func aesDecrypt(src []byte, aesStrKey string) (strDesc string, err error) {
	defer func() {
		//错误处理
		if e := recover(); e != nil {
			err = e.(error)
		}
	}()
	var key []byte
	if key, err = getAesKey(aesStrKey); err != nil {
		return
	}
	var iv = []byte(key)[:aes.BlockSize]
	decrypted := make([]byte, len(src))
	var aesBlockDecrypter cipher.Block
	aesBlockDecrypter, err = aes.NewCipher([]byte(key))
	if err != nil {
		return
	}
	aesDecrypter := cipher.NewCFBDecrypter(aesBlockDecrypter, iv)
	aesDecrypter.XORKeyStream(decrypted, src)
	return string(decrypted), nil
}

func getAesKey(strKey string) (aesKey []byte, err error) {
	keyLen := len(strKey)
	if keyLen < 16 {
		err = errors.New("the length of Aes str key is less then 16:," + strKey)
		return
	}
	arrKey := []byte(strKey)
	if keyLen >= 32 {
		return arrKey[:32], nil //取前32个字节
	}
	if keyLen >= 24 {
		return arrKey[:24], nil //取前24个字节
	}
	return arrKey[:16], nil //取前16个字节
}
