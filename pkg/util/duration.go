package util

import "time"

// SecondsDuration 将配置中的秒值转换为 time.Duration，保留纳秒级精度。
func SecondsDuration(seconds float64) time.Duration {
	return time.Duration(seconds * float64(time.Second))
}
