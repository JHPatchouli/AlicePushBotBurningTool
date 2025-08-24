package terminal

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

// Terminal represents a reusable terminal component.
type Terminal struct {
	App          *tview.Application
	Output       *tview.TextView
	Input        *tview.InputField
	Flex         *tview.Flex
	lastPgDnTime time.Time
	autoScroll   bool
	currentCmd   *exec.Cmd
	cmdCancelCh  chan struct{}
}

// Write writes the given bytes to the terminal output.
func (t *Terminal) Write(p []byte) (n int, err error) {
	return t.Output.Write(p)
}

// GetFlex returns the Flex layout of the terminal.
func (t *Terminal) GetFlex() *tview.Flex {
	return t.Flex
}

// GetInputField returns the InputField of the terminal.
func (t *Terminal) GetInputField() *tview.InputField {
	return t.Input
}

// NewTerminal creates a new Terminal instance.
func NewTerminal(app *tview.Application, advancedMode bool) *Terminal {
	output := tview.NewTextView().SetDynamicColors(true)
	if advancedMode {
		output.SetText("[yellow]Tab键切换到终端可直接输入命令回车执行...[white]\n[green]终端内Tab可以脱离终端使用左侧菜单[white]\n[red]Alt+C终止命令[white]\n")
	} else {
		output.SetText("[yellow]请等待命令执行...[white]\n")
	}
	output.SetScrollable(true).SetWordWrap(true)

	input := tview.NewInputField().
		SetFieldStyle(tcell.StyleDefault.Background(tcell.ColorBlack).Foreground(tcell.ColorWhite)).
		SetFieldBackgroundColor(tcell.ColorBlack).
		SetFieldTextColor(tcell.ColorWhite).
		SetLabel("> ").
		SetLabelColor(tcell.ColorWhite).
		SetText("")

	flex := tview.NewFlex().SetDirection(tview.FlexRow)
	flex.AddItem(output, 0, 3, false)
	if advancedMode {
		flex.AddItem(input, 1, 1, true)
	}
	flex.SetBorder(true).SetTitle("终端").SetTitleAlign(tview.AlignLeft)

	t := &Terminal{
		App:        app,
		Output:     output,
		Input:      input,
		Flex:       flex,
		autoScroll: true,
	}

	// Add keyboard event handling for output area
	t.Output.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyPgUp:
			t.autoScroll = false
			t.Output.ScrollTo(-5, 0)
			return nil
		case tcell.KeyPgDn:
			now := time.Now()
			if now.Sub(t.lastPgDnTime) < 500*time.Millisecond {
				t.autoScroll = true
				t.Output.ScrollToEnd()
			} else {
				t.Output.ScrollTo(5, 0)
			}
			t.lastPgDnTime = now
			return nil
		}
		return event
	})

	// Add input handling for the input field
	t.Input.SetDoneFunc(func(key tcell.Key) {
		if key == tcell.KeyEnter {
			cmdStr := t.Input.GetText()
			t.ExecuteCommand(cmdStr)
		}
	})

	return t
}

// ExecuteCommand executes a command in the terminal.
func (t *Terminal) ExecuteCommand(cmdStr string) chan string {
	resultChan := make(chan string)

	if cmdStr == "" {
		close(resultChan)
		return resultChan
	}

	cmdStrLower := strings.ToLower(cmdStr)
	if strings.Contains(cmdStrLower, "cls") || strings.Contains(cmdStrLower, "clear") || strings.Contains(cmdStrLower, "clear-host") {
		t.App.QueueUpdateDraw(func() {
			t.Output.SetText("")
			t.Input.SetText("")
		})
		close(resultChan)
		return resultChan
	}

	t.Input.SetText("")
	t.Output.Write([]byte("$ " + cmdStr + "\n"))

	t.cmdCancelCh = make(chan struct{})

	go func() {
		defer close(resultChan)
		var innerWg sync.WaitGroup

		var shell, flag string
		if runtime.GOOS == "windows" {
			shell = "powershell"
			flag = "-Command"
			cmdStr = "chcp 65001 | Out-Null; " + cmdStr
		} else {
			shell = "bash"
			flag = "-c"
		}

		t.currentCmd = exec.Command(shell, flag, cmdStr)

		var stdoutBuf, stderrBuf strings.Builder

		stdoutPipe, err := t.currentCmd.StdoutPipe()
		if err != nil {
			t.App.QueueUpdateDraw(func() {
				t.Output.Write([]byte("[red]创建StdoutPipe失败: [white]" + err.Error() + "\n"))
				if t.autoScroll {
						t.Output.ScrollToEnd()
					}
				})
			resultChan <- "Error: " + err.Error()
			return
		}
		stderrPipe, err := t.currentCmd.StderrPipe()
		if err != nil {
			t.App.QueueUpdateDraw(func() {
				t.Output.Write([]byte("[red]创建StderrPipe失败: [white]" + err.Error() + "\n"))
				if t.autoScroll {
						t.Output.ScrollToEnd()
					}
				})
			resultChan <- "Error: " + err.Error()
			return
		}
		innerWg.Add(3)

		go func() {
			defer innerWg.Done()
			scanner := bufio.NewScanner(stdoutPipe)
			for scanner.Scan() {
				line := scanner.Bytes()
				stdoutBuf.Write(line)
				stdoutBuf.WriteByte('\n')
				t.App.QueueUpdateDraw(func() {
					t.Output.Write(append(line, '\n'))
					t.Output.ScrollToEnd()
				})
			}
			if err := scanner.Err(); err != nil {
				if errors.Is(err, io.EOF) ||
					errors.Is(err, os.ErrClosed) ||
					errors.Is(err, syscall.EBADF) {
					if data := scanner.Bytes(); len(data) > 0 {
						stdoutBuf.Write(data)
						stdoutBuf.WriteByte('\n')
						t.App.QueueUpdateDraw(func() {
							t.Output.Write(append(data, '\n'))
							t.Output.ScrollToEnd()
						})
					}
				} else {
					t.App.QueueUpdateDraw(func() {
						t.Output.Write([]byte("[red]读取标准输出失败: [white]" + err.Error() + "\n"))
						t.Output.ScrollToEnd()
					})
				}
			}
		}()
		go func() {
			defer innerWg.Done()
			scanner := bufio.NewScanner(stderrPipe)
			for scanner.Scan() {
				line := scanner.Bytes()
				stderrBuf.Write(line)
				stderrBuf.WriteByte('\n')
				t.App.QueueUpdateDraw(func() {
					t.Output.Write(append(line, '\n'))
					t.Output.ScrollToEnd()
				})
			}
			if err := scanner.Err(); err != nil {
				if errors.Is(err, io.EOF) ||
					errors.Is(err, os.ErrClosed) ||
					errors.Is(err, syscall.EBADF) {
					if data := scanner.Bytes(); len(data) > 0 {
						stderrBuf.Write(data)
						stderrBuf.WriteByte('\n')
						t.App.QueueUpdateDraw(func() {
							t.Output.Write(append(data, '\n'))
							t.Output.ScrollToEnd()
						})
					}
				} else {
					t.App.QueueUpdateDraw(func() {
						t.Output.Write([]byte("[red]读取标准错误失败: [white]" + err.Error() + "\n"))
						t.Output.ScrollToEnd()
					})
				}
			}
		}()

		if err := t.currentCmd.Start(); err != nil {
			t.App.QueueUpdateDraw(func() {
				t.Output.Write([]byte("[red]命令启动失败: [white]" + err.Error() + "\n"))
				if t.autoScroll {
					t.Output.ScrollToEnd()
				}
			})
			resultChan <- "Error: " + err.Error()
			return
		}

		go func() {
			defer innerWg.Done()
			done := make(chan error, 1)
			go func() {
				done <- t.currentCmd.Wait()
			}()

			select {
			case <-t.cmdCancelCh:
				if t.currentCmd != nil && t.currentCmd.Process != nil {
					if runtime.GOOS == "windows" {
						// Windows下使用Terminate，因为Kill对于子进程无效
						_ = t.currentCmd.Process.Kill()
					} else {
						// Unix-like系统发送SIGTERM
						_ = t.currentCmd.Process.Signal(syscall.SIGTERM)
					}
					t.App.QueueUpdateDraw(func() {
						t.Output.Write([]byte("[red]命令已终止。[white]\n"))
						if t.autoScroll {
							t.Output.ScrollToEnd()
						}
					})
				}
			case <-time.After(1 * time.Second):
				// 等待一秒，如果命令还在运行，则继续等待
				break
			}

			if err := <-done; err != nil {
				t.App.QueueUpdateDraw(func() {
					if exitError, ok := err.(*exec.ExitError); ok {
						t.Output.Write([]byte(fmt.Sprintf("[red]命令执行完成，但出现错误: %s (Exit Code: %d)[white]\n", exitError.Error(), exitError.ExitCode())))
					} else {
						t.Output.Write([]byte("[red]命令执行失败: [white]" + err.Error() + "\n"))
					}
					if t.autoScroll {
						t.Output.ScrollToEnd()
					}
				})
				resultChan <- "Error: " + err.Error()
			} else {
				t.App.QueueUpdateDraw(func() {
					t.Output.Write([]byte("[green]命令执行完成。[white]\n"))
					if t.autoScroll {
						t.Output.ScrollToEnd()
					}
				})
				resultChan <- stdoutBuf.String() + stderrBuf.String()
			}
			t.currentCmd = nil
		}()
		innerWg.Wait()
	}()
	return resultChan
}
// ClearOutput clears the terminal output.
func (t *Terminal) ClearOutput() {
	t.Output.SetText("")
}

// WriteOutput writes text to the terminal output.
func (t *Terminal) WriteOutput(text string) {

	t.Output.Write([]byte(text))
	if t.autoScroll {
		t.Output.ScrollToEnd()
	}

}

// SetInputText sets the text of the input field.
func (t *Terminal) SetInputText(text string) {
	t.Input.SetText(text)
}

// FocusInput sets focus to the input field.
func (t *Terminal) FocusInput() {
	t.App.SetFocus(t.Input)
}
