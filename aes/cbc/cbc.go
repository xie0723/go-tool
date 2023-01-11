package cbc

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
)

const (
	//OFFSET 补齐字符的ascii码值与个数的偏移量
	OFFSET               = 0
	BLOCK_SIZE           = 16
	IV_TYPE_KEY          = 1
	IV_TYPE_RAND         = 2
	ENCRYPT_TYPE_COMMON  = 1
	ENCRYPT_TYPE_PRIVATE = 2
)

var _insts = make(map[int]*EncryptCBC)

//EncryptCBC ... AES CBC模式的加密算法
type EncryptCBC struct {
	key       []byte
	blockSize int
	ivType    int //
}

//New ...
func New(key string) *EncryptCBC {
	if _insts[ENCRYPT_TYPE_COMMON] == nil {
		inst := new(EncryptCBC)
		inst.key = []byte(key)
		inst.blockSize = BLOCK_SIZE
		inst.ivType = IV_TYPE_RAND
		_insts[ENCRYPT_TYPE_COMMON] = inst
	} else {
		_insts[ENCRYPT_TYPE_COMMON].key = []byte(key)
	}
	return _insts[ENCRYPT_TYPE_COMMON]
}

// NewPriv  , 隐私整改，和Server统一的加密算法
func NewPri(key string) *EncryptCBC {
	if _insts[ENCRYPT_TYPE_PRIVATE] == nil {
		inst := new(EncryptCBC)
		inst.key = []byte(key)
		inst.blockSize = BLOCK_SIZE
		inst.ivType = IV_TYPE_KEY
		_insts[ENCRYPT_TYPE_PRIVATE] = inst
	} else {
		_insts[ENCRYPT_TYPE_PRIVATE].key = []byte(key)
	}
	return _insts[ENCRYPT_TYPE_PRIVATE]
}

//Encrypt 加密方法
func (e *EncryptCBC) Encrypt(plaintext string) (ciphertext string, err error) {
	if len(plaintext) == 0 {
		return "", errors.New("plaintext is nil")
	}
	var cipherByte []byte
	if cipherByte, err = e.cbcEncypt([]byte(plaintext)); err != nil {
		return
	}
	return base64.StdEncoding.EncodeToString(cipherByte), nil
}

//Decrypt 解密方法
func (e *EncryptCBC) Decrypt(ciphertext string) (plaintext string, err error) {
	if len(ciphertext) == 0 {
		return "", errors.New("ciphertext is nil")
	}
	var cipherByte []byte
	if cipherByte, err = base64.StdEncoding.DecodeString(ciphertext); err != nil {
		return
	}
	var plainByte []byte
	if plainByte, err = e.cbcDecrypt(cipherByte); err != nil {
		return "", err
	}
	return string(plainByte), nil
}

//EncryptBytes 加密方法
func (e *EncryptCBC) EncryptBytes(plainBytes []byte) (cipherBytes []byte, err error) {
	if len(plainBytes) == 0 {
		return nil, errors.New("plainBytes is nil")
	}
	if cipherBytes, err = e.cbcEncypt(plainBytes); err != nil {
		return
	}
	return
}

//Decrypt 解密方法
func (e *EncryptCBC) DecryptBytes(cipherBytes []byte) (plainBytes []byte, err error) {
	if len(cipherBytes) == 0 {
		return nil, errors.New("cipherBytes is nil")
	}
	if plainBytes, err = e.cbcDecrypt(cipherBytes); err != nil {
		return nil, err
	}
	return
}

//私有方法，cbc加密
// CBC mode works on blocks so plaintexts may need to be padded to the
// next whole block. For an example of such padding, see
// https://tools.ietf.org/html/rfc5246#section-6.2.3.2. Here we'll
// assume that the plaintext is already of the correct length.
func (e *EncryptCBC) cbcEncypt(plainBytes []byte) (cipherByte []byte, err error) {
	block, err := aes.NewCipher(e.key)
	if err != nil {
		return
	}
	origData := e.PKCS5Padding(plainBytes, e.blockSize)
	// The IV needs to be unique, but not secure. Therefore it's common to
	// include it at the beginning of the ciphertext.
	var iv []byte
	if e.ivType == IV_TYPE_RAND {
		cipherByte = make([]byte, e.blockSize+len(origData))
		iv = cipherByte[:e.blockSize]
		if _, err = io.ReadFull(rand.Reader, iv); err != nil {
			return
		}
		mode := cipher.NewCBCEncrypter(block, iv)
		mode.CryptBlocks(cipherByte[e.blockSize:], origData)
	} else if e.ivType == IV_TYPE_KEY {
		cipherByte = make([]byte, len(origData))
		iv = e.key[:e.blockSize]
		mode := cipher.NewCBCEncrypter(block, iv)
		mode.CryptBlocks(cipherByte, origData)
	} else {
		return nil, fmt.Errorf("iv type[%d] invalid", e.ivType)
	}

	// It's important to remember that ciphertexts must be authenticated
	// (i.e. by using crypto/hmac) as well as being encrypted in order to
	// be secure.
	return
}

//私有方法，cbc解密
func (e *EncryptCBC) cbcDecrypt(cipherbytes []byte) (plainbyte []byte, err error) {
	block, err := aes.NewCipher(e.key)
	if err != nil {
		return
	}
	// The IV needs to be unique, but not secure. Therefore it's common to
	// include it at the beginning of the ciphertext.
	if len(cipherbytes) < e.blockSize {
		return nil, fmt.Errorf("ciphertext too short")
	}
	var iv []byte
	if e.ivType == IV_TYPE_RAND {
		iv = cipherbytes[:e.blockSize]
		cipherbytes = cipherbytes[e.blockSize:]
		// CBC mode always works in whole blocks.
	} else if e.ivType == IV_TYPE_KEY {
		iv = e.key[:e.blockSize]
	} else {
		return nil, fmt.Errorf("iv type[%d] invalid", e.ivType)
	}
	if len(cipherbytes)%e.blockSize != 0 {
		return nil, fmt.Errorf("cipherbytes is not a multiple of the block size")
	}
	mode := cipher.NewCBCDecrypter(block, iv)
	plainbyte = make([]byte, len(cipherbytes))
	// CryptBlocks can work in-place if the two arguments are the same.
	mode.CryptBlocks(plainbyte, cipherbytes)
	// If the original plaintext lengths are not a multiple of the block
	// size, padding would have to be added when encrypting, which would be
	// removed at this point. For an example, see
	// https://tools.ietf.org/html/rfc5246#section-6.2.3.2. However, it's
	// critical to note that ciphertexts must be authenticated (i.e. by
	// using crypto/hmac) before being decrypted in order to avoid creating
	// a padding oracle.
	plainbyte = e.PKCS5UnPadding(plainbyte)
	return
}

//PKCS5Padding ...
func (e *EncryptCBC) PKCS5Padding(ciphertext []byte, blockSize int) []byte {
	padding := blockSize - len(ciphertext)%blockSize                 //长度不是blockSize时补齐个数
	padtext := bytes.Repeat([]byte{byte(padding + OFFSET)}, padding) //至少补一个
	return append(ciphertext, padtext...)
}

//PKCS5UnPadding ...
func (e *EncryptCBC) PKCS5UnPadding(origData []byte) []byte {
	length := len(origData)
	// 去掉最后一个字节 unpadding 次
	unpadding := int(origData[length-1]) - OFFSET //最后一个字节是补充的，值是补充的个数，可以自定义规则
	return origData[:(length - unpadding)]
}
