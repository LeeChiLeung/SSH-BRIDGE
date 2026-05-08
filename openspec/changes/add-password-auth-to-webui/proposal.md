## Why

当前 WebUI 只能录入私钥路径，无法直接配置 SSH 密码认证，这会让一部分仅使用密码登录的目标机无法通过控制台管理。同时，现有认证顺序需要明确成“密码优先，密码为空时再使用私钥”，避免配置了两者时行为不确定。

## What Changes

- 为 WebUI 的 tunnel profile 表单增加 SSH 密码输入项。
- 调整 profile 保存与运行时配置映射，使 Web 层可以把密码传递给 SSH 连接逻辑。
- 明确并实现 SSH 认证顺序：当密码非空时优先使用密码认证；只有密码为空时才尝试私钥认证。
- 在相关规格中更新 Web 控制台支持的认证字段和 SSH 认证行为。

## Capabilities

### New Capabilities

### Modified Capabilities
- `web-control-console`: tunnel profile 的表单和保存行为需要支持 SSH 密码输入。
- `ssh-tunnel-cli`: 认证策略需要明确为密码优先、密码为空才使用私钥。

## Impact

- 影响 WebUI 表单、profile 请求解析、profile 数据结构和运行时配置转换
- 影响 SSH 鉴权方法的选择顺序
- 需要补充或调整对应测试，覆盖密码认证和密码优先级
