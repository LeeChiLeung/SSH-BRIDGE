package pages

import (
	"fmt"
	"html/template"

	"timber-ssh-go/ui/i18n"
)

func baseStyles() string {
	return `
    :root { color-scheme: light; --bg:#f6f8fb; --card:#ffffff; --line:#dbe3ef; --text:#172033; --muted:#64748b; --accent:#1d4ed8; }
    * { box-sizing:border-box; }
    body { margin:0; font-family:"Noto Sans SC","PingFang SC","Microsoft YaHei",sans-serif; color:var(--text); background:linear-gradient(180deg,#f8fbff 0%,#edf3fb 100%); }
    .page { width:min(1500px, calc(100vw - 32px)); margin:24px auto 40px; }
    .hero { display:flex; justify-content:space-between; align-items:flex-start; gap:16px; background:var(--card); border:1px solid var(--line); border-radius:24px; padding:24px 28px; box-shadow:0 12px 30px rgba(15,23,42,.06); }
    .hero h1 { margin:0 0 8px; font-size:32px; }
    .hero p { margin:0; color:var(--muted); line-height:1.6; }
    .actions { display:flex; align-items:center; gap:12px; flex-wrap:wrap; }
    .link { display:inline-flex; align-items:center; justify-content:center; min-height:44px; padding:0 16px; border-radius:14px; text-decoration:none; border:1px solid var(--line); color:var(--text); background:#fff; cursor:pointer; font:inherit; }
    button.link { appearance:none; }
    .primary { background:#111827; color:#fff; border-color:#111827; }
    .danger { background:#fff1f2; color:#be123c; border-color:#fecdd3; }
    .cards { margin-top:22px; display:grid; grid-template-columns:repeat(auto-fit, minmax(260px, 1fr)); gap:18px; }
    .card { background:var(--card); border:1px solid var(--line); border-radius:20px; padding:18px; box-shadow:0 10px 25px rgba(15,23,42,.04); }
    .detail-card { margin-top:22px; }
    .meta { display:flex; justify-content:space-between; align-items:center; gap:12px; margin-bottom:14px; }
    .meta h3 { margin:0; font-size:20px; }
    .state { display:inline-flex; padding:6px 10px; border-radius:999px; background:#dbeafe; color:#1d4ed8; font-size:12px; font-weight:700; }
    dl { margin:0; display:grid; gap:12px; }
    dt { font-size:12px; color:var(--muted); margin-bottom:4px; }
    dd { margin:0; font-size:15px; word-break:break-all; }
    .footer { margin-top:18px; text-align:right; color:var(--muted); font-size:13px; }
    .empty { margin-top:22px; background:var(--card); border:1px dashed var(--line); border-radius:20px; padding:28px; color:var(--muted); text-align:center; }
    .lang-switch { display:inline-flex; align-items:center; gap:8px; color:var(--muted); font-size:13px; }
    .panel-grid { margin-top:22px; display:grid; grid-template-columns:repeat(2,minmax(0,1fr)); gap:18px; }
    .list-shell { margin-top:22px; background:var(--card); border:1px solid var(--line); border-radius:24px; overflow:hidden; box-shadow:0 10px 25px rgba(15,23,42,.04); }
    .list-head, .row-main { display:grid; grid-template-columns:120px minmax(180px,1.3fr) minmax(220px,1fr) minmax(180px,1fr) minmax(180px,1fr) 190px; gap:18px; align-items:center; }
    .list-head { display:none; }
    .list-body { display:grid; max-height:252px; overflow-y:auto; }
    .list-row { display:grid; grid-template-columns:minmax(0,1fr) auto; gap:16px; align-items:center; min-height:84px; padding:16px 20px; border-top:1px solid var(--line); }
    .list-row:first-child { border-top:none; }
    .list-row.selected { background:#f8fbff; }
    .row-main { color:var(--text); text-decoration:none; min-width:0; }
    .row-main span { min-width:0; }
    .row-title { font-weight:700; }
    .row-actions { display:flex; gap:10px; align-items:center; flex-wrap:wrap; justify-content:flex-end; }
    .detail-shell { margin-top:22px; }
    .section-head { display:flex; align-items:center; justify-content:space-between; gap:16px; margin-bottom:14px; }
    .section-head h2 { margin:0; font-size:24px; }
    .workbench-card { padding:22px; }
    .workbench-head { display:flex; align-items:flex-start; justify-content:space-between; gap:18px; padding-bottom:18px; border-bottom:1px solid var(--line); }
    .workbench-title h3 { margin:0 0 8px; font-size:24px; }
    .workbench-title p { margin:0; color:var(--muted); }
    .workbench-grid { display:grid; grid-template-columns:minmax(240px,.9fr) minmax(320px,1fr) minmax(320px,1fr); gap:20px; margin-top:20px; }
    .workbench-pane { min-width:0; padding:0 4px; }
    .workbench-pane + .workbench-pane { border-left:1px solid var(--line); padding-left:20px; }
    .logs-pane { max-height:360px; overflow:auto; padding-right:8px; }
    .summary-grid { display:grid; gap:16px; }
    .banner { margin-top:22px; background:#fff7ed; border:1px solid #fdba74; border-radius:16px; padding:14px 16px; color:#9a3412; }
    .diagnostics-list { display:grid; gap:12px; }
    .diagnostics-list div { padding-bottom:12px; border-bottom:1px solid var(--line); }
    .diagnostics-list div:last-child { border-bottom:none; padding-bottom:0; }
    .diagnostics-grid { display:grid; grid-template-columns:repeat(2, minmax(0,1fr)); gap:12px 18px; }
    .diagnostics-item { padding-bottom:12px; border-bottom:1px solid var(--line); }
    .logs { margin:0; padding-left:18px; display:grid; gap:10px; color:var(--muted); }
    .logs strong { color:var(--text); }
    .log-empty { margin:0; color:var(--muted); }
    .inline-form { margin:0; }
    .card-actions { margin-top:14px; display:flex; flex-wrap:wrap; gap:10px; }
    .toolbar { display:flex; align-items:center; gap:10px; flex-wrap:wrap; }
    .toggle { color:var(--muted); font-size:13px; display:inline-flex; align-items:center; gap:6px; }
    .grid { display:grid; grid-template-columns:repeat(2,minmax(0,1fr)); gap:14px 18px; }
    .full { grid-column:1/-1; }
    .field label { display:block; margin-bottom:6px; color:var(--muted); font-size:13px; }
    .field input { width:100%; border:1px solid var(--line); border-radius:12px; padding:11px 12px; font-size:14px; }
    .notice { margin:0; color:var(--muted); line-height:1.6; }
    .notice.warn { color:#9a3412; }
    .actions-row { display:flex; justify-content:flex-end; gap:12px; margin-top:18px; }
    .actions-row.start { justify-content:flex-start; }
    .modal-backdrop { position:fixed; inset:0; background:rgba(15,23,42,.42); display:flex; align-items:flex-start; justify-content:center; padding:40px 18px; overflow:auto; backdrop-filter:blur(6px); }
    .modal-shell { width:min(920px, 100%); background:var(--card); border:1px solid var(--line); border-radius:28px; box-shadow:0 24px 60px rgba(15,23,42,.22); }
    .modal-head { display:flex; justify-content:space-between; align-items:flex-start; gap:16px; padding:24px 26px 0; }
    .modal-head h2 { margin:0 0 8px; font-size:28px; }
    .modal-body { padding:22px 26px 26px; }
    .import-summary { display:grid; grid-template-columns:repeat(3,minmax(0,1fr)); gap:14px; margin-top:18px; }
    .import-summary .card { padding:16px; border-radius:18px; }
    .import-summary h3 { margin:0 0 10px; font-size:16px; }
    .import-summary ul { margin:0; padding-left:18px; color:var(--muted); display:grid; gap:8px; }
    .import-summary li { word-break:break-word; }
    .import-result { margin-top:22px; background:#ecfeff; border:1px solid #a5f3fc; border-radius:18px; padding:16px 18px; color:#155e75; }
    @media (max-width: 1350px) { .list-head, .row-main { grid-template-columns:100px minmax(150px,1fr) minmax(180px,1fr) minmax(160px,1fr) minmax(160px,1fr) 170px; } .workbench-grid { grid-template-columns:minmax(220px,.9fr) minmax(280px,1fr) minmax(280px,1fr); } }
    @media (max-width: 900px) { .hero { flex-direction:column; } .page { width:min(100vw - 20px, 1280px); } }
    @media (max-width: 900px) { .panel-grid { grid-template-columns:1fr; } .grid { grid-template-columns:1fr; } .modal-backdrop { padding:18px 10px; } .modal-head { padding:18px 18px 0; flex-direction:column; } .modal-body { padding:18px; } .list-row { grid-template-columns:1fr; } .row-main { grid-template-columns:1fr; gap:8px; } .row-actions { justify-content:flex-start; } .section-head { flex-direction:column; align-items:flex-start; } .workbench-head { flex-direction:column; } .workbench-grid { grid-template-columns:1fr; } .workbench-pane { padding:0; border-left:none !important; } .diagnostics-grid { grid-template-columns:1fr; } .import-summary { grid-template-columns:1fr; } }
`
}

func languageSwitcher(basePath string, current string, t i18n.Translator) string {
	zhClass := "link"
	enClass := "link"
	if i18n.Normalize(current) == i18n.LangZH {
		zhClass = "link primary"
	} else {
		enClass = "link primary"
	}
	return fmt.Sprintf(
		`<span class="lang-switch">%s</span><a class="%s" href="%s">中文</a><a class="%s" href="%s">EN</a>`,
		template.HTMLEscapeString(t.T("lang.switch")),
		zhClass,
		template.HTMLEscapeString(withLangPage(basePath, i18n.LangZH)),
		enClass,
		template.HTMLEscapeString(withLangPage(basePath, i18n.LangEN)),
	)
}

func withLangPage(path, lang string) string {
	if path == "" {
		path = "/"
	}
	return fmt.Sprintf("%s?lang=%s", path, lang)
}
