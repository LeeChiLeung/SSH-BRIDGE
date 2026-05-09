## 1. Data Model And Import/Export Services

- [x] 1.1 定义对外使用的 profile 导入导出 JSON 结构，包含 schema version、导出时间和 profiles 数组
- [x] 1.2 在存储层实现 profile 导出能力，将当前全部 profile 转换为稳定 JSON 结构
- [x] 1.3 在导入流程中实现 JSON 解析、schema 校验和 title 为空等非法项检测
- [x] 1.4 实现基于 profile title 的冲突预检查，并生成新增项、冲突项和非法项摘要
- [x] 1.5 实现正式导入逻辑，支持 conflict policy 为 skip 或 merge，并输出 added/skipped/merged/invalid 统计结果

## 2. HTTP Endpoints And Controller Integration

- [x] 2.1 为新控制台增加导出端点，返回 JSON 文件下载响应
- [x] 2.2 为新控制台增加导入预检查端点，接收上传文件并返回解析结果
- [x] 2.3 为新控制台增加导入确认端点，接收冲突处理策略并执行保存
- [x] 2.4 为导入导出流程补充错误响应和日志记录，确保无效文件或部分失败时能明确反馈

## 3. WebUI And Validation

- [x] 3.1 在新控制台首页增加“导出配置”和“导入配置”按钮，并与当前布局对齐
- [x] 3.2 实现导入文件选择和预检查结果展示，列出新增项、冲突项和非法项
- [x] 3.3 在存在冲突时提供用户可选的 skip 与 merge 操作入口，并显示合并覆盖说明
- [x] 3.4 在导入完成后显示 added/skipped/merged/invalid 汇总结果，并刷新首页列表
- [x] 3.5 为导入导出服务、HTTP 处理器和 UI 关键流程补充测试
