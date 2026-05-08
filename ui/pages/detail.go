package pages

import (
	"context"
	"fmt"
	"html/template"
	"io"
	"strings"

	"github.com/a-h/templ"

	"timber-ssh-go/ui/i18n"
	"timber-ssh-go/ui/viewmodel"
)

func Detail(data viewmodel.DetailPageData) templ.Component {
	return templ.ComponentFunc(func(ctx context.Context, w io.Writer) error {
		t := i18n.New(data.Lang)
		errorBanner := ""
		if data.ErrorMessage != "" {
			errorBanner = fmt.Sprintf(`<section class="banner">%s</section>`, template.HTMLEscapeString(data.ErrorMessage))
		}
		_, err := fmt.Fprintf(w, `<!doctype html>
<html lang="%s"><head><meta charset="utf-8"><meta name="viewport" content="width=device-width, initial-scale=1">
<title>%s</title><style>%s</style></head><body><main class="page">
<section class="hero"><div><h1>%s</h1><p>%s</p></div>
<div class="actions">%s <a class="link primary" href="%s">%s</a></div></section>
%s
<section class="card detail-card"><dl>
<div><dt>%s</dt><dd>%s</dd></div>
<div><dt>%s</dt><dd>%s</dd></div>
<div><dt>%s</dt><dd>%s</dd></div>
<div><dt>%s</dt><dd>%s</dd></div>
</dl></section>
<section id="diagnostics-root">%s</section>
%s
<script>%s</script>
<footer class="footer">Version %s</footer></main></body></html>`,
			template.HTMLEscapeString(data.Lang),
			template.HTMLEscapeString(data.Title),
			baseStyles(),
			template.HTMLEscapeString(data.Card.Title),
			template.HTMLEscapeString(t.T("detail.title")),
			detailActions(data, t),
			template.HTMLEscapeString(data.HomeURL),
			template.HTMLEscapeString(t.T("detail.back")),
			errorBanner,
			template.HTMLEscapeString(t.T("field.state")),
			template.HTMLEscapeString(data.Card.StateLabel),
			template.HTMLEscapeString(t.T("field.ssh")),
			template.HTMLEscapeString(data.Card.SSHAddress),
			template.HTMLEscapeString(t.T("field.local")),
			template.HTMLEscapeString(data.Card.LocalEndpoint),
			template.HTMLEscapeString(t.T("field.remote")),
			template.HTMLEscapeString(data.Card.RemoteTarget),
			renderDiagnosticsMarkup(data, t),
			renderOptionalFormModal(data.FormModal),
			detailRefreshScript(data.Diagnostics.RefreshURL),
			template.HTMLEscapeString(data.Version),
		)
		return err
	})
}

func DiagnosticsPanel(data viewmodel.DetailPageData) templ.Component {
	return templ.ComponentFunc(func(ctx context.Context, w io.Writer) error {
		t := i18n.New(data.Lang)
		_, err := io.WriteString(w, renderDiagnosticsMarkup(data, t))
		return err
	})
}

func detailActions(data viewmodel.DetailPageData, t i18n.Translator) string {
	actions := []string{
		fmt.Sprintf(`<a class="link" href="%s">%s</a>`, template.HTMLEscapeString(data.EditURL), template.HTMLEscapeString(t.T("detail.edit"))),
		actionForm(data.StartURL, data.Lang, t.T("detail.start"), "link primary", data.CanStart),
		actionForm(data.StopURL, data.Lang, t.T("detail.stop"), "link", data.CanStop),
		actionForm(data.DeleteURL, data.Lang, t.T("detail.delete"), "link danger", data.CanStart),
	}
	if data.LegacyURL != "" {
		actions = append(actions, fmt.Sprintf(`<a class="link" href="%s">%s</a>`, template.HTMLEscapeString(data.LegacyURL), template.HTMLEscapeString(t.T("home.open_legacy"))))
	}
	return strings.Join(actions, " ")
}

func actionForm(actionURL, lang, label, className string, enabled bool) string {
	disabled := ""
	if !enabled {
		disabled = ` disabled style="opacity:.45;cursor:not-allowed;"`
	}
	return fmt.Sprintf(
		`<form class="inline-form" action="%s" method="post"><input type="hidden" name="lang" value="%s"><button class="%s" type="submit"%s>%s</button></form>`,
		template.HTMLEscapeString(actionURL),
		template.HTMLEscapeString(lang),
		template.HTMLEscapeString(className),
		disabled,
		template.HTMLEscapeString(label),
	)
}

func renderLogs(logs []viewmodel.DiagnosticsLog, t i18n.Translator) string {
	if len(logs) == 0 {
		return fmt.Sprintf(`<p class="log-empty">%s</p>`, template.HTMLEscapeString(t.T("detail.no_logs")))
	}
	var b strings.Builder
	b.WriteString(`<ol class="logs">`)
	for _, entry := range logs {
		fmt.Fprintf(&b, `<li><strong>%s</strong><br>%s</li>`,
			template.HTMLEscapeString(entry.TimeText),
			template.HTMLEscapeString(entry.Message),
		)
	}
	b.WriteString(`</ol>`)
	return b.String()
}

func renderDiagnosticsMarkup(data viewmodel.DetailPageData, t i18n.Translator) string {
	return fmt.Sprintf(`<section class="workbench-grid">
<article class="workbench-pane">
  <div class="meta"><h3>%s</h3></div>
  <dl class="summary-grid">
    <div><dt>%s</dt><dd>%s</dd></div>
    <div><dt>%s</dt><dd>%s</dd></div>
    <div><dt>%s</dt><dd>%s</dd></div>
    <div><dt>%s</dt><dd>%s</dd></div>
  </dl>
</article>
<article class="workbench-pane">
  <div class="meta"><h3>%s</h3><div class="toolbar"><button type="button" class="link" data-refresh-now>%s</button><label class="toggle"><input type="checkbox" data-auto-refresh> %s</label></div></div>
  %s
</article>
<article class="workbench-pane logs-pane">
  <div class="meta"><h3>%s</h3></div>
  %s
</article>
</section>`,
		template.HTMLEscapeString(t.T("detail.summary")),
		template.HTMLEscapeString(t.T("field.state")),
		template.HTMLEscapeString(data.Card.StateLabel),
		template.HTMLEscapeString(t.T("field.ssh")),
		template.HTMLEscapeString(data.Card.SSHAddress),
		template.HTMLEscapeString(t.T("field.local")),
		template.HTMLEscapeString(data.Card.LocalEndpoint),
		template.HTMLEscapeString(t.T("field.remote")),
		template.HTMLEscapeString(data.Card.RemoteTarget),
		template.HTMLEscapeString(t.T("detail.diagnostics")),
		template.HTMLEscapeString(t.T("detail.refresh_now")),
		template.HTMLEscapeString(t.T("detail.auto_refresh")),
		renderDiagnosticsSummary(data.Diagnostics, t),
		template.HTMLEscapeString(t.T("detail.logs")),
		renderLogs(data.Diagnostics.Logs, t),
	)
}

func renderDiagnosticsSummary(data viewmodel.DiagnosticsPanel, t i18n.Translator) string {
	rows := [][2]string{
		{t.T("detail.listener_ready"), boolText(data.ListenerReady, t)},
		{t.T("detail.ssh_connected"), boolText(data.SSHConnected, t)},
		{t.T("detail.active_connections"), data.ActiveConnections},
		{t.T("detail.upload_rate"), data.UploadRateText},
		{t.T("detail.download_rate"), data.DownloadRateText},
		{t.T("detail.upload_total"), data.UploadTotalText},
		{t.T("detail.download_total"), data.DownloadTotalText},
		{t.T("detail.last_error"), orFallback(data.LastError, t.T("detail.none"))},
	}
	var b strings.Builder
	b.WriteString(`<div class="diagnostics-grid">`)
	for _, row := range rows {
		fmt.Fprintf(&b, `<div class="diagnostics-item"><dt>%s</dt><dd>%s</dd></div>`,
			template.HTMLEscapeString(row[0]),
			template.HTMLEscapeString(row[1]),
		)
	}
	b.WriteString(`</div>`)
	return b.String()
}

func detailRefreshScript(refreshURL string) string {
	return fmt.Sprintf(`(function(){
  const root = document.getElementById('diagnostics-root');
  if (!root) return;
  let timer = null;
  let autoEnabled = false;
  const bind = function() {
    const toggle = root.querySelector('[data-auto-refresh]');
    const refreshBtn = root.querySelector('[data-refresh-now]');
    if (toggle) {
      toggle.checked = autoEnabled;
    }
    const refresh = async function() {
      try {
        const response = await fetch(%q, { headers: { 'X-Requested-With': 'fetch' } });
        if (!response.ok) return;
        root.innerHTML = await response.text();
        bind();
      } catch (_) {}
    };
    if (refreshBtn) {
      refreshBtn.addEventListener('click', refresh);
    }
    if (toggle) {
      toggle.addEventListener('change', function() {
        autoEnabled = toggle.checked;
        if (timer) {
          clearInterval(timer);
          timer = null;
        }
        if (autoEnabled) {
          refresh();
          timer = window.setInterval(refresh, 3000);
        }
      });
    }
  };
  bind();
})();`, refreshURL)
}

func renderOptionalFormModal(data *viewmodel.FormPageData) string {
	if data == nil {
		return ""
	}
	return renderFormModal(*data)
}

func orFallback(value, fallback string) string {
	if value == "" {
		return fallback
	}
	return value
}

func boolText(value bool, t i18n.Translator) string {
	if value {
		return t.T("detail.yes")
	}
	return t.T("detail.no")
}
