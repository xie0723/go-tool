package functions

import (
	"math/rand"
	"strconv"
	"time"
)

// 生成指定位数的随机数字
func CreateRandomNumber(count int) string {
	var numbers = []int{1, 2, 3, 4, 5, 7, 8, 9}
	var container string
	length := len(numbers)
	for i := 1; i <= count; i++ {
		rand.Seed(time.Now().UnixNano())
		random := rand.Intn(length)
		container += strconv.Itoa(numbers[random])
	}
	return container
}

// 生成指定长度的随机字符串
func CreateRandomString(count int) string {
	var letters = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ1234567890")
	length := len(letters)
	b := make([]rune, count)
	for i := 0; i < count; i++ {
		rand.Seed(time.Now().UnixNano())
		randomInt := rand.Intn(length)
		b[i] = letters[randomInt]
	}
	return string(b)
}
