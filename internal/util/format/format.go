package util

import (
	"fmt"
	"strings"
	"time"

	model "github.com/lin-snow/ech0/internal/model/metric"
	"github.com/robfig/cron/v3"
)

// 格式化指标（四舍五入 + 单位转换）
func FormatMetrics(m *model.Metrics) *model.Metrics {
	formatted := *m // 复制原数据，避免修改原始指标

	// CPU 精度控制
	formatted.CPU.UsagePercent = round(m.CPU.UsagePercent, 2)

	// 内存转 GB + 百分比精度控制
	formatted.Memory.Total = bytesToGB(m.Memory.Total)
	formatted.Memory.Used = bytesToGB(m.Memory.Used)
	formatted.Memory.Available = bytesToGB(m.Memory.Available)
	formatted.Memory.Percentage = round(m.Memory.Percentage, 2)

	// 磁盘转 GB + 百分比精度控制
	formatted.Disk.Total = bytesToGB(m.Disk.Total)
	formatted.Disk.Used = bytesToGB(m.Disk.Used)
	formatted.Disk.Available = bytesToGB(m.Disk.Available)
	formatted.Disk.Percentage = round(m.Disk.Percentage, 2)

	// 网络速率转 MB/s，总量转 MB
	formatted.Network.TotalBytesSent = bytesToMB(m.Network.TotalBytesSent)
	formatted.Network.TotalBytesReceived = bytesToMB(m.Network.TotalBytesReceived)
	formatted.Network.BytesSentPerSecond = round(m.Network.BytesSentPerSecond/1024/1024, 2)
	formatted.Network.BytesReceivedPerSecond = round(m.Network.BytesReceivedPerSecond/1024/1024, 2)

	// 系统指标：Uptime 保留小时数（float）
	formatted.System.Uptime = hoursDuration(m.System.Uptime)

	return &formatted
}

func ValidateCrontabExpression(expr string) error {
	// 去掉多余空格，防止 "  * * * * *  " 之类情况
	expr = strings.TrimSpace(expr)
	fields := strings.Fields(expr)

	switch len(fields) {
	case 5:
		// 标准 cron（分钟、小时、日、月、星期）
		_, err := cron.ParseStandard(expr)
		if err != nil {
			return fmt.Errorf("invalid 5-field cron expression: %w", err)
		}
	case 6:
		// 含秒字段（秒、分钟、小时、日、月、星期）
		parser := cron.NewParser(
			cron.Second | cron.Minute | cron.Hour | cron.Dom | cron.Month | cron.Dow,
		)
		_, err := parser.Parse(expr)
		if err != nil {
			return fmt.Errorf("invalid 6-field cron expression: %w", err)
		}
	default:
		return fmt.Errorf("cron expression must have 5 or 6 fields, got %d", len(fields))
	}

	return nil
}

// ====================== 工具函数 ======================

// round 保留 n 位小数
func round(f float64, n int) float64 {
	format := fmt.Sprintf("%%.%df", n)
	_, _ = fmt.Sscanf(fmt.Sprintf(format, f), "%f", &f)
	return f
}

// bytesToGB 将字节转为 GB
func bytesToGB(b uint64) uint64 {
	return uint64(float64(b) / (1024 * 1024 * 1024))
}

// bytesToMB 将字节转为 MB
func bytesToMB(b uint64) uint64 {
	return uint64(float64(b) / (1024 * 1024))
}

// hoursDuration 将 Duration 转为小时级（近似）
func hoursDuration(d time.Duration) time.Duration {
	return d / (3600 * 1e9)
}
