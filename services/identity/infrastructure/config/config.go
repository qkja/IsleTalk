package config

import (
	"os"
	"strconv"
	"time"
)

// Duration 读取时长配置（如 "5m"、"168h"），缺失或非法回退默认值。
// 可调参数 MUST 经环境变量注入，MUST NOT 硬编码（plan 待办 1）。
func Duration(key string, def time.Duration) time.Duration {
	v := os.Getenv(key)
	if v == "" {
		return def
	}
	d, err := time.ParseDuration(v)
	if err != nil {
		return def
	}
	return d
}

// Days 读取天数配置，缺失或非法回退默认值。
func Days(key string, def int) int {
	v := os.Getenv(key)
	if v == "" {
		return def
	}
	n, err := strconv.Atoi(v)
	if err != nil || n <= 0 {
		return def
	}
	return n
}
