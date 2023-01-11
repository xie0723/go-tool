package cfb

import (
	"testing"
)

func Test_New(t *testing.T) {
	_new := New("123.456789.abcxx", "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789+/")
	if _new.aesKey != "123.456789.abcxx" || _new.base64Table != "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789+/" {
		t.Error("New 一个实例测试未通过")
	}
}
func Test_Encrypt(t *testing.T) {
	_new := New("123.456789.abcxx", "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789+/")
	if cipherstr, _ := _new.Encrypt("HelloWord!1503113870"); cipherstr != "sCYrFAthKfXHannv53f4PRFSAH4=" {
		t.Error("加密方法测试失败")
	}
}

func Test_Decrypt(t *testing.T) {
	_new := New("123.456789.abcxx", "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789+/")
	if plainstr, _ := _new.Decrypt("sCYrFAthKfXHannv53f4PRFSAH4="); plainstr != "HelloWord!1503113870" {
		t.Error("解密方法测试失败")
	}
}
