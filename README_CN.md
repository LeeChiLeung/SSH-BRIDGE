# SSH BRIDGE

一个用 Go 编写的本地 SSH 隧道工具，支持本地端口转发、多实例 profile、新旧两套本地 Web 控制台、系统托盘入口，以及 Linux/Windows 发布产物。

当前项目适合这类场景：

- 通过 SSH jump server 把本地端口转发到远端内网服务
- 保存多套常用 tunnel profile，按列表/工作台方式启停
- 在本地查看连接状态、重连日志、健康检查、上下行速度和累计流量
- 通过 JSON 导入/导出 profile 配置
- 打包成裸二进制、`.deb`、`AppImage`

## 核心能力

- 本地端口转发：监听本地 TCP 端口，转发到 SSH 后面的目标地址
- 连接保活：指数退避重连、SSH keepalive、半开连接检测
- 多实例管理：同一进程内管理多个 tunnel profile
- 本地 WebUI：新增、编辑、复制、删除、启动、停止、导入、导出 profile
- 系统托盘：打开控制台、退出程序
- 运行指标：上下行速度、累计流量、活跃连接数、运行日志
- 版本注入：`--version` 可输出版本、commit、构建时间

## 环境要求

- Go 1.20+
- Linux 或 Windows 用于当前默认构建矩阵
- 打包 AppImage 时需要 `appimagetool`
- 打包 `.deb` 时需要 `ar`、`tar`、`gzip`

## 快速开始

直接运行单条隧道：

```bash
go run . \
  --ssh-host ssh.example.com \
  --ssh-port 22 \
  --ssh-user user \
  --ssh-key ~/.ssh/example_id_rsa \
  --insecure-ignore-host-key \
  --local-addr 127.0.0.1 \
  --local-port 15432 \
  --remote-host 127.0.0.1 \
  --remote-port 5432
```

这等价于类似下面这条 SSH 命令：

```bash
ssh -p 22 -N -C -L 127.0.0.1:15432:127.0.0.1:5432 user@ssh.example.com
```

查看版本：

```bash
go run . --version
```

启动旧版本地 Web 控制台：

```bash
go run . --web
```

启动新版本地 Web 控制台：

```bash
go run . --web-new
```

默认访问地址：

```text
http://127.0.0.1:8765
```

启动托盘模式：

```bash
go run . --tray
```

托盘模式会同时启动：

- 旧控制台：`http://127.0.0.1:8765`
- 新控制台：`http://127.0.0.1:8910`

托盘菜单中的“打开控制台”当前会打开新的控制台页面。

## 常用运行模式

### 1. 直接隧道模式

适合命令行单条转发：

```bash
go run . \
  --ssh-host ssh.example.com \
  --ssh-port 22 \
  --ssh-user user \
  --ssh-key ~/.ssh/example_id_rsa \
  --remote-host 127.0.0.1 \
  --remote-port 5432 \
  --local-addr 127.0.0.1 \
  --local-port 15432
```

### 2. `--env` 注入本地端口

如果本地端口使用 `0` 让系统自动分配，可以用 `--env` 输出 shell export 命令：

```bash
eval "$(go run . \
  --ssh-host ssh.example.com \
  --ssh-user user \
  --ssh-key ~/.ssh/example_id_rsa \
  --local-port 0 \
  --env)"
```

程序会输出类似：

```bash
export TUNNEL_LOCAL_PORT=43017
```

### 3. 旧版 Web 控制台模式

```bash
go run . --web --web-addr 127.0.0.1:8765
```

### 4. 新版 Web 控制台模式

```bash
go run . --web-new --web-addr 127.0.0.1:8765
```

新版 WebUI 当前提供：

- tunnel profile 列表首页
- 列表 + 下方工作台详情
- 启动 / 停止
- 编辑 / 复制 / 删除
- 导入 / 导出 JSON 配置
- 实时状态：`disconnected`、`connecting`、`connected`、`retrying`
- 状态徽标颜色：未启动灰色、启动中蓝色、成功绿色、失败红色
- 健康检查：监听状态、SSH 在线状态
- 日志面板、流量统计、活跃连接数

### 5. 托盘模式

```bash
go run . --tray --web-addr 127.0.0.1:8765
```

当前托盘菜单：

- 打开控制台
- 退出

说明：

- 托盘模式内部会同时拉起旧控制台和新控制台
- 托盘入口当前默认打开新控制台
- Windows 双击 `exe` 时默认进入托盘模式

## 主要 CLI 参数

常用参数如下：

- `--ssh-host` / `--ssh-port` / `--ssh-user`
- `--ssh-key` 或 `--ssh-password`
- `--remote-host` / `--remote-port`
- `--local-addr` / `--local-port`
- `--connect-timeout`
- `--known-hosts`
- `--insecure-ignore-host-key`
- `--env`
- `--web`
- `--web-new`
- `--tray`
- `--web-addr`
- `--version`

完整帮助：

```bash
go run . --help
```

## 配置存储

WebUI 使用 profile 持久化保存常用隧道配置。

默认保存位置：

- Linux: `$XDG_CONFIG_HOME/timber-ssh-go/profiles.json`
- 如果系统配置目录不可用，则回退到当前目录下的 `./profiles.json`

说明：

- 会保存 profile 参数，包括 SSH 密码和私钥路径
- 不会保存私钥文件内容本身
- 会记录上次启用时间

### JSON 导入导出

新版控制台支持通过按钮导出全部 profile 为一个 JSON 文件，也支持从 JSON 文件导入。

导入流程特点：

- 导入前会先做预检查
- 按隧道名称（title）识别冲突
- 同名冲突时可选“跳过”或“合并”
- 导入完成后会显示 added / merged / skipped / invalid 汇总结果

## 打包与发布

默认构建矩阵：

```bash
./build.sh
```

也可以显式传参透传给发布脚本：

```bash
./build.sh build
./build.sh deb amd64
./build.sh appimage amd64
```

- `linux/amd64`
- `linux/arm64`
- `windows/amd64`

构建裸二进制：

```bash
./scripts/release.sh build
```

构建全部默认产物：

```bash
./scripts/release.sh all
```

构建 `.deb`：

```bash
./scripts/release.sh deb amd64
```

构建 AppImage：

```bash
./scripts/release.sh appimage amd64
```

默认会使用仓库内的 `public/obsolete-runtime-x86_64` 作为 AppImage runtime；如果需要覆盖，仍可显式传入 `APPIMAGE_RUNTIME_FILE=/path/to/runtime-x86_64`。

默认产物输出到：

```text
dist/
```

Windows 产物当前使用 GUI subsystem 构建：

- 双击启动时不会弹出 `cmd` 控制台窗口
- 默认双击行为会进入托盘模式

### 版本控制

构建脚本会统一处理产物文件名和二进制内部版本信息。

- 如果显式传入 `VERSION`，则使用该版本
- 如果未传入，则自动尝试使用 `git describe --tags --dirty --always`
- 二进制可通过 `--version` 查看注入后的版本、commit、构建时间

示例：

```bash
VERSION=0.0.2 ./scripts/release.sh all
```

生成的文件名会类似：

- `timber-ssh-go_0.0.2_linux_amd64`
- `timber-ssh-go_0.0.2_linux_amd64.deb`
- `timber-ssh-go_0.0.2_linux_amd64.AppImage`

如果工作区有未提交修改，自动推导版本时文件名可能出现类似 `229bbaa-dirty`。

## AppImage 默认行为

AppImage 直接启动时默认进入：

- `--tray`
- `--web-addr 127.0.0.1:8765`

当前打包脚本已强制使用 `xz` 压缩的 squashfs，以兼容较旧的 AppImage runtime。

## 开发与测试

运行测试：

```bash
go test -buildvcs=false ./...
```

如果你在国内网络环境下拉依赖，常见做法是：

```bash
GOPROXY=https://goproxy.cn,direct go test -buildvcs=false ./...
```

## 已知限制

- 默认构建矩阵暂时不包含 macOS
- 原因是当前托盘依赖在这台 Linux 构建环境下无法直接完成 Darwin 交叉构建
- AppImage 打包依赖本机可用的 `appimagetool` 与 runtime 文件
- README 中的示例以当前实现为准，后续如果 CLI 或 WebUI 默认值变化，需要同步更新

## 仓库现状

当前主要能力已经包含：

- SSH 隧道核心转发
- 多实例 tunnel profile
- 本地 Web 控制台
- 托盘入口
- 流量统计与日志
- 版本注入
- Linux/Windows 发布脚本

如果你要继续扩展功能，建议先从 `openspec/changes/` 里的变更记录看起。
