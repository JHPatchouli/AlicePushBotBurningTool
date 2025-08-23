package main

import (
	"AlicePushBotBurningTool/console"
	"AlicePushBotBurningTool/pages"

	"github.com/rivo/tview"
)

func main() {
	app := tview.NewApplication()
	// 创建页面管理器
	pagesContainer := tview.NewPages()
	// 创建主菜单页面
	mainMenu := pages.NewMainMenu(app, pagesContainer)
	pagesContainer.AddPage("main_menu", mainMenu, true, true)
	// 创建简易模式处理器
	easyHandler := console.NewEasyHandler(app, pagesContainer)
	// 设置简易模式回调
	mainMenu.SetEasyHandler(easyHandler)
	// 启动应用
	if err := app.SetRoot(pagesContainer, true).Run(); err != nil {
		panic(err)
	}
}
