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

func Home(data viewmodel.HomePageData) templ.Component {
	return templ.ComponentFunc(func(ctx context.Context, w io.Writer) error {
		t := i18n.New(data.Lang)
		_, err := fmt.Fprintf(w, `<!doctype html>
<html lang="%s">
<head>
  <meta charset="utf-8">
  <meta name="viewport" content="width=device-width, initial-scale=1">
  <title>%s</title>
  <style>%s</style>
</head>
<body>
  <main class="page">
    <section class="hero">
      <div><h1>%s</h1></div>
      <div class="actions">
        %s
        <a class="link" href="%s">%s</a>
        <a class="link" href="%s">%s</a>
        <a class="link" href="%s">%s</a>
        <a class="link primary" href="%s">%s</a>
      </div>
    </section>`,
			template.HTMLEscapeString(data.Lang),
			template.HTMLEscapeString(data.Title),
			baseStyles(),
			template.HTMLEscapeString(data.Title),
			languageSwitcher(data.HomeURL, data.Lang, t),
			template.HTMLEscapeString(data.ExportURL),
			template.HTMLEscapeString(t.T("home.export")),
			template.HTMLEscapeString(data.ImportURL),
			template.HTMLEscapeString(t.T("home.import")),
			template.HTMLEscapeString(data.NewURL),
			template.HTMLEscapeString(t.T("home.new")),
			withLangPage(data.HomeURL, data.Lang),
			template.HTMLEscapeString(t.T("home.refresh")),
		)
		if err != nil {
			return err
		}

		if data.ImportResult != nil {
			if _, err = io.WriteString(w, renderImportResult(*data.ImportResult, t)); err != nil {
				return err
			}
		}

		if len(data.Cards) == 0 {
			_, err = fmt.Fprintf(w, `<section class="empty">%s</section>`, template.HTMLEscapeString(t.T("home.empty")))
			if err != nil {
				return err
			}
		} else {
			if _, err = io.WriteString(w, renderListSection(data, t)); err != nil {
				return err
			}
			if data.SelectedDetail != nil {
				if _, err = io.WriteString(w, renderEmbeddedDetail(*data.SelectedDetail, t)); err != nil {
					return err
				}
				if _, err = fmt.Fprintf(w, `<script>%s</script>`, detailRefreshScript(data.SelectedDetail.Diagnostics.RefreshURL)); err != nil {
					return err
				}
			}
		}

		if data.FormModal != nil {
			if _, err = io.WriteString(w, renderFormModal(*data.FormModal)); err != nil {
				return err
			}
		}
		if data.ImportModal != nil {
			if _, err = io.WriteString(w, renderImportModal(*data.ImportModal)); err != nil {
				return err
			}
		}

		_, err = fmt.Fprintf(w, `<footer class="footer">Version %s</footer></main></body></html>`, template.HTMLEscapeString(data.Version))
		return err
	})
}

func renderImportResult(data viewmodel.ImportResultData, t i18n.Translator) string {
	return fmt.Sprintf(`<section class="import-result"><strong>%s</strong> %s %d，%s %d，%s %d，%s %d。</section>`,
		template.HTMLEscapeString(t.T("import.result_title")),
		template.HTMLEscapeString(t.T("import.added")),
		data.Added,
		template.HTMLEscapeString(t.T("import.merged")),
		data.Merged,
		template.HTMLEscapeString(t.T("import.skipped")),
		data.Skipped,
		template.HTMLEscapeString(t.T("import.invalid")),
		data.Invalid,
	)
}

func renderListSection(data viewmodel.HomePageData, t i18n.Translator) string {
	var b strings.Builder
	b.WriteString(`<section class="list-shell"><div class="list-head">`)
	b.WriteString(fmt.Sprintf(`<span>%s</span><span>%s</span><span>%s</span><span>%s</span><span>%s</span><span>%s</span><span class="align-right">%s</span>`,
		template.HTMLEscapeString(t.T("field.state")),
		template.HTMLEscapeString(t.T("field.title")),
		template.HTMLEscapeString(t.T("field.ssh")),
		template.HTMLEscapeString(t.T("field.local")),
		template.HTMLEscapeString(t.T("field.remote")),
		template.HTMLEscapeString(t.T("field.enabled_at")),
		template.HTMLEscapeString(t.T("home.actions")),
	))
	b.WriteString(`</div><div class="list-body">`)
	for _, row := range data.Cards {
		b.WriteString(renderListRow(row, t))
	}
	b.WriteString(`</div></section>`)
	return b.String()
}

func renderListRow(row viewmodel.ProfileCard, t i18n.Translator) string {
	rowClass := "list-row"
	if row.Selected {
		rowClass += " selected"
	}
	return fmt.Sprintf(`<article class="%s">
<a class="row-main" href="%s">
  <span><span class="state">%s</span></span>
  <span class="row-title">%s</span>
  <span>%s</span>
  <span>%s</span>
  <span>%s</span>
  <span>%s</span>
</a>
<div class="row-actions">
  %s
  %s
  %s
</div>
</article>`,
		rowClass,
		template.HTMLEscapeString(row.DetailURL),
		template.HTMLEscapeString(row.StateLabel),
		template.HTMLEscapeString(row.Title),
		template.HTMLEscapeString(row.SSHAddress),
		template.HTMLEscapeString(row.LocalEndpoint),
		template.HTMLEscapeString(row.RemoteTarget),
		template.HTMLEscapeString(row.LastEnabledText),
		fmt.Sprintf(`<a class="link" href="%s">%s</a>`, template.HTMLEscapeString(row.DetailURL), template.HTMLEscapeString(t.T("home.view"))),
		rowActionButton(row.StartURL, row.RedirectURL, row.Lang, t.T("detail.start"), "link primary", row.CanStart),
		rowActionButton(row.StopURL, row.RedirectURL, row.Lang, t.T("detail.stop"), "link", row.CanStop),
	)
}

func rowActionButton(actionURL, redirectURL, lang, label, className string, enabled bool) string {
	disabled := ""
	if !enabled {
		disabled = ` disabled style="opacity:.45;cursor:not-allowed;"`
	}
	return fmt.Sprintf(
		`<form class="inline-form" action="%s" method="post"><input type="hidden" name="lang" value="%s"><input type="hidden" name="redirect_to" value="%s"><button class="%s" type="submit"%s>%s</button></form>`,
		template.HTMLEscapeString(actionURL),
		template.HTMLEscapeString(lang),
		template.HTMLEscapeString(redirectURL),
		template.HTMLEscapeString(className),
		disabled,
		template.HTMLEscapeString(label),
	)
}

func renderEmbeddedDetail(data viewmodel.DetailPageData, t i18n.Translator) string {
	return fmt.Sprintf(`<section class="detail-shell">
<div class="section-head"><h2>%s</h2></div>
<section class="card workbench-card">
<div class="workbench-head">
  <div class="workbench-title">
    <h3>%s</h3>
  </div>
  <div class="actions">%s</div>
</div>
<div id="diagnostics-root">%s</div>
</section>
</section>`,
		template.HTMLEscapeString(t.T("home.detail_panel")),
		template.HTMLEscapeString(data.Card.Title),
		detailActions(data, t),
		renderDiagnosticsMarkup(data, t),
	)
}
