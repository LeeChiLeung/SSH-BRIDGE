package pages

import (
	"context"
	"fmt"
	"html/template"
	"io"

	"github.com/a-h/templ"

	"timber-ssh-go/ui/i18n"
	"timber-ssh-go/ui/viewmodel"
)

func Form(data viewmodel.FormPageData) templ.Component {
	return templ.ComponentFunc(func(ctx context.Context, w io.Writer) error {
		_, err := io.WriteString(w, standaloneFormPage(data))
		return err
	})
}

func standaloneFormPage(data viewmodel.FormPageData) string {
	t := i18n.New(data.Lang)
	return fmt.Sprintf(`<!doctype html>
<html lang="%s"><head><meta charset="utf-8"><meta name="viewport" content="width=device-width, initial-scale=1">
<title>%s</title>%s<style>%s</style></head><body><main class="page">%s<footer class="footer">Version %s</footer></main></body></html>`,
		template.HTMLEscapeString(data.Lang),
		template.HTMLEscapeString(t.T("app.brand")),
		faviconLinks(),
		baseStyles(),
		renderFormModal(data),
		template.HTMLEscapeString(data.Version),
	)
}

func renderFormModal(data viewmodel.FormPageData) string {
	t := i18n.New(data.Lang)
	title := formTitle(data, t)
	checked := ""
	if data.InsecureIgnoreHost {
		checked = "checked"
	}
	autoStartChecked := ""
	if data.AutoStart {
		autoStartChecked = "checked"
	}
	return fmt.Sprintf(`<section class="modal-backdrop">
<div class="modal-shell">
  <div class="modal-head">
    <div>
      <h2>%s</h2>
      <p class="notice">%s</p>
    </div>
    <a class="link" href="%s" aria-label="%s">X</a>
  </div>
  <form class="modal-body" action="%s" method="post">
    <input type="hidden" name="id" value="%s">
    <input type="hidden" name="lang" value="%s">
    <input type="hidden" name="redirect_to" value="%s">
    <div class="grid">
      %s
      <div class="field full check-field"><label><input type="checkbox" name="auto_start" %s><span>%s</span></label></div>
      <div class="field full check-field"><label><input type="checkbox" name="insecure_ignore_host" %s><span>%s</span></label></div>
    </div>
    <div class="actions-row">
      <a class="link" href="%s">%s</a>
      <button class="link primary" type="submit">%s</button>
    </div>
  </form>
</div>
</section>`,
		template.HTMLEscapeString(title),
		template.HTMLEscapeString(t.T("form.password_notice")),
		template.HTMLEscapeString(data.CancelURL),
		template.HTMLEscapeString(t.T("form.close")),
		template.HTMLEscapeString(data.PostURL),
		template.HTMLEscapeString(data.ProfileID),
		template.HTMLEscapeString(data.Lang),
		template.HTMLEscapeString(data.ReturnURL),
		formFields(data, t),
		autoStartChecked,
		template.HTMLEscapeString(t.T("field.auto_start")),
		checked,
		template.HTMLEscapeString(t.T("field.ignore_host_key")),
		template.HTMLEscapeString(data.CancelURL),
		template.HTMLEscapeString(t.T("form.cancel")),
		template.HTMLEscapeString(t.T("form.save")),
	)
}

func formTitle(data viewmodel.FormPageData, t i18n.Translator) string {
	if data.IsEdit {
		return t.T("form.edit_title")
	}
	return t.T("form.new_title")
}

func formFields(data viewmodel.FormPageData, t i18n.Translator) string {
	return fmt.Sprintf(`
<div class="field"><label>%s</label><input name="title" value="%s"></div>
<div class="field"><label>%s</label><input name="ssh_user" value="%s"></div>
<div class="field"><label>%s</label><input name="ssh_host" value="%s"></div>
<div class="field"><label>%s</label><input name="ssh_port" value="%d"></div>
<div class="field"><label>%s</label><input type="password" name="ssh_password" value="%s"></div>
<div class="field"><label>%s</label><input name="ssh_key_path" value="%s" placeholder="%s"></div>
<div class="field"><label>%s</label><input name="local_addr" value="%s"></div>
<div class="field"><label>%s</label><input name="local_port" value="%d"></div>
<div class="field"><label>%s</label><input name="remote_host" value="%s"></div>
<div class="field"><label>%s</label><input name="remote_port" value="%d"></div>
<div class="field"><label>%s</label><input name="connect_timeout_sec" value="%d"></div>
<div class="field"><label>%s</label><input name="known_hosts_path" value="%s"></div>`,
		template.HTMLEscapeString(t.T("field.title")), template.HTMLEscapeString(data.ProfileTitle),
		template.HTMLEscapeString(t.T("field.ssh_user")), template.HTMLEscapeString(data.SSHUser),
		template.HTMLEscapeString(t.T("field.ssh_host")), template.HTMLEscapeString(data.SSHHost),
		template.HTMLEscapeString(t.T("field.ssh_port")), data.SSHPort,
		template.HTMLEscapeString(t.T("field.ssh_password")), template.HTMLEscapeString(data.SSHPassword),
		template.HTMLEscapeString(t.T("field.ssh_key")), template.HTMLEscapeString(data.SSHKeyPath), template.HTMLEscapeString(t.T("placeholder.key")),
		template.HTMLEscapeString(t.T("field.local_addr")), template.HTMLEscapeString(data.LocalAddr),
		template.HTMLEscapeString(t.T("field.local_port")), data.LocalPort,
		template.HTMLEscapeString(t.T("field.remote_host")), template.HTMLEscapeString(data.RemoteHost),
		template.HTMLEscapeString(t.T("field.remote_port")), data.RemotePort,
		template.HTMLEscapeString(t.T("field.timeout")), data.ConnectTimeoutSec,
		template.HTMLEscapeString(t.T("field.known_hosts")), template.HTMLEscapeString(data.KnownHostsPath),
	)
}
