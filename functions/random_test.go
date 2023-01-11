package functions

import (
	"fmt"
	"testing"
)

func TestCreateRandomNumber(t *testing.T) {
	a := CreateRandomNumber(3)
	fmt.Println("CreateRandomNumbera:", a)
	if len(a) != 3 {
		t.Error("not 3")
	}
	b := CreateRandomNumber(3)
	if a == b {
		t.Error("not random")
	}
	fmt.Println("CreateRandomNumberb:", b)
}

func TestCreateRandomString(t *testing.T) {
	a := CreateRandomString(8)
	fmt.Println("CreateRandomStringa:", a)
	if len(a) != 8 {
		t.Error("not 8")
	}
	b := CreateRandomString(8)
	if a == b {
		t.Error("not random")
	}
	fmt.Println("CreateRandomStringb:", b)
}
