package ssh

import (
	"context"
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/log"
	"github.com/charmbracelet/ssh"
	"github.com/charmbracelet/wish"
	"github.com/muesli/termenv"
)

// BubbleTeaHandler 是 Bubble Tea 应用程序实现的函数，用于挂钩到
// SSH 中间件。这将为每个连接创建一个新的 tea.Program 并
// 使用返回的 tea.ProgramOptions 启动它。
//
// 已弃用：请使用 Handler 代替。
type BubbleTeaHandler = Handler // nolint: revive

// Handler 是 Bubble Tea 应用程序实现的函数，用于挂钩到
// SSH 中间件。这将为每个连接创建一个新的 tea.Program 并
// 使用返回的 tea.ProgramOptions 启动它。
type Handler func(sess ssh.Session) (tea.Model, []tea.ProgramOption)

// ProgramHandler 是 Bubble Tea 应用程序实现的函数，用于挂钩到 SSH
// 中间件。这应该返回一个新的 tea.Program。此处理程序与
// 默认处理程序的不同之处在于它返回 tea.Program 而不是
// (tea.Model, tea.ProgramOptions)。
//
// 确保将 tea.WithInput 和 tea.WithOutput 设置为 ssh.Session
// 否则程序将无法正常工作。
type ProgramHandler func(sess ssh.Session) *tea.Program

// BubbleteaMiddleware 接受一个 Handler 并将 ssh.Session 的输入和输出
// 挂钩到 tea.Program 中。
//
// 它还捕获窗口调整大小事件并将它们作为 tea.WindowSizeMsgs
// 发送到 tea.Program。
func BubbleteaMiddleware(handler Handler) wish.Middleware {
	return MiddlewareWithProgramHandler(newDefaultProgramHandler(handler), termenv.TrueColor)
}

// MiddlewareWithColorProfile 允许您指定此程序正常工作所需的最少颜色数。
//
// 如果客户端的颜色配置文件的颜色少于 p，将强制使用 p。
// 请谨慎使用。
//func MiddlewareWithColorProfile(handler Handler, profile termenv.Profile) wish.Middleware {
//	return MiddlewareWithProgramHandler(newDefaultProgramHandler(handler), profile)
//}

// MiddlewareWithProgramHandler 允许您指定 ProgramHandler 以便
// 能够访问底层的 tea.Program，以及最低支持的颜色配置文件。
//
// 这对于创建需要访问 tea.Program 的自定义中间件很有用，
// 例如使用 p.Send() 向 tea.Program 发送消息。
//
// 确保将 tea.WithInput 和 tea.WithOutput 设置为 ssh.Session
// 否则程序将无法正常工作。推荐的做法是使用 MakeOptions。
//
// 如果客户端的颜色配置文件的颜色少于 p，将强制使用 p。
// 请谨慎使用。
func MiddlewareWithProgramHandler(handler ProgramHandler, profile termenv.Profile) wish.Middleware {
	return func(next ssh.Handler) ssh.Handler {
		return func(sess ssh.Session) {
			sess.Context().SetValue(minColorProfileKey, profile)
			program := handler(sess)
			if program == nil {
				next(sess)
				return
			}
			_, windowChanges, ok := sess.Pty()
			if !ok {
				wish.Fatalln(sess, "no active terminal, skipping")
				return
			}
			ctx, cancel := context.WithCancel(sess.Context())
			go func() {
				for {
					select {
					case <-ctx.Done():
						program.Quit()
						return
					case w := <-windowChanges:
						program.Send(tea.WindowSizeMsg{Width: w.Width, Height: w.Height})
					}
				}
			}()
			if _, err := program.Run(); err != nil {
				log.Error("app exit with error", "error", err)
			}
			// p.Kill() 将强制终止程序（如果它仍在运行），
			// 并在 TUI 崩溃的情况下将终端恢复到原始状态
			program.Kill()
			cancel()
			next(sess)
		}
	}
}

var minColorProfileKey struct{}

var profileNames = [4]string{"TrueColor", "ANSI256", "ANSI", "Ascii"}

// MakeRenderer 为当前会话返回一个 lipgloss 渲染器。
// 此函数也处理 PTY，应该用于为您的应用程序设置样式。
func MakeRenderer(sess ssh.Session) *lipgloss.Renderer {
	cp, ok := sess.Context().Value(minColorProfileKey).(termenv.Profile)
	if !ok {
		cp = termenv.Ascii
	}

	r := newRenderer(sess)

	// 只有在请求的会话是 PTY 时，我们才强制使用颜色配置文件。
	_, _, ok = sess.Pty()
	if !ok {
		return r
	}

	if r.ColorProfile() > cp {
		_, _ = fmt.Fprintf(sess.Stderr(), "Warning: Client's terminal is %q, forcing %q\r\n",
			profileNames[r.ColorProfile()], profileNames[cp])
		r.SetColorProfile(cp)
	}
	return r
}

// MakeOptions 返回 tea.WithInput 和 tea.WithOutput 程序选项，
// 考虑可能的模拟或分配的 PTY。
//func MakeOptions(sess ssh.Session) []tea.ProgramOption {
//	return makeOpts(sess)
//}

type sshEnviron []string

var _ termenv.Environ = sshEnviron(nil)

// Environ 实现 termenv.Environ 接口。
func (e sshEnviron) Environ() []string {
	return e
}

// Getenv 实现 termenv.Environ 接口。
func (e sshEnviron) Getenv(k string) string {
	for _, v := range e {
		if strings.HasPrefix(v, k+"=") {
			return v[len(k)+1:]
		}
	}
	return ""
}

func newDefaultProgramHandler(handler Handler) ProgramHandler {
	return func(s ssh.Session) *tea.Program {
		m, opts := handler(s)
		if m == nil {
			return nil
		}
		return tea.NewProgram(m, append(opts, makeOpts(s)...)...)
	}
}

func makeOpts(s ssh.Session) []tea.ProgramOption {
	return []tea.ProgramOption{
		tea.WithInput(s),
		tea.WithOutput(s),
	}
}

func newRenderer(s ssh.Session) *lipgloss.Renderer {
	pty, _, _ := s.Pty()
	env := sshEnviron(append(s.Environ(), "TERM="+pty.Term))
	return lipgloss.NewRenderer(
		s,
		termenv.WithEnvironment(env),
		termenv.WithUnsafe(),
		termenv.WithColorCache(true),
	)
}
