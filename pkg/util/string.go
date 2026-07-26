package util

import (
	"crypto/rand"
	"errors"
	"math/big"
)

const digitsAndLowercase = "1234567890abcdefghijklmnopqrstuvwxyz"

// String 返回由数字和小写字母组成的加密随机字符串。
func String(length int) (string, error) {
	if length <= 0 {
		return "", errors.New("length must be positive")
	}
	buf := make([]byte, length)
	maxBound := big.NewInt(int64(len(digitsAndLowercase)))
	for i := range buf {
		// rand.Int 使用拒绝采样，避免直接取模造成字符分布偏差。
		n, err := rand.Int(rand.Reader, maxBound)
		if err != nil {
			return "", err
		}
		buf[i] = digitsAndLowercase[n.Int64()]
	}
	return string(buf), nil
}
