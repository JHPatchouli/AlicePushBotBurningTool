package main

import (
	"AlicePushBotBurningTool/pages"

	"github.com/rivo/tview"
)

func main() {
	app := tview.NewApplication()
	// 创建页面管理器
	pages.CreateMainMenu(app)
}
