package pages

import (
	"AlicePushBotBurningTool/console"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

// MainMenu 主菜单页面结构体
type MainMenu struct {
	flex        *tview.Flex // 主布局
	app         *tview.Application
	title       *tview.TextView      // 工具标题
	notice      *tview.TextView      // 公告栏
	list        *tview.List          // 菜单列表
	footer      *tview.TextView      // 页脚提示
	pages       *tview.Pages         // 页面管理器
	easyHandler *console.EasyHandler // 简易模式处理器
}

// NewMainMenu 创建主菜单实例
func NewMainMenu(app *tview.Application, pages *tview.Pages) *MainMenu {
	menu := &MainMenu{
		app:   app,
		pages: pages,
		flex: tview.NewFlex().
			SetDirection(tview.FlexRow), // 垂直布局
	}
	// 初始化UI组件
	menu.initTitle()
	menu.initNotice()
	menu.initList()
	menu.initFooter()
	// 组合组件
	menu.flex.
		AddItem(menu.title, 2, 1, false).   // 工具标题（不可聚焦）
		AddItem(menu.notice, 13, 1, false). // 公告栏（不可聚焦）
		AddItem(menu.list, 0, 2, true).     // 菜单列表（默认聚焦）
		AddItem(menu.footer, 1, 1, false)   // 页脚提示（不可聚焦）
	return menu
}

// SetEasyHandler 设置简易模式处理器
func (m *MainMenu) SetEasyHandler(handler *console.EasyHandler) {
	m.easyHandler = handler
}

// initTitle 初始化工具标题
func (m *MainMenu) initTitle() {
	m.title = tview.NewTextView().
		SetDynamicColors(true).
		SetTextAlign(tview.AlignCenter)
	// 设置工具名称
	titleText := `[::b]AlicePushBotBurningTool | 爱丽丝推送机器人烧录工具[::-]`
	m.title.SetText(titleText)
}

// initNotice 初始化公告栏
func (m *MainMenu) initNotice() {
	m.notice = tview.NewTextView().
		SetDynamicColors(true).
		SetTextAlign(tview.AlignCenter).
		SetWordWrap(true)
	// 设置公告内容
	noticeText := `[::u]理论通杀中兴微所有设备[::-]，已确认通过以下设备：
  - 中兴系列设备
  - 1869系列设备
  - MF782等设备

[::u]使用前请确认设备是否可以读写[::-]：
  - 无法读写的设备：启动脚本无法刷入设备中
  - 只能临时试用，重启后失效

[::u]批量刷入脚本正在开发中[::-]，如有需要请联系：
  [yellow]微信: Amaiyashion`
	m.notice.SetText(noticeText)
	m.notice.SetBorder(true).SetTitle(" 重要公告 ")
}

// initList 初始化菜单列表
func (m *MainMenu) initList() {
	m.list = tview.NewList().
		SetWrapAround(false).
		SetSecondaryTextColor(tview.Styles.SecondaryTextColor)
	// 添加菜单项
	m.list.
		AddItem("1. 简易模式", "简易刷入模式（推荐小白）", '1', m.onSimpleModeSelected).
		AddItem("2. 高级模式", "高级模式（支持自定义文件目录等全部参数）", '2', m.onAdvancedModeSelected).
		AddItem("3. 临时模式", "临时刷入模式（重启后失效）", '3', m.onTemporaryModeSelected).
		AddItem("Q. 退出程序", "关闭应用程序", 'q', m.onExitSelected)
	m.list.SetBorder(true).SetTitle(" 请选择操作模式 ")
}

// initFooter 初始化页脚提示（不可选择）
func (m *MainMenu) initFooter() {
	m.footer = tview.NewTextView().
		SetDynamicColors(true).
		SetTextAlign(tview.AlignCenter).
		SetText("[gray]操作提示: ↑/↓ 移动 | 1/2/3/Q 快捷键 | Enter 选择")
	// 设置为不可聚焦
	m.footer.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		return event // 阻止所有输入事件
	})
}

// 实现完整的Primitive接口
func (m *MainMenu) Blur() {
	m.list.Blur()
}

func (m *MainMenu) Draw(screen tcell.Screen) {
	m.flex.Draw(screen)
}

func (m *MainMenu) Focus(delegate func(p tview.Primitive)) {
	delegate(m.list) // 默认聚焦到菜单列表
}

func (m *MainMenu) GetRect() (int, int, int, int) {
	return m.flex.GetRect()
}

func (m *MainMenu) HasFocus() bool {
	return m.list.HasFocus()
}

func (m *MainMenu) InputHandler() func(event *tcell.EventKey, setFocus func(p tview.Primitive)) {
	return m.list.InputHandler()
}

func (m *MainMenu) MouseHandler() func(action tview.MouseAction, event *tcell.EventMouse, setFocus func(p tview.Primitive)) (consumed bool, capture tview.Primitive) {
	return m.list.MouseHandler()
}

func (m *MainMenu) PasteHandler() func(pasted string, setFocus func(p tview.Primitive)) {
	return m.list.PasteHandler()
}

func (m *MainMenu) SetRect(x, y, width, height int) {
	m.flex.SetRect(x, y, width, height)
}

// 模式选择回调函数
func (m *MainMenu) onSimpleModeSelected() {
	if m.easyHandler != nil {
		// 使用协程避免阻塞主线程
		go func() {
			m.easyHandler.HandleEasyManagement()
			// UI操作需回到主线程
			m.app.QueueUpdateDraw(func() {
				m.pages.SwitchToPage("easy_mode")
			})
		}()
	}
}

func (m *MainMenu) onAdvancedModeSelected() {
	// 切换到高级模式
}

func (m *MainMenu) onTemporaryModeSelected() {
	// 切换到临时模式
}

func (m *MainMenu) onExitSelected() {
	m.app.Stop()
}
