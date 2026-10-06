# AlicePushBotBurningTool

中兴微随身路由的 Webhook 刷入工具。通过 ADB 查询设备，并在终端界面中执行刷入操作。

本仓库不再维护。相关项目：

- [alice-pusher-bot-zxic](https://github.com/Amamiyashi0n/alice-pusher-bot-zxic)：运行于中兴微设备的 Alice Pusher Bot。
- [alice-nl80211-webui-zxic](https://github.com/Amamiyashi0n/alice-nl80211-webui-zxic)：全新改版。

## 当前实现

- 使用 Go 和 tview 提供终端界面。
- 简易模式调用随仓库提供的 ADB 查询已连接设备。
- 菜单预留了简易、高级和临时三种刷入模式。
- 公告中列出的目标包括中兴系列、1869 系列和 MF782 等中兴微设备。

当前源码中的烧录、日志和参数命令仍是占位实现，不构成完整的刷入流程。
