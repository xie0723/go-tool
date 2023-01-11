package stringmatch

import (
	"strings"
	"testing"
)

func TestFalse(t *testing.T) {
	var str = "false OR (true AND (false AND ff))"
	if ret, err := Calculate(str, 20, func(s string) bool {
		return s != "false"
	}); err != nil {
		t.Error(err.Error())
	} else {
		if ret {
			t.Error("result shoud be false")
		}
	}
}

func TestTrue(t *testing.T) {
	var str = "false OR (true AND (x AND ff))"
	if ret, err := Calculate(str, 20, func(s string) bool {
		return strings.Contains(s, "false")
	}); err != nil {
		t.Error(err.Error())
	} else {
		if !ret {
			t.Error("result shoud be true")
		}
	}
}

func TestInvalid(t *testing.T) {
	var str = "false OR (true AND ((x AND ff))"
	if _, err := Calculate(str, 1, func(s string) bool {
		return strings.Contains(s, "false")
	}); err == nil {
		t.Error(err.Error())
	}
}
