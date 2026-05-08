# UI Workspace

这个目录预留给新的 `chi + templ` WebUI。

当前约定：

- `pages/`: 页面级 templ 组件
- `components/`: 可复用 UI 组件
- `assets/`: CSS、图片、少量前端脚本
- `i18n/`: 后续中英文字典与文案适配层
- `viewmodel/`: 新 UI 使用的页面数据模型

迁移约束：

- `ui/` 只承载表现层，不复制 tunnel controller、profile store 或运行状态逻辑
- 迁移期由 `--web` 保留 legacy `web.go`，`--web-new` 提供新的 `chi + templ` 入口
