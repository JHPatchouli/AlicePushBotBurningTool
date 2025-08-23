package console

import (
	"bufio"
	"fmt"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"AlicePushBotBurningTool/terminal"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

// EasyHandler 处理简易模式的操作
type EasyHandler struct {
	App            *tview.Application
	Pages          *tview.Pages
	Terminal       *terminal.Terminal
	currentStep    int      // 当前步骤 (0-3)
	devices        []string // 检测到的设备列表
	selectedDevice string   // 用户选择的设备
	phoneNumber    string   // 设置的号码
	webhookURL     string   // 设置的Webhook URL
	initialized    bool     // 是否已初始化
}

// NewEasyHandler 创建一个新的简易模式处理器
func NewEasyHandler(app *tview.Application, pages *tview.Pages) *EasyHandler {
	h := &EasyHandler{
		App:         app,
		Pages:       pages,
		Terminal:    terminal.NewTerminal(app, false),
		currentStep: 0,
		initialized: false,
	}
	return h
}

// HandleEasyManagement 处理简易模式的管理界面
func (h *EasyHandler) HandleEasyManagement() {
	// 如果尚未初始化，则初始化
	if !h.initialized {
		h.initUI()
		h.initialized = true
	}
}

// initUI 初始化用户界面
func (h *EasyHandler) initUI() {
	// 创建一个新的Flex布局作为主容器
	mainFlex := tview.NewFlex().SetDirection(tview.FlexRow)

	// 添加标题
	title := tview.NewTextView().
		SetText("简易模式 - ADB烧录").
		SetTextAlign(tview.AlignCenter).
		SetTextColor(tview.Styles.TitleColor)
	mainFlex.AddItem(title, 1, 1, false)

	// 创建下方的水平分割Flex
	bottomFlex := tview.NewFlex().SetDirection(tview.FlexColumn)

	// 获取终端布局
	terminalFlex := h.Terminal.GetFlex()

	// 创建左侧选项列表 - 使用标准快捷键
	optionsList := tview.NewList().ShowSecondaryText(false)
	optionsList.SetBorder(true).SetTitle(" 操作步骤 ").SetTitleAlign(tview.AlignLeft)
	// 添加步骤选项
	optionsList.AddItem("1. 设备检测(adb devices)", "检测连接的设备", '1', func() {
		h.detectDevices(optionsList)
	})
	optionsList.AddItem("2. 设备选择", "选择要操作的设备", '2', func() {
		h.selectDevice(optionsList)
	})
	optionsList.AddItem("3. 设置号码", "输入手机号码", '3', func() {
		h.setPhoneNumber(optionsList)
	})
	optionsList.AddItem("4. 设置Webhook URL", "输入Webhook URL", '4', func() {
		h.setWebhookURL(optionsList)
	})
	optionsList.AddItem("返回主菜单 (Q/Esc)", "返回主菜单", 'q', func() {
		h.App.QueueUpdateDraw(func() {
			h.Pages.SwitchToPage("main_menu")
		})
	})

	// 初始禁用后续步骤
	optionsList.SetItemText(1, "2. 设备选择 (请先完成步骤1)", "选择要操作的设备")
	optionsList.SetItemText(2, "3. 设置号码 (请先完成步骤2)", "输入手机号码")
	optionsList.SetItemText(3, "4. 设置Webhook URL (请先完成步骤3)", "输入Webhook URL")

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
		if event.Key() == tcell.KeyBacktab {
			h.App.SetFocus(optionsList)
			return nil
		}
		return event
	})

	// 将左侧选项列表和右侧终端添加到底部Flex中
	bottomFlex.AddItem(optionsList, 35, 1, true)  // 左侧菜单宽度固定为35字符
	bottomFlex.AddItem(terminalFlex, 0, 2, false) // 终端占据剩余空间

	// 将底部Flex添加到主Flex中
	mainFlex.AddItem(bottomFlex, 0, 1, true)

	// 添加初始提示
	h.Terminal.WriteOutput("[yellow]简易模式已启动！[white]\n")
	h.Terminal.WriteOutput("[green]请按照步骤进行操作[white]\n")
	h.Terminal.WriteOutput("[red]注意：必须按顺序完成每一步[white]\n\n")
	// 将页面添加到页面管理器
	h.Pages.AddPage("easy_mode", mainFlex, true, false)
	// 添加全局按键捕获
	mainFlex.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyEscape || event.Rune() == 'q' || event.Rune() == 'Q' {
			h.App.QueueUpdateDraw(func() {
				h.Pages.SwitchToPage("main_menu")
			})
			return nil
		}
		return event
	})
}

// enableNextStep 启用下一步骤
func (h *EasyHandler) enableNextStep(optionsList *tview.List) {
	h.currentStep++
	// 启用下一步骤（如果存在）
	if h.currentStep < 4 {
		switch h.currentStep {
		case 1:
			optionsList.SetItemText(1, "2. 设备选择", "选择要操作的设备")
		case 2:
			optionsList.SetItemText(2, "3. 设置号码", "输入手机号码")
		case 3:
			optionsList.SetItemText(3, "4. 设置Webhook URL", "输入Webhook URL")
		}
	}
	// 重新绘制列表
	h.App.Draw()
}

// detectDevices 执行设备检测
func (h *EasyHandler) detectDevices(optionsList *tview.List) {
	if h.currentStep != 0 {
		h.Terminal.WriteOutput("[red]请先完成上一步操作！[white]\n")
		return
	}
	h.Terminal.WriteOutput("[green]开始检测设备...[white]\n")
	// 根据操作系统确定adb路径
	var adbPath string
	switch runtime.GOOS {
	case "windows":
		adbPath = filepath.Join("adb", "windows", "adb.exe")
	case "darwin":
		adbPath = filepath.Join("adb", "darwin", "adb")
	case "linux":
		adbPath = filepath.Join("adb", "linux", "adb")
	default:
		h.Terminal.WriteOutput("[red]不支持的操作系统[white]\n")
		return
	}

	// 执行adb devices命令
	cmd := exec.Command(adbPath, "devices")
	output, err := cmd.CombinedOutput()
	if err != nil {
		h.Terminal.WriteOutput("[red]执行adb devices失败: " + err.Error() + "[white]\n")
		return
	}

	// 解析输出，获取设备列表
	scanner := bufio.NewScanner(strings.NewReader(string(output)))
	h.devices = nil // 清空之前的设备列表
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasSuffix(line, "device") {
			fields := strings.Fields(line)
			if len(fields) >= 2 {
				h.devices = append(h.devices, fields[0])
			}
		}
	}
	if len(h.devices) == 0 {
		h.Terminal.WriteOutput("[red]未检测到设备，请连接设备后重试[white]\n")
		return
	}

	// 显示设备列表
	h.Terminal.WriteOutput("[green]检测到的设备：[white]\n")
	for i, device := range h.devices {
		h.Terminal.WriteOutput(fmt.Sprintf("%d. %s\n", i+1, device))
	}
	h.Terminal.WriteOutput("\n[green]设备检测完成！[white]\n")
	// 启用下一步骤
	h.enableNextStep(optionsList)
}

// selectDevice 选择设备
func (h *EasyHandler) selectDevice(optionsList *tview.List) {
	if h.currentStep != 1 {
		h.Terminal.WriteOutput("[red]请先完成设备检测！[white]\n")
		return
	}
	if len(h.devices) == 0 {
		h.Terminal.WriteOutput("[red]没有可选的设备，请先检测设备[white]\n")
		return
	}

	// 创建模态框选择设备
	modal := tview.NewModal().
		SetText("请选择要操作的设备：").
		AddButtons(h.devices).
		AddButtons([]string{"取消"})
	modal.SetDoneFunc(func(buttonIndex int, buttonLabel string) {
		if buttonIndex < len(h.devices) {
			h.selectedDevice = buttonLabel
			h.Terminal.WriteOutput(fmt.Sprintf("[green]已选择设备: %s[white]\n", h.selectedDevice))
			h.enableNextStep(optionsList)
		}
		h.App.QueueUpdateDraw(func() {
			h.Pages.HidePage("device_select")
		})
	})
	h.Pages.AddPage("device_select", modal, true, true)
	h.App.SetFocus(modal)
}

// setPhoneNumber 设置号码
func (h *EasyHandler) setPhoneNumber(optionsList *tview.List) {
	if h.currentStep != 2 {
		h.Terminal.WriteOutput("[red]请先选择设备！[white]\n")
		return
	}
	// 创建输入框
	inputField := tview.NewInputField().
		SetLabel("请输入手机号码: ").
		SetFieldWidth(20)
	inputField.SetDoneFunc(func(key tcell.Key) {
		if key == tcell.KeyEnter {
			h.phoneNumber = inputField.GetText()
			if h.phoneNumber != "" {
				h.Terminal.WriteOutput(fmt.Sprintf("[green]已设置号码: %s[white]\n", h.phoneNumber))
				h.enableNextStep(optionsList)
			}
			h.App.QueueUpdateDraw(func() {
				h.Pages.HidePage("phone_input")
			})
		} else if key == tcell.KeyEsc {
			h.App.QueueUpdateDraw(func() {
				h.Pages.HidePage("phone_input")
			})
		}
	})
	// 创建模态框容器
	flex := tview.NewFlex().SetDirection(tview.FlexRow)
	flex.AddItem(tview.NewBox(), 0, 1, false)
	flex.AddItem(inputField, 3, 1, true)
	flex.AddItem(tview.NewBox(), 0, 1, false)
	h.Pages.AddPage("phone_input", flex, true, true)
	h.App.SetFocus(inputField)
}

// setWebhookURL 设置Webhook URL
func (h *EasyHandler) setWebhookURL(optionsList *tview.List) {
	if h.currentStep != 3 {
		h.Terminal.WriteOutput("[red]请先设置手机号码！[white]\n")
		return
	}
	// 创建输入框
	inputField := tview.NewInputField().
		SetLabel("请输入Webhook URL: ").
		SetFieldWidth(40)
	inputField.SetDoneFunc(func(key tcell.Key) {
		if key == tcell.KeyEnter {
			h.webhookURL = inputField.GetText()
			if h.webhookURL != "" {
				h.Terminal.WriteOutput(fmt.Sprintf("[green]已设置Webhook URL: %s[white]\n", h.webhookURL))
				h.Terminal.WriteOutput("[green]所有设置已完成！[white]\n")
				h.executeBurning()
			}
			h.App.QueueUpdateDraw(func() {
				h.Pages.HidePage("webhook_input")
			})
		} else if key == tcell.KeyEsc {
			h.App.QueueUpdateDraw(func() {
				h.Pages.HidePage("webhook_input")
			})
		}
	})
	// 创建模态框容器
	flex := tview.NewFlex().SetDirection(tview.FlexRow)
	flex.AddItem(tview.NewBox(), 0, 1, false)
	flex.AddItem(inputField, 3, 1, true)
	flex.AddItem(tview.NewBox(), 0, 1, false)
	h.Pages.AddPage("webhook_input", flex, true, true)
	h.App.SetFocus(inputField)
}

// executeBurning 执行烧录操作
func (h *EasyHandler) executeBurning() {
	h.Terminal.WriteOutput("[green]开始执行烧录操作...[white]\n")
	// 根据操作系统确定adb路径
	var adbPath string
	switch runtime.GOOS {
	case "windows":
		adbPath = filepath.Join("adb", "windows", "adb.exe")
	case "darwin":
		adbPath = filepath.Join("adb", "darwin", "adb")
	case "linux":
		adbPath = filepath.Join("adb", "linux", "adb")
	default:
		h.Terminal.WriteOutput("[red]不支持的操作系统[white]\n")
		return
	}
	// 设置设备ID
	h.Terminal.ExecuteCommand(fmt.Sprintf("%s -s %s devices", adbPath, h.selectedDevice))
	// 设置号码 (示例命令，实际命令根据需求调整)
	h.Terminal.ExecuteCommand(fmt.Sprintf("%s -s %s shell am broadcast -a com.example.SET_NUMBER --es number \"%s\"",
		adbPath, h.selectedDevice, h.phoneNumber))
	// 设置Webhook URL (示例命令，实际命令根据需求调整)
	h.Terminal.ExecuteCommand(fmt.Sprintf("%s -s %s shell am broadcast -a com.example.SET_WEBHOOK --es url \"%s\"",
		adbPath, h.selectedDevice, h.webhookURL))
	// 重启设备
	h.Terminal.ExecuteCommand(fmt.Sprintf("%s -s %s reboot", adbPath, h.selectedDevice))
	h.Terminal.WriteOutput("[green]烧录完成！设备正在重启...[white]\n")
}
