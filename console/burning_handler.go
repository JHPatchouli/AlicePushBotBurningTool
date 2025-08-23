package console

import (
	"AlicePushBotBurningTool/terminal"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

// BurningHandler 处理烧录相关的操作
type BurningHandler struct {
	App      *tview.Application
	Pages    *tview.Pages
	Terminal *terminal.Terminal
}

// NewBurningHandler 创建一个新的烧录处理器
func NewBurningHandler(app *tview.Application, pages *tview.Pages, advancedMode bool) *BurningHandler {
	h := &BurningHandler{
		App:      app,
		Pages:    pages,
		Terminal: terminal.NewTerminal(app, advancedMode),
	}
	return h
}

// HandleBurningManagement 处理特定的烧录管理
func (h *BurningHandler) HandleBurningManagement() {

	// 创建一个新的Flex布局作为主容器
	mainFlex := tview.NewFlex().SetDirection(tview.FlexRow)

	// 添加标题
	title := tview.NewTextView().SetText("简易模式").SetTextAlign(tview.AlignCenter).SetTextColor(tview.Styles.TitleColor)
	mainFlex.AddItem(title, 1, 1, false)

	// 创建下方的水平分割Flex
	bottomFlex := tview.NewFlex().SetDirection(tview.FlexColumn)

	terminalFlex := h.Terminal.GetFlex()

	// 创建左侧选项列表 - 使用标准快捷键
	optionsList := tview.NewList().ShowSecondaryText(false)

	// 添加选项 - 使用标准快捷键
	optionsList.AddItem("执行烧录操作", "执行烧录命令", '1', func() {
		// 定义烧录命令
		cmd := "echo '执行烧录操作'"
		h.Terminal.ExecuteCommand(cmd)
	})

	optionsList.AddItem("查看烧录日志", "查看烧录过程中的日志", '2', func() {
		// 定义查看日志命令
		cmd := "echo '查看烧录日志'"
		h.Terminal.ExecuteCommand(cmd)
	})

	optionsList.AddItem("配置烧录参数", "配置烧录相关的参数", '3', func() {
		// 定义配置参数命令
		cmd := "echo '配置烧录参数'"
		h.Terminal.ExecuteCommand(cmd)
	})
	optionsList.AddItem("退出", "", 'q', func() {
		h.App.QueueUpdateDraw(func() {
			h.Pages.SwitchToPage("main_menu")
		})
	})

	// 为选项列表添加Tab键处理，方便在选项列表和终端之间切换焦点
	optionsList.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		// 处理Tab键切换焦点
		if event.Key() == tcell.KeyTab {
			h.App.SetFocus(h.Terminal.GetInputField())
			return nil
		}
		return event
	})

	// 为终端输入框添加Shift+Tab键处理，方便从终端返回选项列表
	h.Terminal.GetInputField().SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		// 处理Shift+Tab键切换焦点
		if event.Key() == tcell.KeyBacktab || event.Key() == tcell.KeyTab {
			h.App.SetFocus(optionsList)
			return nil
		}
		return event
	})

	// 将左侧选项列表和右侧终端添加到底部Flex中
	bottomFlex.AddItem(optionsList, 0, 1, true)
	bottomFlex.AddItem(terminalFlex, 0, 2, false)

	// 将底部Flex添加到主Flex中
	mainFlex.AddItem(bottomFlex, 0, 1, true)

	// 设置主Flex为应用程序的根元素，并将焦点设置在终端上
	h.Pages.AddPage("burning_management", mainFlex, true, true)
	h.App.SetFocus(optionsList)

}
