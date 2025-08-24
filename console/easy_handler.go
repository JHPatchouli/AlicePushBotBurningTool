package console

import (
	"fmt"
	"runtime"
	"strings"
	"sync"

	"AlicePushBotBurningTool/terminal"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

// EasyHandler 简易模式处理
type EasyHandler struct {
	app         *tview.Application
	mainMenu    tview.Primitive
	title       *tview.TextView
	box         *tview.Flex
	list        *tview.List // 菜单列表
	terminal    *terminal.Terminal
	footer      *tview.TextView // 页脚提示
	deviceIndex int
}

func CreateEasyHandler(app *tview.Application, mainMenu tview.Primitive) *EasyHandler {
	return &EasyHandler{
		app:      app,
		mainMenu: mainMenu,
	}
}

func (e *EasyHandler) InitUI() tview.Primitive {
	flex := tview.NewFlex().SetDirection(tview.FlexRow)

	e.initTitle(`[::b]AlicePushBotBurningTool | 爱丽丝推送机器人烧录工具[::-]`)
	e.initBox()
	e.initFooter()
	flex.
		AddItem(e.title, 2, 1, false).
		AddItem(e.box, 0, 1, true).
		AddItem(e.footer, 1, 1, false) // 页脚提示（不可聚焦）
	return flex
}

var (
	deviceListOnce  sync.Once
	deviceListMutex sync.Mutex
)

func (e *EasyHandler) InitDeviceList() {
	deviceListOnce.Do(func() {
		e.list.Clear()
		e.list.SetTitle("设备列表")
		switch runtime.GOOS {
		case "windows":
			e.terminal.WriteOutput("开始查询设备\n")
			resultChan := e.terminal.ExecuteCommand(".\\adb\\windows\\adb devices")
			go func() {
				result := <-resultChan
				deviceListMutex.Lock()
				defer deviceListMutex.Unlock()
				resultLines := strings.Split(result, "\n")
				if len(resultLines) < 2 {
					e.app.QueueUpdateDraw(func() {
						e.terminal.WriteOutput(fmt.Sprintf("%d未查询到设备\n", len(resultLines)))
						e.list.AddItem("未查询到设备", "", '0', nil)
						e.list.AddItem("再次查询", "", 'r', func() {
							deviceListOnce = sync.Once{}
							e.InitDeviceList()
						})
						e.list.AddItem("返回主菜单", "", 'q', func() {
							e.app.SetRoot(e.mainMenu, true).SetFocus(e.mainMenu)
						})
					})
					return
				}
				for i, line := range resultLines {
					if strings.Contains(line, "device") && !strings.Contains(line, "List of devices attached") {
						deviceSN := strings.Split(line, "\t")[0]
						e.app.QueueUpdateDraw(func() {
							e.list.AddItem(fmt.Sprintf("设备 %d: %s", i, deviceSN), "", rune(fmt.Sprintf("%d", i)[0]), func() {
								e.terminal.WriteOutput(fmt.Sprintf("已选择设备: %s\n", deviceSN))
								e.deviceIndex = i
							})
						})
					}
				}
				e.app.QueueUpdateDraw(func() {
					e.list.AddItem("返回主菜单", "", 'q', func() {
						e.app.SetRoot(e.mainMenu, true).SetFocus(e.mainMenu)
					})
				})
			}()
		case "darwin":
			e.terminal.WriteOutput("This is a macOS system.\n")
		case "linux":
			e.terminal.WriteOutput("This is a Linux system.\n")
		default:
			e.terminal.WriteOutput(fmt.Sprintf("Unknown OS: %s\n", runtime.GOOS))
		}
	})
}



// 设置标题
func (e *EasyHandler) initTitle(title string) {
	e.title = tview.NewTextView().
		SetDynamicColors(true).
		SetTextAlign(tview.AlignCenter)
	// 设置工具名称
	titleText := title
	e.title.SetText(titleText)
}

func (e *EasyHandler) initBox() {
	e.box = tview.NewFlex().SetDirection(tview.FlexColumn)
	e.initList()
	e.initTerminal()
	e.box.AddItem(e.list, 0, 1, true).
		AddItem(e.terminal.GetFlex(), 0, 2, false)
}

func (e *EasyHandler) initList() {
	e.list = tview.NewList().ShowSecondaryText(false)
	e.list.SetBorder(true).SetTitle("操作步骤").SetTitleAlign(tview.AlignLeft)

	e.list.AddItem("查询设备", "", 'c', func() {
		deviceListOnce = sync.Once{}
		e.InitDeviceList()
	})
	e.list.AddItem("返回主菜单", "", 'q', func() {
		e.app.SetRoot(e.mainMenu, true).SetFocus(e.mainMenu)
	})
}

func (e *EasyHandler) initTerminal() {
	e.terminal = terminal.NewTerminal(e.app, false)
}

func (e *EasyHandler) initFooter() {
	e.footer = tview.NewTextView().
		SetDynamicColors(true).
		SetTextAlign(tview.AlignCenter).
		SetText("[gray]操作提示: ↑/↓ 移动 | 步骤左侧提示-快捷键 | Enter 选择")
	// 设置为不可聚焦
	e.footer.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		return event // 阻止所有输入事件
	})
}
