# Timber SSH Go

一个用 Go 编写的本地 SSH 隧道工具，支持本地端口转发、多实例 profile、本地 Web 控制台、系统托盘入口，以及 Linux/Windows 发布产物。

当前项目适合这类场景：

- 通过 SSH jump server 把本地端口转发到远端内网服务
- 保存多套常用 tunnel profile，按卡片方式启停
- 在本地查看连接状态、重连日志、健康检查、上下行速度和累计流量
- 打包成裸二进制、`.deb`、`AppImage`

## 核心能力

- 本地端口转发：监听本地 TCP 端口，转发到 SSH 后面的目标地址
- 连接保活：指数退避重连、SSH keepalive、半开连接检测
- 多实例管理：同一进程内管理多个 tunnel profile
- 本地 WebUI：新增、编辑、复制、删除、启动、停止 profile
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
  --ssh-host 8.138.105.156 \
  --ssh-port 7659 \
  --ssh-user root \
  --ssh-key ~/.ssh/id_rsa \
  --insecure-ignore-host-key \
  --local-addr 127.0.0.1 \
  --local-port 3180 \
  --remote-host 127.0.0.1 \
  --remote-port 7897
```

这等价于类似下面这条 SSH 命令：

```bash
ssh -p 7659 -N -C -L 127.0.0.1:3180:127.0.0.1:7897 root@8.138.105.156
```

查看版本：

```bash
go run . --version
```

启动本地 Web 控制台：

```bash
go run . --web
```

默认访问地址：

```text
http://127.0.0.1:8848
```

启动托盘模式：

```bash
go run . --tray
```

托盘模式会同时启动本地 Web 控制台，并在浏览器中通过托盘菜单打开 `http://127.0.0.1:8848`。

## 常用运行模式

### 1. 直接隧道模式

适合命令行单条转发：

```bash
go run . \
  --ssh-host ssh.example.com \
  --ssh-port 22 \
  --ssh-user user \
  --ssh-key ~/.ssh/id_rsa \
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
  --ssh-key ~/.ssh/id_rsa \
  --local-port 0 \
  --env)"
```

程序会输出类似：

```bash
export TUNNEL_LOCAL_PORT=43017
```

### 3. Web 控制台模式

```bash
go run . --web --web-addr 127.0.0.1:8848
```

WebUI 当前提供：

- tunnel profile 列表首页
- 卡片式布局
- 启动 / 停止
- 编辑 / 复制 / 删除
- 实时状态：`disconnected`、`connecting`、`connected`、`retrying`
- 健康检查：监听状态、SSH 在线状态
- 日志面板、流量统计、活跃连接数

### 4. 托盘模式

```bash
go run . --tray --web-addr 127.0.0.1:8848
```

当前托盘菜单：

- 打开控制台
- 退出

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

- 保存的是 profile 参数，不会持久化 SSH 密码或私钥内容
- 会保存私钥路径
- 会记录上次启用时间

## 打包与发布

默认构建矩阵：

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
APPIMAGE_RUNTIME_FILE=/path/to/runtime-x86_64 ./scripts/release.sh appimage amd64
```

默认产物输出到：

```text
dist/
```

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
- `--web-addr 127.0.0.1:8848`

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
