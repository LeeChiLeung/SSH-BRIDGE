## Why

当前控制台已经支持多隧道 profile 的创建、编辑和本地保存，但缺少在机器之间迁移配置、备份配置、批量恢复配置的能力。随着 WebUI 成为主要入口，导入/导出 profile 已经成为高频操作，需要有明确的 JSON 交换格式和冲突处理规则。

## What Changes

- 在新控制台首页新增“导出配置”和“导入配置”按钮。
- 支持将全部 tunnel profiles 导出为一个 JSON 文件。
- 支持从 JSON 文件导入 tunnel profiles。
- 导入时按隧道名称识别冲突；当发现同名 profile 时，允许用户选择跳过或合并。
- 为导入流程增加预检查和确认步骤，避免无提示覆盖已有配置。

## Capabilities

### New Capabilities
- `profile-import-export`: 定义 tunnel profile 的 JSON 导出格式、导入流程、冲突检测以及跳过/合并确认行为。

### Modified Capabilities

## Impact

- 受影响代码包括 `ui/` 新控制台页面、表单/动作处理器、`ProfileStore` 相关存取逻辑，以及可能新增的导入导出 HTTP 端点。
- 需要定义稳定的 JSON 文件结构，供后续版本兼容读取。
- 需要在 WebUI 中增加文件上传、导出下载和冲突确认交互。
