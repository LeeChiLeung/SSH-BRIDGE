## Why

当前 Web 控制台已经能用，但 `web.go` 把路由、模板、文案、样式和前端脚本都塞在一个文件里，后续要做中英切换、组件化和界面持续演进时，维护成本会迅速上升。现在需要把 WebUI 重构成可渐进迁移的新结构，同时保留现有可用界面，避免一次性重写带来的回归风险。

## What Changes

- 在根目录新增 `ui/` 目录，作为新的 Web 表现层实现位置。
- 引入 `chi` 作为新 UI 的路由层，引入 `templ` 作为新的服务端组件模板层。
- 保留现有 `web.go` 作为 legacy WebUI，在迁移期间与新 UI 共存。
- 规定新旧 UI 共用同一套隧道控制、profile 存储和运行状态逻辑，不允许复制业务内核。
- 为后续国际化和组件化重构建立稳定的目录结构、切换入口和迁移路径。

## Capabilities

### New Capabilities
- `templ-web-ui-shell`: 为项目新增一个基于 `chi + templ` 的新 WebUI 外壳和组件化目录结构，支持逐步接管旧界面。

### Modified Capabilities
- `web-control-console`: 本地 Web 控制台的实现方式需要扩展为“legacy 界面 + 新 UI 并存”的渐进迁移模型，而不是只允许单一 `web.go` 模板实现。

## Impact

- 影响 Web 启动入口、路由注册和页面渲染结构
- 引入 `chi` 与 `templ` 依赖及其开发流程
- 需要明确新旧 UI 切换方式、共用业务边界和迁移节奏
- 不直接改变隧道控制、profile 存储、托盘逻辑和打包主线，但会影响它们与 Web 层的集成方式
