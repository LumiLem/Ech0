package tui

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"

	"github.com/charmbracelet/lipgloss"

	commonModel "github.com/lin-snow/ech0/internal/model/common"
)

var (
	// 信息样式（每行）
	infoStyle = lipgloss.NewStyle().
			PaddingLeft(2).
			Foreground(lipgloss.AdaptiveColor{
			Light: "236", Dark: "252",
		})

	// 标题样式
	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.AdaptiveColor{
			Light: "#4338ca", Dark: "#FF7F7F",
		})

	// 高亮样式
	highlight = lipgloss.NewStyle().
			Bold(false).
			Italic(true).
			Foreground(lipgloss.AdaptiveColor{
			Light: "#7c3aed", Dark: "#53b7f5ff",
		})

	// 外框
	boxStyle = lipgloss.NewStyle().
			Bold(true).
			BorderStyle(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#fb5151ff")).
			Padding(1, 1).
			Margin(1, 1)
)

const (
	banner = `
    ______     __    ____ 
   / ____/____/ /_  / __ \
  / __/ / ___/ __ \/ / / /
 / /___/ /__/ / / / /_/ / 
/_____/\___/_/ /_/\____/  
`
)

// GetLogoBanner 获取Logo横幅
func GetLogoBanner() string {
	lines := strings.Split(banner, "\n")
	var rendered []string

	colors := []string{
		"#FF7F7F", // 珊瑚红
		"#FFB347", // 桃橙色
		"#FFEB9C", // 金黄色
		"#B8E6B8", // 薄荷绿
		"#87CEEB", // 天空蓝
		"#DDA0DD", // 梅花紫
		"#F0E68C", // 卡其色
	}

	for i, line := range lines {
		color := lipgloss.Color(colors[i%len(colors)])
		style := lipgloss.NewStyle().Foreground(color)
		rendered = append(rendered, style.Render(line))
	}
	gradientBanner := lipgloss.JoinVertical(lipgloss.Left, rendered...)

	full := lipgloss.JoinVertical(lipgloss.Left,
		gradientBanner,
	)

	return full
}

// PrintCLIBanner 打印CLI横幅
func PrintCLIBanner() {
	banner := GetLogoBanner()

	if _, err := fmt.Fprintln(os.Stdout, banner); err != nil {
		fmt.Fprintf(os.Stderr, "failed to print banner: %v\n", err)
	}
}

// PrintCLIInfo 打印CLI信息
func PrintCLIInfo(title, msg string) {
	// 使用 lipgloss 渲染 CLI 信息
	if _, err := fmt.Fprintln(os.Stdout, infoStyle.Render(titleStyle.Render(title)+": "+highlight.Render(msg))); err != nil {
		fmt.Fprintf(os.Stderr, "failed to print cli info: %v\n", err)
	}
}

// CLIInfoItem 定义了一个CLI信息项，包含标题和消息
type CLIInfoItem struct {
	Title string
	Msg   string
}

// GetCLIPrintWithBox 获取带边框的CLI信息打印内容
func GetCLIPrintWithBox(items ...CLIInfoItem) string {
	if len(items) == 0 {
		return ""
	}

	var content string
	for i, item := range items {
		line := infoStyle.Render(titleStyle.Render(item.Title) + ": " + highlight.Render(item.Msg))
		if i > 0 {
			content += "\n"
		}
		content += line
	}

	boxedContent := boxStyle.Render(content)
	return boxedContent
}

// PrintCLIWithBox 打印带边框的CLI信息
func PrintCLIWithBox(items ...CLIInfoItem) {
	if _, err := fmt.Fprintln(os.Stdout, GetCLIPrintWithBox(items...)); err != nil {
		fmt.Fprintf(os.Stderr, "failed to print cli box: %v\n", err)
	}
}

// ClearScreen 清屏函数，根据操作系统执行不同的清屏命令
func ClearScreen() {
	var cmd *exec.Cmd
	if runtime.GOOS == "windows" {
		cmd = exec.Command("cmd", "/c", "cls") // Windows 清屏命令
	} else {
		cmd = exec.Command("clear") // Linux/macOS 清屏命令
	}
	cmd.Stdout = os.Stdout
	if err := cmd.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "failed to clear screen: %v\n", err)
	}
}

// GetEch0Info 获取Ech0信息
func GetEch0Info() string {
	content := lipgloss.JoinVertical(lipgloss.Left,
		infoStyle.Render("📦 "+titleStyle.Render("Version")+": "+highlight.Render(commonModel.Version)),
		infoStyle.Render("🧙 "+titleStyle.Render("Author")+": "+highlight.Render("L1nSn0w")),
		infoStyle.Render("👉 "+titleStyle.Render("Website")+": "+highlight.Render("https://ech0.app/")),
		infoStyle.Render("👉 "+titleStyle.Render("GitHub")+": "+highlight.Render("https://github.com/lin-snow/Ech0")),
	)

	full := lipgloss.JoinVertical(lipgloss.Left,
		boxStyle.Render(content),
	)

	return full
}

// GetSSHView 获取SSH会话的视图
func GetSSHView() string {
	// header是一个长方形横向方框，内部是欢迎标题
	header := lipgloss.NewStyle().
		Width(80).
		Border(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color("#FF6B6B")).
		Render(
			lipgloss.JoinHorizontal(lipgloss.Center,
				"👋 Welcome to Ech0 SSH Session!",
			),
		)

	// body是一个长方形横向方框，内部分为左右两部分，左边是Logo,右边是简介
	body := lipgloss.NewStyle().
		Render(
			lipgloss.JoinHorizontal(lipgloss.Center,
				lipgloss.NewStyle().
					Width(40).
					Height(8).
					Render(GetLogoBanner()), // 使用logo
				lipgloss.NewStyle().
					Width(40).
					Height(8).
					Border(lipgloss.NormalBorder()).
					BorderForeground(lipgloss.Color("#dbe8f4ff")).
					Render(
						"Ech0 is a lightweight, self-hosted platform designed for quick sharing of your ideas, texts, and links.",
					),
			),
		)

	// footer是一个长方形横向方框，内部是退出提示
	footer := lipgloss.NewStyle().
		Width(80).
		Border(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color("#FF6B6B")).
		Render(
			lipgloss.JoinHorizontal(lipgloss.Center,
				"🧙 Press 'Ctrl+C' to exit the session.",
			),
		)

	// 将header, body, footer垂直连接起来
	full := lipgloss.NewStyle().
		Render(
			lipgloss.JoinVertical(lipgloss.Left,
				header,
				body,
				footer,
			),
		)

	return full
}
