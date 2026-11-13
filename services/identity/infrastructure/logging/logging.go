package logging

import (
	"log/slog"
	"os"
)

// Init 初始化 JSON 结构化分级日志（宪法 v2.8.0 D 组）。
// 固定字段：service、version；级别 debug/info/warn/error，debug 默认关闭。
func Init(service, version string) {
	level := slog.LevelInfo
	if os.Getenv("IDENTITY_LOG_LEVEL") == "debug" {
		level = slog.LevelDebug
	}
	h := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: level})
	logger := slog.New(h).With("service", service, "version", version)
	slog.SetDefault(logger)
}
