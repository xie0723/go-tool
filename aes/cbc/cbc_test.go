package cbc

import (
	"testing"
)

func Test_Encrypt(t *testing.T) {
	_new := New("1234567890123456")
	cipherstr, _ := _new.Encrypt("HelloWord!1503113870")
	if plainstr, _ := _new.Decrypt(cipherstr); plainstr != "HelloWord!1503113870" {
		t.Error("加密方法测试失败")
	}
}

func Test_Decrypt(t *testing.T) {
	_new := New("1234567890123456")
	if plainstr, _ := _new.Decrypt("Hy36YI++4lYMGq6ETSPExCI2zQv391dsq+KkNKYnuGpfxA+a1jE8vOd7lJBi4Jqa"); plainstr != "HelloWord!9876543210" {
		t.Error("解密方法测试失败")
	}
}
