package components

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

func ProfileCard(card viewmodel.ProfileCard) templ.Component {
	return templ.ComponentFunc(func(ctx context.Context, w io.Writer) error {
		t := i18n.New(card.Lang)
		_, err := fmt.Fprintf(w, `
<article class="card">
  <div class="meta">
    <span class="state">%s</span>
    <h3>%s</h3>
  </div>
  <dl>
    <div><dt>%s</dt><dd>%s</dd></div>
    <div><dt>%s</dt><dd>%s</dd></div>
    <div><dt>%s</dt><dd>%s</dd></div>
    <div><dt>%s</dt><dd>%s</dd></div>
  </dl>
  <div class="card-actions">%s</div>
</article>`,
			template.HTMLEscapeString(card.StateLabel),
			template.HTMLEscapeString(card.Title),
			template.HTMLEscapeString(t.T("field.ssh")),
			template.HTMLEscapeString(card.SSHAddress),
			template.HTMLEscapeString(t.T("field.local")),
			template.HTMLEscapeString(card.LocalEndpoint),
			template.HTMLEscapeString(t.T("field.remote")),
			template.HTMLEscapeString(card.RemoteTarget),
			template.HTMLEscapeString(t.T("field.enabled_at")),
			template.HTMLEscapeString(card.LastEnabledText),
			cardActions(card, t),
		)
		return err
	})
}

func cardActions(card viewmodel.ProfileCard, t i18n.Translator) string {
	return strings.Join([]string{
		fmt.Sprintf(`<a class="link" href="%s">%s</a>`,
			template.HTMLEscapeString(card.DetailURL),
			template.HTMLEscapeString(t.T("home.view")),
		),
		actionButton(card.StartURL, card.RedirectURL, card.Lang, t.T("detail.start"), "link primary", card.CanStart),
		actionButton(card.StopURL, card.RedirectURL, card.Lang, t.T("detail.stop"), "link", card.CanStop),
	}, " ")
}

func actionButton(actionURL, redirectURL, lang, label, className string, enabled bool) string {
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
