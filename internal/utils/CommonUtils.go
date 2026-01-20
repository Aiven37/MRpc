package utils

import (
	"math/rand"
	"strconv"
	"time"
)

var (
	charset    = []byte("0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")
	randSource = rand.New(rand.NewSource(time.Now().UnixNano()))
)

func GenerateSessionId(length int) string {
	timestamp := strconv.FormatInt(time.Now().UnixMilli(), 10)
	current := []byte(timestamp)

	if len(current) >= length {
		return string(current[:length])
	}

	// 每次插入一个字符到随机位置
	for len(current) < length {
		char := charset[randSource.Intn(len(charset))]
		pos := randSource.Intn(len(current) + 1) // +1 允许插入末尾
		current = append(current[:pos], append([]byte{char}, current[pos:]...)...)
	}
	return string(current)
}
