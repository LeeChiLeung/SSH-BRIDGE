package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"text/template"
	"time"
)

type profileView struct {
	Profile       Profile
	State         TunnelState
	Endpoint      string
	Health        TunnelHealth
	Traffic       TrafficStats
	LastError     string
	IsActive      bool
	LocalPortText string
	RemotePort    int
}

type diagnosticsView struct {
	View profileView
	Logs []LogEntry
}

type detailPanelData struct {
	Mode        string
	Profile     Profile
	View        profileView
	Diagnostics diagnosticsView
	FormAction  string
}

type indexPageData struct {
	PageTitle string
	Cards     []profileView
	Detail    detailPanelData
}

type apiResponse struct {
	OK       bool                      `json:"ok"`
	Message  string                    `json:"message,omitempty"`
	Profile  *Profile                  `json:"profile,omitempty"`
	Profiles []Profile                 `json:"profiles,omitempty"`
	Status   *TunnelSnapshot           `json:"status,omitempty"`
	Statuses map[string]TunnelSnapshot `json:"statuses,omitempty"`
}

type WebServer struct {
	addr       string
	controller *TunnelController
	store      *ProfileStore
	logger     *log.Logger
	templates  *template.Template
}

func NewWebServer(addr string, controller *TunnelController, store *ProfileStore, logger *log.Logger) (*WebServer, error) {
	tpls, err := template.New("page").Funcs(template.FuncMap{
		"formatTime":  formatTime,
		"formatBytes": formatBytes,
		"formatRate":  formatRate,
		"statusLabel": func(state TunnelState) string {
			switch state {
			case TunnelStateConnected:
				return "已连接"
			case TunnelStateConnecting:
				return "连接中"
			case TunnelStateRetrying:
				return "重试中"
			default:
				return "未连接"
			}
		},
	}).Parse(webTemplates)
	if err != nil {
		return nil, err
	}
	return &WebServer{
		addr: addr, controller: controller, store: store, logger: logger, templates: tpls,
	}, nil
}

func (s *WebServer) Run(ctx context.Context) error {
	mux := http.NewServeMux()
	mux.HandleFunc("/", s.handleIndex)
	mux.HandleFunc("/cards", s.handleCards)
	mux.HandleFunc("/profiles/new", s.handleNewProfile)
	mux.HandleFunc("/profiles/save", s.handleSaveProfile)
	mux.HandleFunc("/profiles/", s.handleProfileRoute)
	mux.HandleFunc("/api/status", s.handleAPIStatus)
	mux.HandleFunc("/api/profiles", s.handleAPIProfiles)
	mux.HandleFunc("/api/profiles/", s.handleAPIProfileRoute)

	server := &http.Server{Addr: s.addr, Handler: mux}
	errCh := make(chan error, 1)

	go func() {
		if s.logger != nil {
			s.logger.Printf("web console listening on http://%s", s.addr)
		}
		errCh <- server.ListenAndServe()
	}()

	select {
	case <-ctx.Done():
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		return server.Shutdown(shutdownCtx)
	case err := <-errCh:
		if err == http.ErrServerClosed {
			return nil
		}
		return err
	}
}

func (s *WebServer) handleIndex(w http.ResponseWriter, r *http.Request) {
	cards := s.buildProfileViews()
	detail := s.defaultDetail(cards)
	s.render(w, "page", indexPageData{
		PageTitle: "SSH Tunnel Console",
		Cards:     cards,
		Detail:    detail,
	})
}

func (s *WebServer) handleCards(w http.ResponseWriter, r *http.Request) {
	s.render(w, "cardsPanel", s.buildProfileViews())
}

func (s *WebServer) handleNewProfile(w http.ResponseWriter, r *http.Request) {
	profile := profileFromConfig("", defaultTunnelConfig())
	s.render(w, "detailPanel", detailPanelData{
		Mode:       "form",
		Profile:    profile,
		FormAction: "/profiles/save",
	})
}

func (s *WebServer) handleSaveProfile(w http.ResponseWriter, r *http.Request) {
	profile, err := profileFromRequest(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	saved, err := s.store.Save(profile)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	s.render(w, "detailPanel", s.profileDetail(saved))
}

func (s *WebServer) handleProfileRoute(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/profiles/")
	parts := strings.Split(strings.Trim(path, "/"), "/")
	if len(parts) < 2 {
		http.NotFound(w, r)
		return
	}
	id, action := parts[0], parts[1]

	switch action {
	case "edit":
		profile, ok := s.store.Get(id)
		if !ok {
			http.NotFound(w, r)
			return
		}
		s.render(w, "detailPanel", detailPanelData{
			Mode:       "form",
			Profile:    profile,
			FormAction: "/profiles/save",
		})
	case "detail":
		profile, ok := s.store.Get(id)
		if !ok {
			http.NotFound(w, r)
			return
		}
		s.render(w, "detailPanel", s.profileDetail(profile))
	case "diagnostics":
		profile, ok := s.store.Get(id)
		if !ok {
			http.NotFound(w, r)
			return
		}
		s.render(w, "diagnosticsPanel", s.buildDiagnostics(profile))
	case "start":
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		profile, ok := s.store.Get(id)
		if !ok {
			http.NotFound(w, r)
			return
		}
		if err := s.controller.Start(profile.ID, profile.TunnelConfig()); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		s.render(w, "cardsPanel", s.buildProfileViews())
	case "stop":
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		s.controller.Stop(id)
		s.render(w, "cardsPanel", s.buildProfileViews())
	case "duplicate":
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		profile, err := s.store.Duplicate(id)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		s.render(w, "detailPanel", detailPanelData{
			Mode:       "form",
			Profile:    profile,
			FormAction: "/profiles/save",
		})
	case "delete":
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		snapshot := s.controller.Snapshot(id)
		if snapshot.State != TunnelStateDisconnected {
			http.Error(w, "cannot delete a running profile", http.StatusBadRequest)
			return
		}
		if err := s.store.Delete(id); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		s.render(w, "cardsPanel", s.buildProfileViews())
	default:
		http.NotFound(w, r)
	}
}

func (s *WebServer) handleAPIStatus(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, apiResponse{OK: true, Statuses: s.controller.Snapshots()})
}

func (s *WebServer) handleAPIProfiles(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		writeJSON(w, http.StatusOK, apiResponse{OK: true, Profiles: s.store.List()})
	case http.MethodPost:
		var profile Profile
		if err := json.NewDecoder(r.Body).Decode(&profile); err != nil {
			writeJSON(w, http.StatusBadRequest, apiResponse{OK: false, Message: err.Error()})
			return
		}
		saved, err := s.store.Save(profile)
		if err != nil {
			writeJSON(w, http.StatusBadRequest, apiResponse{OK: false, Message: err.Error()})
			return
		}
		writeJSON(w, http.StatusOK, apiResponse{OK: true, Profile: &saved})
	default:
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}

func (s *WebServer) handleAPIProfileRoute(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/api/profiles/")
	parts := strings.Split(strings.Trim(path, "/"), "/")
	if len(parts) < 2 {
		http.NotFound(w, r)
		return
	}
	id, action := parts[0], parts[1]
	profile, ok := s.store.Get(id)
	if !ok && action != "duplicate" {
		http.NotFound(w, r)
		return
	}

	switch action {
	case "start":
		if err := s.controller.Start(profile.ID, profile.TunnelConfig()); err != nil {
			writeJSON(w, http.StatusBadRequest, apiResponse{OK: false, Message: err.Error()})
			return
		}
		writeJSON(w, http.StatusOK, apiResponse{OK: true, Status: ptrSnapshot(s.controller.Snapshot(id))})
	case "stop":
		s.controller.Stop(id)
		writeJSON(w, http.StatusOK, apiResponse{OK: true, Status: ptrSnapshot(s.controller.Snapshot(id))})
	case "duplicate":
		clone, err := s.store.Duplicate(id)
		if err != nil {
			writeJSON(w, http.StatusBadRequest, apiResponse{OK: false, Message: err.Error()})
			return
		}
		writeJSON(w, http.StatusOK, apiResponse{OK: true, Profile: &clone})
	case "delete":
		snapshot := s.controller.Snapshot(id)
		if snapshot.State != TunnelStateDisconnected {
			writeJSON(w, http.StatusBadRequest, apiResponse{OK: false, Message: "cannot delete a running profile"})
			return
		}
		if err := s.store.Delete(id); err != nil {
			writeJSON(w, http.StatusBadRequest, apiResponse{OK: false, Message: err.Error()})
			return
		}
		writeJSON(w, http.StatusOK, apiResponse{OK: true, Profiles: s.store.List()})
	case "diagnostics":
		writeJSON(w, http.StatusOK, apiResponse{OK: true, Status: ptrSnapshot(s.controller.Snapshot(id))})
	default:
		http.NotFound(w, r)
	}
}

func (s *WebServer) buildProfileViews() []profileView {
	profiles := s.store.List()
	snapshots := s.controller.Snapshots()
	views := make([]profileView, 0, len(profiles))
	for _, profile := range profiles {
		cfg := profile.TunnelConfig()
		view := profileView{
			Profile:       profile,
			State:         TunnelStateDisconnected,
			Endpoint:      cfg.LocalEndpoint(),
			RemotePort:    profile.RemotePort,
			LocalPortText: strconv.Itoa(profile.LocalPort),
		}
		if snapshot, ok := snapshots[profile.ID]; ok {
			view.State = snapshot.State
			view.Health = snapshot.Health
			view.Traffic = snapshot.Traffic
			view.LastError = snapshot.LastError
			view.IsActive = snapshot.State != TunnelStateDisconnected || snapshot.LastError != ""
			if snapshot.Endpoint != "" {
				view.Endpoint = snapshot.Endpoint
			}
		}
		views = append(views, view)
	}
	return views
}

func (s *WebServer) defaultDetail(cards []profileView) detailPanelData {
	if len(cards) == 0 {
		return detailPanelData{Mode: "empty"}
	}
	return s.profileDetail(cards[0].Profile)
}

func (s *WebServer) profileDetail(profile Profile) detailPanelData {
	diagnostics := s.buildDiagnostics(profile)
	return detailPanelData{
		Mode:        "detail",
		Profile:     profile,
		View:        diagnostics.View,
		Diagnostics: diagnostics,
	}
}

func (s *WebServer) buildDiagnostics(profile Profile) diagnosticsView {
	view := profileView{
		Profile:       profile,
		State:         TunnelStateDisconnected,
		Endpoint:      profile.TunnelConfig().LocalEndpoint(),
		RemotePort:    profile.RemotePort,
		LocalPortText: strconv.Itoa(profile.LocalPort),
	}
	var logs []LogEntry
	snapshot := s.controller.Snapshot(profile.ID)
	if snapshot.State != "" {
		view.State = snapshot.State
		view.Health = snapshot.Health
		view.Traffic = snapshot.Traffic
		view.LastError = snapshot.LastError
		view.IsActive = snapshot.State != TunnelStateDisconnected || snapshot.LastError != ""
		if snapshot.Endpoint != "" {
			view.Endpoint = snapshot.Endpoint
		}
		logs = snapshot.Logs
	}
	return diagnosticsView{View: view, Logs: logs}
}

func (s *WebServer) render(w http.ResponseWriter, name string, data any) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := s.templates.ExecuteTemplate(w, name, data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func profileFromRequest(r *http.Request) (Profile, error) {
	if err := r.ParseMultipartForm(8 << 20); err != nil && !strings.Contains(err.Error(), "not multipart") {
		return Profile{}, err
	}

	profile := Profile{
		ID:                 strings.TrimSpace(r.FormValue("id")),
		Title:              strings.TrimSpace(r.FormValue("title")),
		SSHHost:            strings.TrimSpace(r.FormValue("ssh_host")),
		SSHUser:            strings.TrimSpace(r.FormValue("ssh_user")),
		SSHKeyPath:         strings.TrimSpace(r.FormValue("ssh_key_path")),
		RemoteHost:         strings.TrimSpace(r.FormValue("remote_host")),
		LocalAddr:          strings.TrimSpace(r.FormValue("local_addr")),
		KnownHostsPath:     strings.TrimSpace(r.FormValue("known_hosts_path")),
		InsecureIgnoreHost: r.FormValue("insecure_ignore_host") == "on",
	}

	var err error
	if profile.SSHPort, err = strconv.Atoi(r.FormValue("ssh_port")); err != nil {
		return Profile{}, fmt.Errorf("invalid SSH port")
	}
	if profile.LocalPort, err = strconv.Atoi(r.FormValue("local_port")); err != nil {
		return Profile{}, fmt.Errorf("invalid local port")
	}
	if profile.RemotePort, err = strconv.Atoi(r.FormValue("remote_port")); err != nil {
		return Profile{}, fmt.Errorf("invalid remote port")
	}
	timeout, err := strconv.Atoi(r.FormValue("connect_timeout_sec"))
	if err != nil {
		return Profile{}, fmt.Errorf("invalid timeout")
	}
	profile.ConnectTimeoutSec = timeout

	cfg := profile.TunnelConfig()
	cfg.Title = profile.Title
	if err := cfg.ValidateRuntime(); err != nil {
		return Profile{}, err
	}
	return profile, nil
}

func formatTime(ts *time.Time) string {
	if ts == nil {
		return "未启用"
	}
	return ts.Local().Format("2006-01-02 15:04:05")
}

func formatBytes(value uint64) string {
	units := []string{"B", "KB", "MB", "GB", "TB"}
	size := float64(value)
	unit := units[0]
	for idx := 0; idx < len(units)-1 && size >= 1024; idx++ {
		size /= 1024
		unit = units[idx+1]
	}
	if unit == "B" {
		return fmt.Sprintf("%d %s", value, unit)
	}
	return fmt.Sprintf("%.1f %s", size, unit)
}

func formatRate(value uint64) string {
	return formatBytes(value) + "/s"
}

func ptrSnapshot(snapshot TunnelSnapshot) *TunnelSnapshot {
	return &snapshot
}

func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}

const webTemplates = `
{{define "page"}}
<!doctype html>
<html lang="zh-CN">
<head>
  <meta charset="utf-8">
  <meta name="viewport" content="width=device-width, initial-scale=1">
  <title>{{.PageTitle}}</title>
  <style>
    :root { color-scheme: light; --bg:#f5f7fb; --card:#fff; --line:#d9e0ea; --text:#16202f; --muted:#64748b; --accent:#0f766e; --warn:#b45309; --danger:#b91c1c; }
    * { box-sizing:border-box; }
    body { margin:0; font-family:"Noto Sans SC","PingFang SC","Microsoft YaHei",sans-serif; background:linear-gradient(180deg,#f7fafc 0%,#eef3f8 100%); color:var(--text); }
    .page { width:min(1440px, calc(100vw - 48px)); margin:24px auto 40px; }
    .toolbar { display:flex; justify-content:space-between; align-items:flex-end; gap:24px; margin-bottom:20px; }
    .titlebox { background:var(--card); border:1px solid var(--line); border-radius:18px; padding:22px 28px; min-width:260px; box-shadow:0 12px 28px rgba(15,23,42,.06); }
    .titlebox h1 { margin:0 0 8px; font-size:28px; letter-spacing:.04em; }
    .titlebox p { margin:0; color:var(--muted); }
    .addbtn, .btn { border:0; border-radius:14px; padding:12px 18px; font-size:15px; cursor:pointer; }
    .addbtn { background:#111827; color:#fff; box-shadow:0 10px 20px rgba(17,24,39,.18); }
    .toolbar-actions { display:flex; align-items:center; gap:12px; }
    .toggle-refresh { display:inline-flex; align-items:center; gap:10px; background:var(--card); border:1px solid var(--line); border-radius:14px; padding:10px 14px; color:var(--muted); font-size:14px; box-shadow:0 8px 20px rgba(15,23,42,.04); }
    .toggle-refresh input { width:18px; height:18px; }
    .layout { display:grid; gap:22px; }
    .cards-panel, .detail-panel { background:rgba(255,255,255,.64); border:1px solid rgba(217,224,234,.8); border-radius:24px; padding:18px; backdrop-filter:blur(8px); }
    .cards-wrap { display:grid; gap:18px; }
    .card { display:grid; grid-template-columns: minmax(280px,1.2fr) minmax(240px,1fr) 280px; gap:18px; background:var(--card); border:1px solid var(--line); border-radius:22px; padding:16px; box-shadow:0 10px 30px rgba(15,23,42,.05); }
    .zone { border:1px solid #e7ecf3; border-radius:18px; padding:18px 20px; min-height:128px; }
    .zone h3 { margin:0 0 14px; font-size:15px; color:var(--muted); font-weight:600; }
    .info-grid { display:grid; grid-template-columns:repeat(2,minmax(0,1fr)); gap:10px 18px; font-size:14px; }
    .field label { display:block; color:var(--muted); margin-bottom:4px; }
    .field strong, .field span { display:block; word-break:break-all; }
    .state-zone { cursor:pointer; background:linear-gradient(180deg,#f8fbff 0%,#f3f8fd 100%); }
    .badge { display:inline-flex; align-items:center; padding:8px 12px; border-radius:999px; font-size:13px; font-weight:700; letter-spacing:.04em; }
    .badge.connected { background:#dcfce7; color:#166534; }
    .badge.connecting { background:#dbeafe; color:#1d4ed8; }
    .badge.retrying { background:#fef3c7; color:#92400e; }
    .badge.disconnected { background:#e5e7eb; color:#374151; }
    .status-main { display:flex; flex-direction:column; gap:10px; }
    .endpoint { font-size:18px; font-weight:700; }
    .error { color:var(--danger); font-size:13px; }
    .health { display:flex; gap:10px; flex-wrap:wrap; color:var(--muted); font-size:13px; }
    .actions { display:grid; gap:10px; align-content:start; }
    .btn { width:100%; background:#eff6ff; color:#1e3a8a; }
    .btn.stop { background:#fee2e2; color:#991b1b; }
    .btn.edit { background:#ecfeff; color:#115e59; }
    .btn.copy { background:#f3e8ff; color:#6b21a8; }
    .btn.delete { background:#fee2e2; color:#991b1b; }
    .detail-head { display:flex; justify-content:space-between; align-items:center; gap:16px; margin-bottom:16px; }
    .detail-title { font-size:22px; font-weight:700; }
    .detail-sub { color:var(--muted); }
    .detail-grid { display:grid; grid-template-columns:1.1fr 1fr; gap:18px; }
    .diagnostics { background:var(--card); border:1px solid var(--line); border-radius:18px; padding:18px; }
    .logs { max-height:260px; overflow:auto; display:grid; gap:10px; }
    .log { border-left:3px solid #cbd5e1; padding-left:12px; font-size:13px; }
    .log time { color:var(--muted); display:block; margin-bottom:4px; }
    .placeholder { padding:48px 18px; text-align:center; color:var(--muted); }
    .form-grid { display:grid; grid-template-columns:repeat(2,minmax(0,1fr)); gap:14px 18px; }
    .form-field label { display:block; margin-bottom:6px; color:var(--muted); font-size:13px; }
    .form-field input { width:100%; border:1px solid var(--line); border-radius:12px; padding:11px 12px; font-size:14px; }
    .full { grid-column:1 / -1; }
    .form-actions { display:flex; justify-content:flex-end; gap:12px; margin-top:18px; }
    .submit { background:#111827; color:#fff; }
    .ghost { background:#f8fafc; color:#334155; }
    @media (max-width: 1080px) {
      .card, .detail-grid, .form-grid { grid-template-columns:1fr; }
      .page { width:min(100vw - 24px, 1440px); }
    }
  </style>
</head>
<body>
  <main class="page">
    <header class="toolbar">
      <div class="titlebox">
        <h1>{{.PageTitle}}</h1>
        <p>本地 SSH 隧道控制台，按卡片管理 tunnel profile</p>
      </div>
      <div class="toolbar-actions">
        <label class="toggle-refresh">
          <input type="checkbox" id="auto-refresh-toggle">
          <span>开启自动刷新</span>
        </label>
        <button class="addbtn" data-load-url="/profiles/new" data-target="#detail-panel">新增</button>
      </div>
    </header>
    <section class="layout">
      {{template "cardsPanel" .Cards}}
      <section id="detail-panel" class="detail-panel">{{template "detailPanel" .Detail}}</section>
    </section>
  </main>
  <script>
    (() => {
      let diagnosticsTimer = null;
      let autoRefreshEnabled = false;

      async function fetchText(url, options = {}) {
        const response = await fetch(url, {
          headers: { "X-Requested-With": "fetch" },
          ...options,
        });
        if (!response.ok) {
          const message = await response.text();
          throw new Error(message || response.statusText);
        }
        return response.text();
      }

      async function replaceTarget(url, targetSelector, options = {}) {
        const target = document.querySelector(targetSelector);
        if (!target) return;
        try {
          target.innerHTML = await fetchText(url, options);
          if (targetSelector === "#detail-panel") {
            restartDiagnosticsPolling();
          }
        } catch (error) {
          console.error(error);
          alert("请求失败: " + error.message);
        }
      }

      async function refreshCards() {
        const panel = document.querySelector("#cards-panel");
        if (!panel) return;
        try {
          panel.outerHTML = await fetchText("/cards");
        } catch (error) {
          console.error(error);
        }
      }

      function restartDiagnosticsPolling() {
        if (diagnosticsTimer) {
          clearInterval(diagnosticsTimer);
          diagnosticsTimer = null;
        }
        if (!autoRefreshEnabled) return;
        const panel = document.querySelector("#diagnostics-panel");
        if (!panel) return;
        const url = panel.dataset.pollUrl;
        if (!url) return;
        diagnosticsTimer = setInterval(async () => {
          if (document.hidden) return;
          try {
            panel.innerHTML = await fetchText(url);
          } catch (error) {
            console.error(error);
          }
        }, 5000);
      }

      document.addEventListener("click", async (event) => {
        const loadBtn = event.target.closest("[data-load-url]");
        if (loadBtn) {
          event.preventDefault();
          await replaceTarget(loadBtn.dataset.loadUrl, loadBtn.dataset.target || "#detail-panel");
          return;
        }

        const actionBtn = event.target.closest("[data-post-url]");
        if (actionBtn) {
          event.preventDefault();
          if (actionBtn.dataset.confirm && !window.confirm(actionBtn.dataset.confirm)) {
            return;
          }
          try {
            const targetSelector = actionBtn.dataset.target || "#cards-panel";
            const target = document.querySelector(targetSelector);
            if (target) {
              target.outerHTML = await fetchText(actionBtn.dataset.postUrl, { method: "POST" });
            } else {
              await fetchText(actionBtn.dataset.postUrl, { method: "POST" });
            }
            const detailUrl = actionBtn.dataset.detailUrl;
            if (detailUrl) {
              await replaceTarget(detailUrl, "#detail-panel");
            } else {
              restartDiagnosticsPolling();
            }
          } catch (error) {
            console.error(error);
            alert("操作失败: " + error.message);
          }
          return;
        }

        const card = event.target.closest("[data-detail-url]");
        if (card && !event.target.closest("button, input, form, label")) {
          event.preventDefault();
          await replaceTarget(card.dataset.detailUrl, "#detail-panel");
        }
      });

      document.addEventListener("submit", async (event) => {
        const form = event.target.closest("form[data-async-form]");
        if (!form) return;
        event.preventDefault();
        try {
          const html = await fetchText(form.action, {
            method: form.method || "POST",
            body: new FormData(form),
          });
          const targetSelector = form.dataset.target || "#detail-panel";
          const target = document.querySelector(targetSelector);
          if (target) target.innerHTML = html;
          await refreshCards();
          restartDiagnosticsPolling();
        } catch (error) {
          console.error(error);
          alert("保存失败: " + error.message);
        }
      });

      const toggle = document.querySelector("#auto-refresh-toggle");
      if (toggle) {
        toggle.addEventListener("change", () => {
          autoRefreshEnabled = toggle.checked;
          restartDiagnosticsPolling();
        });
      }
    })();
  </script>
</body>
</html>
{{end}}

{{define "cardsPanel"}}
<section id="cards-panel" class="cards-panel">
  <div class="cards-wrap">
    {{if .}}
      {{range .}}
      <article class="card">
        <section class="zone">
          <h3>信息区</h3>
          <div class="info-grid">
            <div class="field"><label>标题</label><strong>{{.Profile.Title}}</strong></div>
            <div class="field"><label>SSH 地址</label><span>{{.Profile.SSHHost}}:{{.Profile.SSHPort}}</span></div>
            <div class="field"><label>代理端口</label><span>{{.Profile.LocalAddr}}:{{.Profile.LocalPort}}</span></div>
            <div class="field"><label>远程端口</label><span>{{.Profile.RemoteHost}}:{{.Profile.RemotePort}}</span></div>
            <div class="field full"><label>启用时间</label><span>{{formatTime .Profile.LastEnabledAt}}</span></div>
          </div>
        </section>
        <section class="zone state-zone" data-detail-url="/profiles/{{.Profile.ID}}/detail">
          <h3>状态区</h3>
          <div class="status-main">
            <span class="badge {{.State}}">{{statusLabel .State}}</span>
            <div class="endpoint">{{.Endpoint}}</div>
            <div class="health">
              <span>监听 {{if .Health.ListenerReady}}在线{{else}}离线{{end}}</span>
              <span>SSH {{if .Health.SSHConnected}}在线{{else}}离线{{end}}</span>
            </div>
            <div class="health">
              <span>上行 {{formatRate .Traffic.UploadBytesPerSec}}</span>
              <span>下行 {{formatRate .Traffic.DownloadBytesPerSec}}</span>
            </div>
            <div class="health">
              <span>累计上行 {{formatBytes .Traffic.UploadBytesTotal}}</span>
              <span>累计下行 {{formatBytes .Traffic.DownloadBytesTotal}}</span>
              <span>连接 {{.Traffic.ActiveConnections}}</span>
            </div>
            {{if .LastError}}<div class="error">{{.LastError}}</div>{{end}}
          </div>
        </section>
        <section class="zone">
          <h3>功能区</h3>
          <div class="actions">
            {{if or (eq .State "connected") (eq .State "connecting") (eq .State "retrying")}}
            <button class="btn stop" data-post-url="/profiles/{{.Profile.ID}}/stop" data-target="#cards-panel" data-detail-url="/profiles/{{.Profile.ID}}/detail">停止</button>
            {{else}}
            <button class="btn" data-post-url="/profiles/{{.Profile.ID}}/start" data-target="#cards-panel" data-detail-url="/profiles/{{.Profile.ID}}/detail">启动</button>
            {{end}}
            <button class="btn edit" data-load-url="/profiles/{{.Profile.ID}}/edit" data-target="#detail-panel">编辑</button>
            <button class="btn copy" data-post-url="/profiles/{{.Profile.ID}}/duplicate" data-target="#detail-panel">复制</button>
            <button class="btn delete" data-post-url="/profiles/{{.Profile.ID}}/delete" data-target="#cards-panel" data-detail-url="/profiles/new" data-confirm="确认删除这个 profile 吗？">删除</button>
          </div>
        </section>
      </article>
      {{end}}
    {{else}}
      <div class="placeholder">暂无 tunnel profile，点击右上角“新增”开始创建。</div>
    {{end}}
  </div>
</section>
{{end}}

{{define "detailPanel"}}
  {{if eq .Mode "form"}}
    <div class="detail-head">
      <div>
        <div class="detail-title">{{if .Profile.ID}}编辑配置{{else}}新增配置{{end}}</div>
        <div class="detail-sub">保存的是 profile 配置，不会持久化密码或私钥内容。</div>
      </div>
    </div>
    <form action="{{.FormAction}}" method="post" data-async-form="true" data-target="#detail-panel">
      <input type="hidden" name="id" value="{{.Profile.ID}}">
      <div class="form-grid">
        <div class="form-field"><label>标题</label><input name="title" value="{{.Profile.Title}}"></div>
        <div class="form-field"><label>SSH 用户</label><input name="ssh_user" value="{{.Profile.SSHUser}}"></div>
        <div class="form-field"><label>SSH 地址</label><input name="ssh_host" value="{{.Profile.SSHHost}}"></div>
        <div class="form-field"><label>SSH 端口</label><input name="ssh_port" value="{{.Profile.SSHPort}}"></div>
        <div class="form-field"><label>本地地址</label><input name="local_addr" value="{{.Profile.LocalAddr}}"></div>
        <div class="form-field"><label>本地端口</label><input name="local_port" value="{{.Profile.LocalPort}}"></div>
        <div class="form-field"><label>远程地址</label><input name="remote_host" value="{{.Profile.RemoteHost}}"></div>
        <div class="form-field"><label>远程端口</label><input name="remote_port" value="{{.Profile.RemotePort}}"></div>
        <div class="form-field full"><label>私钥路径</label><input name="ssh_key_path" value="{{.Profile.SSHKeyPath}}"></div>
        <div class="form-field"><label>超时秒数</label><input name="connect_timeout_sec" value="{{if .Profile.ConnectTimeoutSec}}{{.Profile.ConnectTimeoutSec}}{{else}}10{{end}}"></div>
        <div class="form-field"><label>Known Hosts 路径</label><input name="known_hosts_path" value="{{.Profile.KnownHostsPath}}"></div>
        <div class="form-field full"><label><input type="checkbox" name="insecure_ignore_host" {{if .Profile.InsecureIgnoreHost}}checked{{end}}> 忽略 Host Key 校验</label></div>
      </div>
      <div class="form-actions">
        <button type="button" class="btn ghost" data-load-url="/profiles/{{.Profile.ID}}/detail" data-target="#detail-panel">取消</button>
        <button type="submit" class="btn submit">保存</button>
      </div>
    </form>
  {{else if eq .Mode "detail"}}
    <div class="detail-head">
      <div>
        <div class="detail-title">{{.Profile.Title}}</div>
        <div class="detail-sub">SSH {{.Profile.SSHHost}}:{{.Profile.SSHPort}} · 代理 {{.Profile.LocalAddr}}:{{.Profile.LocalPort}} · 远程 {{.Profile.RemoteHost}}:{{.Profile.RemotePort}}</div>
      </div>
    </div>
    <div class="detail-grid">
      <section class="diagnostics">
        <h3>运行状态</h3>
        <div class="status-main">
          <span class="badge {{.View.State}}">{{statusLabel .View.State}}</span>
          <div class="endpoint">{{.View.Endpoint}}</div>
          <div class="health">
            <span>监听 {{if .View.Health.ListenerReady}}在线{{else}}离线{{end}}</span>
            <span>SSH {{if .View.Health.SSHConnected}}在线{{else}}离线{{end}}</span>
          </div>
          <div class="health">
            <span>当前上行 {{formatRate .View.Traffic.UploadBytesPerSec}}</span>
            <span>当前下行 {{formatRate .View.Traffic.DownloadBytesPerSec}}</span>
          </div>
          <div class="health">
            <span>累计上行 {{formatBytes .View.Traffic.UploadBytesTotal}}</span>
            <span>累计下行 {{formatBytes .View.Traffic.DownloadBytesTotal}}</span>
            <span>活跃连接 {{.View.Traffic.ActiveConnections}}</span>
          </div>
          {{if .View.LastError}}<div class="error">{{.View.LastError}}</div>{{end}}
        </div>
      </section>
      <section class="diagnostics" id="diagnostics-panel" data-poll-url="/profiles/{{.Profile.ID}}/diagnostics">
        {{template "diagnosticsPanel" .Diagnostics}}
      </section>
    </div>
  {{else}}
    <div class="placeholder">选择一个 tunnel card 查看详情，或点击右上角新增。</div>
  {{end}}
{{end}}

{{define "diagnosticsPanel"}}
  <h3>详情区</h3>
  <div class="health" style="margin-bottom:14px;">
    <span>启用时间 {{formatTime .View.Profile.LastEnabledAt}}</span>
    <span>当前入口 {{.View.Endpoint}}</span>
    <span>上行速度 {{formatRate .View.Traffic.UploadBytesPerSec}}</span>
    <span>下行速度 {{formatRate .View.Traffic.DownloadBytesPerSec}}</span>
    <span>累计上行 {{formatBytes .View.Traffic.UploadBytesTotal}}</span>
    <span>累计下行 {{formatBytes .View.Traffic.DownloadBytesTotal}}</span>
    <span>活跃连接 {{.View.Traffic.ActiveConnections}}</span>
  </div>
  <div class="logs">
    {{if .Logs}}
      {{range .Logs}}
      <div class="log">
        <time>{{.At.Local.Format "2006-01-02 15:04:05"}}</time>
        <div>{{.Message}}</div>
      </div>
      {{end}}
    {{else}}
      <div class="placeholder" style="padding:20px 0;">当前没有可显示的运行日志。</div>
    {{end}}
  </div>
{{end}}
`
