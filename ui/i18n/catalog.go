package i18n

import "strings"

const (
	LangZH = "zh-CN"
	LangEN = "en"
)

type Translator struct {
	Lang string
}

func New(lang string) Translator {
	return Translator{Lang: Normalize(lang)}
}

func Normalize(lang string) string {
	switch strings.ToLower(strings.TrimSpace(lang)) {
	case "en", "en-us", "en-gb":
		return LangEN
	default:
		return LangZH
	}
}

func (t Translator) T(key string) string {
	if msg, ok := catalog[t.Lang][key]; ok {
		return msg
	}
	if msg, ok := catalog[LangZH][key]; ok {
		return msg
	}
	return key
}

var catalog = map[string]map[string]string{
	LangZH: {
		"app.title":                 "Timber SSH Go",
		"home.subtitle":             "这是新的 chi + templ WebUI 迁移外壳。当前先迁移首页概览、详情和表单，复杂操作仍会逐步从 legacy 控制台迁入。",
		"home.empty":                "当前没有可显示的 tunnel profile。",
		"home.refresh":              "刷新",
		"home.new":                  "新增配置",
		"home.view":                 "查看详情",
		"home.actions":              "操作",
		"home.detail_panel":         "详情展示",
		"home.open_legacy":          "打开 Legacy 控制台",
		"home.legacy_hint":          "Legacy 控制台通过 --web 启动",
		"detail.title":              "配置详情",
		"detail.edit":               "编辑",
		"detail.back":               "返回首页",
		"detail.start":              "启动",
		"detail.stop":               "停止",
		"detail.delete":             "删除",
		"detail.diagnostics":        "诊断信息",
		"detail.summary":            "基础信息",
		"detail.listener_ready":     "本地监听就绪",
		"detail.ssh_connected":      "SSH 已连接",
		"detail.active_connections": "活跃连接数",
		"detail.upload_rate":        "上行速率",
		"detail.download_rate":      "下行速率",
		"detail.upload_total":       "累计上行",
		"detail.download_total":     "累计下行",
		"detail.last_error":         "最后错误",
		"detail.logs":               "日志",
		"detail.no_logs":            "当前没有日志。",
		"detail.none":               "无",
		"detail.yes":                "是",
		"detail.no":                 "否",
		"detail.refresh_now":        "立即刷新",
		"detail.auto_refresh":       "自动刷新",
		"form.new_title":            "新增配置",
		"form.edit_title":           "编辑配置",
		"form.save":                 "保存",
		"form.cancel":               "取消",
		"form.close":                "关闭",
		"form.password_notice":      "密码优先于私钥认证。当前 profile 会把密码和私钥路径一起保存到本地配置文件。",
		"field.ssh":                 "SSH",
		"field.local":               "本地入口",
		"field.remote":              "远程目标",
		"field.enabled_at":          "启用时间",
		"field.state":               "状态",
		"field.title":               "标题",
		"field.ssh_host":            "SSH 地址",
		"field.ssh_port":            "SSH 端口",
		"field.ssh_user":            "SSH 用户",
		"field.ssh_password":        "SSH 密码",
		"field.ssh_key":             "私钥路径",
		"field.local_addr":          "本地地址",
		"field.local_port":          "本地端口",
		"field.remote_host":         "远程地址",
		"field.remote_port":         "远程端口",
		"field.timeout":             "超时秒数",
		"field.known_hosts":         "Known Hosts 路径",
		"field.ignore_host_key":     "忽略 Host Key 校验",
		"placeholder.key":           "密码为空时才会使用私钥",
		"lang.switch":               "语言",
		"state.connected":           "已连接",
		"state.connecting":          "连接中",
		"state.retrying":            "重试中",
		"state.disconnected":        "未连接",
	},
	LangEN: {
		"app.title":                 "Timber SSH Go",
		"home.subtitle":             "This is the new chi + templ WebUI shell. The home overview, detail view, and form are migrated first, while more complex actions will continue moving over from the legacy console.",
		"home.empty":                "No tunnel profiles are available yet.",
		"home.refresh":              "Refresh",
		"home.new":                  "New Profile",
		"home.view":                 "View Details",
		"home.actions":              "Actions",
		"home.detail_panel":         "Detail Panel",
		"home.open_legacy":          "Open Legacy Console",
		"home.legacy_hint":          "Use --web to launch the legacy console",
		"detail.title":              "Profile Detail",
		"detail.edit":               "Edit",
		"detail.back":               "Back to Home",
		"detail.start":              "Start",
		"detail.stop":               "Stop",
		"detail.delete":             "Delete",
		"detail.diagnostics":        "Diagnostics",
		"detail.summary":            "Summary",
		"detail.listener_ready":     "Local Listener Ready",
		"detail.ssh_connected":      "SSH Connected",
		"detail.active_connections": "Active Connections",
		"detail.upload_rate":        "Upload Rate",
		"detail.download_rate":      "Download Rate",
		"detail.upload_total":       "Upload Total",
		"detail.download_total":     "Download Total",
		"detail.last_error":         "Last Error",
		"detail.logs":               "Logs",
		"detail.no_logs":            "No logs yet.",
		"detail.none":               "None",
		"detail.yes":                "Yes",
		"detail.no":                 "No",
		"detail.refresh_now":        "Refresh Now",
		"detail.auto_refresh":       "Auto Refresh",
		"form.new_title":            "New Profile",
		"form.edit_title":           "Edit Profile",
		"form.save":                 "Save",
		"form.cancel":               "Cancel",
		"form.close":                "Close",
		"form.password_notice":      "Password authentication takes priority over private key authentication. The current profile persists both the password and the key path in the local config file.",
		"field.ssh":                 "SSH",
		"field.local":               "Local Endpoint",
		"field.remote":              "Remote Target",
		"field.enabled_at":          "Enabled At",
		"field.state":               "State",
		"field.title":               "Title",
		"field.ssh_host":            "SSH Host",
		"field.ssh_port":            "SSH Port",
		"field.ssh_user":            "SSH User",
		"field.ssh_password":        "SSH Password",
		"field.ssh_key":             "Private Key Path",
		"field.local_addr":          "Local Address",
		"field.local_port":          "Local Port",
		"field.remote_host":         "Remote Host",
		"field.remote_port":         "Remote Port",
		"field.timeout":             "Timeout Seconds",
		"field.known_hosts":         "Known Hosts Path",
		"field.ignore_host_key":     "Ignore Host Key Verification",
		"placeholder.key":           "Private key is only used when password is empty",
		"lang.switch":               "Language",
		"state.connected":           "Connected",
		"state.connecting":          "Connecting",
		"state.retrying":            "Retrying",
		"state.disconnected":        "Disconnected",
	},
}
