package pages

import (
	"fmt"
	"html/template"
	"strings"

	"timber-ssh-go/ui/i18n"
	"timber-ssh-go/ui/viewmodel"
)

func renderImportModal(data viewmodel.ImportModalData) string {
	t := i18n.New(data.Lang)
	var b strings.Builder
	b.WriteString(`<section class="modal-backdrop"><div class="modal-shell"><div class="modal-head"><div>`)
	b.WriteString(fmt.Sprintf(`<h2>%s</h2>`, template.HTMLEscapeString(t.T("import.title"))))
	if data.OverwriteNotice != "" {
		b.WriteString(fmt.Sprintf(`<p class="notice warn">%s</p>`, template.HTMLEscapeString(data.OverwriteNotice)))
	} else {
		b.WriteString(fmt.Sprintf(`<p class="notice">%s</p>`, template.HTMLEscapeString(t.T("import.notice"))))
	}
	b.WriteString(`</div>`)
	b.WriteString(fmt.Sprintf(`<a class="link" href="%s">%s</a></div>`, template.HTMLEscapeString(data.CancelURL), template.HTMLEscapeString(t.T("form.close"))))

	if !data.HasPreview {
		b.WriteString(fmt.Sprintf(`<form class="modal-body" action="%s" method="post" enctype="multipart/form-data">`, template.HTMLEscapeString(data.PreviewURL)))
		b.WriteString(fmt.Sprintf(`<input type="hidden" name="lang" value="%s">`, template.HTMLEscapeString(data.Lang)))
		b.WriteString(`<div class="grid"><div class="field full">`)
		b.WriteString(fmt.Sprintf(`<label>%s</label>`, template.HTMLEscapeString(t.T("import.file"))))
		b.WriteString(`<input type="file" name="file" accept="application/json,.json" required>`)
		b.WriteString(`</div></div><div class="actions-row">`)
		b.WriteString(fmt.Sprintf(`<a class="link" href="%s">%s</a>`, template.HTMLEscapeString(data.CancelURL), template.HTMLEscapeString(t.T("form.cancel"))))
		b.WriteString(fmt.Sprintf(`<button class="link primary" type="submit">%s</button>`, template.HTMLEscapeString(t.T("import.preview"))))
		b.WriteString(`</div></form></div></section>`)
		return b.String()
	}

	b.WriteString(fmt.Sprintf(`<form class="modal-body" action="%s" method="post">`, template.HTMLEscapeString(data.ConfirmURL)))
	b.WriteString(fmt.Sprintf(`<input type="hidden" name="lang" value="%s">`, template.HTMLEscapeString(data.Lang)))
	b.WriteString(fmt.Sprintf(`<input type="hidden" name="payload" value="%s">`, template.HTMLEscapeString(data.PayloadBase64)))
	b.WriteString(`<section class="import-summary">`)
	b.WriteString(renderImportSummaryCard(t.T("import.added"), data.AddedTitles, nil))
	b.WriteString(renderImportSummaryCard(t.T("import.conflicts"), data.ConflictTitles, nil))
	b.WriteString(renderImportSummaryCard(t.T("import.invalid"), nil, data.InvalidItems))
	b.WriteString(`</section><div class="actions-row start">`)
	b.WriteString(fmt.Sprintf(`<a class="link" href="%s">%s</a>`, template.HTMLEscapeString(data.CancelURL), template.HTMLEscapeString(t.T("form.cancel"))))
	if data.HasConflicts {
		b.WriteString(fmt.Sprintf(`<button class="link" type="submit" name="policy" value="skip">%s</button>`, template.HTMLEscapeString(t.T("import.skip_conflicts"))))
		b.WriteString(fmt.Sprintf(`<button class="link primary" type="submit" name="policy" value="merge">%s</button>`, template.HTMLEscapeString(t.T("import.merge_conflicts"))))
	} else {
		b.WriteString(fmt.Sprintf(`<button class="link primary" type="submit" name="policy" value="skip">%s</button>`, template.HTMLEscapeString(t.T("import.confirm"))))
	}
	b.WriteString(`</div></form></div></section>`)
	return b.String()
}

func renderImportSummaryCard(title string, items []string, invalid []viewmodel.ImportInvalidItem) string {
	var b strings.Builder
	b.WriteString(`<div class="card"><h3>`)
	b.WriteString(template.HTMLEscapeString(title))
	b.WriteString(`</h3><ul>`)
	if len(items) == 0 && len(invalid) == 0 {
		b.WriteString(`<li>0</li>`)
	}
	for _, item := range items {
		b.WriteString(`<li>`)
		b.WriteString(template.HTMLEscapeString(item))
		b.WriteString(`</li>`)
	}
	for _, item := range invalid {
		label := item.Reason
		if strings.TrimSpace(item.Title) != "" {
			label = item.Title + ": " + item.Reason
		}
		b.WriteString(`<li>`)
		b.WriteString(template.HTMLEscapeString(label))
		b.WriteString(`</li>`)
	}
	b.WriteString(`</ul></div>`)
	return b.String()
}
