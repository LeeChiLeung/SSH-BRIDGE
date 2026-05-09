package ui

import (
	"encoding/base64"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"

	"timber-ssh-go/ui/i18n"
	"timber-ssh-go/ui/pages"
)

const localeCookieName = "timber_lang"

func (a *App) handleHome(w http.ResponseWriter, r *http.Request) {
	lang := resolvedLanguage(r)
	persistLanguagePreference(w, lang)
	data := a.deps.HomeData(lang)
	data.CurrentURL = currentRequestURL(r, lang)
	selectedID := r.URL.Query().Get("selected")
	if selectedID == "" && len(data.Cards) > 0 {
		selectedID = profileIDFromDetailURL(data.Cards[0].DetailURL)
	}
	for i := range data.Cards {
		if profileIDFromDetailURL(data.Cards[i].DetailURL) == selectedID {
			data.Cards[i].Selected = true
		}
	}
	if selectedID != "" {
		if detail, ok := a.deps.DetailData(selectedID, lang); ok {
			returnURL := selectedHomeURL(lang, selectedID)
			detail.EditURL = appendRedirect(detail.EditURL, returnURL)
			detail.CurrentURL = currentRequestURL(r, lang)
			data.SelectedDetail = &detail
		}
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := pages.Home(data).Render(r.Context(), w); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (a *App) handleDetail(w http.ResponseWriter, r *http.Request) {
	lang := resolvedLanguage(r)
	persistLanguagePreference(w, lang)
	data, ok := a.deps.DetailData(chi.URLParam(r, "id"), lang)
	if !ok {
		http.NotFound(w, r)
		return
	}
	data.CurrentURL = currentRequestURL(r, lang)
	data.ErrorMessage = r.URL.Query().Get("error")
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := pages.Detail(data).Render(r.Context(), w); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (a *App) handleDiagnostics(w http.ResponseWriter, r *http.Request) {
	lang := resolvedLanguage(r)
	persistLanguagePreference(w, lang)
	data, ok := a.deps.DetailData(chi.URLParam(r, "id"), lang)
	if !ok {
		http.NotFound(w, r)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := pages.DiagnosticsPanel(data).Render(r.Context(), w); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (a *App) handleNewProfile(w http.ResponseWriter, r *http.Request) {
	lang := resolvedLanguage(r)
	persistLanguagePreference(w, lang)
	data := a.deps.HomeData(lang)
	data.CurrentURL = currentRequestURL(r, lang)
	form := a.deps.NewFormData(lang)
	form.CancelURL = withLang("/", lang)
	form.ReturnURL = withLang("/", lang)
	data.FormModal = &form
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := pages.Home(data).Render(r.Context(), w); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (a *App) handleImportProfile(w http.ResponseWriter, r *http.Request) {
	lang := resolvedLanguage(r)
	persistLanguagePreference(w, lang)
	data := a.deps.HomeData(lang)
	data.CurrentURL = currentRequestURL(r, lang)
	modal := a.deps.NewImportData(lang)
	data.ImportModal = &modal
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := pages.Home(data).Render(r.Context(), w); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (a *App) handleExportProfiles(w http.ResponseWriter, r *http.Request) {
	filename, data, err := a.deps.ExportProfiles()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Header().Set("Content-Disposition", fmt.Sprintf(`attachment; filename="%s"`, filename))
	_, _ = w.Write(data)
}

func (a *App) handleImportPreview(w http.ResponseWriter, r *http.Request) {
	lang := requestLang(r)
	persistLanguagePreference(w, lang)
	fileData, err := readUploadedImportFile(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	modal, err := a.deps.PreviewImport(fileData)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	modal.Lang = lang
	if modal.OverwriteNotice == "" && modal.HasConflicts {
		modal.OverwriteNotice = i18n.New(lang).T("import.overwrite_notice")
	}
	data := a.deps.HomeData(lang)
	data.CurrentURL = currentRequestURL(r, lang)
	data.ImportModal = &modal
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := pages.Home(data).Render(r.Context(), w); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (a *App) handleImportConfirm(w http.ResponseWriter, r *http.Request) {
	lang := requestLang(r)
	persistLanguagePreference(w, lang)
	if err := r.ParseForm(); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	payload, err := decodeImportPayload(r.FormValue("payload"))
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	result, err := a.deps.ApplyImport(payload, r.FormValue("policy"))
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	data := a.deps.HomeData(lang)
	data.CurrentURL = currentRequestURL(r, lang)
	data.ImportResult = &result
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := pages.Home(data).Render(r.Context(), w); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (a *App) handleEditProfile(w http.ResponseWriter, r *http.Request) {
	lang := resolvedLanguage(r)
	persistLanguagePreference(w, lang)
	profileID := chi.URLParam(r, "id")
	formData, ok := a.deps.EditFormData(chi.URLParam(r, "id"), lang)
	if !ok {
		http.NotFound(w, r)
		return
	}
	returnURL := r.URL.Query().Get("redirect_to")
	if returnURL == "" {
		returnURL = withLang("/profiles/"+profileID, lang)
	}
	formData.CancelURL = returnURL
	formData.ReturnURL = returnURL

	data, ok := a.deps.DetailData(profileID, lang)
	if !ok {
		http.NotFound(w, r)
		return
	}
	data.FormModal = &formData
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if isHomeRedirect(returnURL) {
		home := a.deps.HomeData(lang)
		home.CurrentURL = currentRequestURL(r, lang)
		selectedID := r.URL.Query().Get("selected")
		if selectedID == "" {
			selectedID = profileID
		}
		for i := range home.Cards {
			if profileIDFromDetailURL(home.Cards[i].DetailURL) == selectedID {
				home.Cards[i].Selected = true
			}
		}
		if detail, ok := a.deps.DetailData(selectedID, lang); ok {
			detail.FormModal = &formData
			detail.EditURL = appendRedirect(detail.EditURL, returnURL)
			home.SelectedDetail = &detail
		}
		home.FormModal = &formData
		if err := pages.Home(home).Render(r.Context(), w); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}
	data.CurrentURL = currentRequestURL(r, lang)
	data.EditURL = appendRedirect(data.EditURL, returnURL)
	if err := pages.Detail(data).Render(r.Context(), w); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (a *App) handleSaveProfile(w http.ResponseWriter, r *http.Request) {
	lang := requestLang(r)
	persistLanguagePreference(w, lang)
	id, err := a.deps.SaveProfile(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	http.Redirect(w, r, withLang(fallbackRedirect(r, "/profiles/"+id), lang), http.StatusSeeOther)
}

func (a *App) handleStartProfile(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	lang := requestLang(r)
	persistLanguagePreference(w, lang)
	if err := a.deps.StartProfile(id); err != nil {
		http.Redirect(w, r, withError(fallbackRedirect(r, "/profiles/"+id), lang, err), http.StatusSeeOther)
		return
	}
	http.Redirect(w, r, withLang(fallbackRedirect(r, "/profiles/"+id), lang), http.StatusSeeOther)
}

func (a *App) handleStopProfile(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	lang := requestLang(r)
	persistLanguagePreference(w, lang)
	if err := a.deps.StopProfile(id); err != nil {
		http.Redirect(w, r, withError(fallbackRedirect(r, "/profiles/"+id), lang, err), http.StatusSeeOther)
		return
	}
	http.Redirect(w, r, withLang(fallbackRedirect(r, "/profiles/"+id), lang), http.StatusSeeOther)
}

func (a *App) handleDeleteProfile(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	lang := requestLang(r)
	persistLanguagePreference(w, lang)
	if err := a.deps.DeleteProfile(id); err != nil {
		http.Redirect(w, r, withError("/profiles/"+id, lang, err), http.StatusSeeOther)
		return
	}
	http.Redirect(w, r, withLang("/", lang), http.StatusSeeOther)
}

func languageFromRequest(r *http.Request) string {
	return r.URL.Query().Get("lang")
}

func localeFromCookie(r *http.Request) string {
	cookie, err := r.Cookie(localeCookieName)
	if err != nil {
		return ""
	}
	return cookie.Value
}

func resolvedLanguage(r *http.Request) string {
	if lang := languageFromRequest(r); lang != "" {
		return i18n.Normalize(lang)
	}
	if lang := localeFromCookie(r); lang != "" {
		return i18n.Normalize(lang)
	}
	return i18n.LangZH
}

func requestLang(r *http.Request) string {
	if lang := languageFromRequest(r); lang != "" {
		return i18n.Normalize(lang)
	}
	if err := r.ParseForm(); err != nil {
		return resolvedLanguage(r)
	}
	if lang := r.FormValue("lang"); lang != "" {
		return i18n.Normalize(lang)
	}
	return resolvedLanguage(r)
}

func withLang(path, lang string) string {
	if lang == "" {
		return path
	}
	return withQueryValue(path, "lang", i18n.Normalize(lang))
}

func withError(path, lang string, err error) string {
	values := map[string]string{}
	if lang != "" {
		values["lang"] = i18n.Normalize(lang)
	}
	if err != nil {
		values["error"] = err.Error()
	}
	return withQueryValues(path, values)
}

func fallbackRedirect(r *http.Request, fallback string) string {
	if err := r.ParseForm(); err == nil {
		if redirectTo := r.FormValue("redirect_to"); redirectTo != "" {
			return redirectTo
		}
	}
	return fallback
}

func profileIDFromDetailURL(detailURL string) string {
	const prefix = "/profiles/"
	if len(detailURL) <= len(prefix) || detailURL[:len(prefix)] != prefix {
		return ""
	}
	id := detailURL[len(prefix):]
	for i := 0; i < len(id); i++ {
		if id[i] == '?' || id[i] == '/' {
			return id[:i]
		}
	}
	return id
}

func appendRedirect(path, redirectTo string) string {
	if redirectTo == "" {
		return path
	}
	return withQueryValue(path, "redirect_to", redirectTo)
}

func isHomeRedirect(path string) bool {
	return path == "/" || strings.HasPrefix(path, "/?")
}

func selectedHomeURL(lang, profileID string) string {
	return withQueryValue(withLang("/", lang), "selected", profileID)
}

func currentRequestURL(r *http.Request, lang string) string {
	if r.URL == nil {
		return withLang("/", lang)
	}
	current := r.URL.RequestURI()
	if current == "" {
		current = "/"
	}
	return withLang(current, lang)
}

func persistLanguagePreference(w http.ResponseWriter, lang string) {
	http.SetCookie(w, &http.Cookie{
		Name:     localeCookieName,
		Value:    i18n.Normalize(lang),
		Path:     "/",
		MaxAge:   int((365 * 24 * time.Hour) / time.Second),
		HttpOnly: false,
		SameSite: http.SameSiteLaxMode,
	})
}

func withQueryValue(rawPath, key, value string) string {
	return withQueryValues(rawPath, map[string]string{key: value})
}

func withQueryValues(rawPath string, updates map[string]string) string {
	if rawPath == "" {
		rawPath = "/"
	}
	parsed, err := url.Parse(rawPath)
	if err != nil {
		return rawPath
	}
	values := parsed.Query()
	for key, value := range updates {
		if strings.TrimSpace(value) == "" {
			values.Del(key)
			continue
		}
		values.Set(key, value)
	}
	parsed.RawQuery = values.Encode()
	pathValue := parsed.Path
	if pathValue == "" {
		pathValue = "/"
	}
	if parsed.RawQuery == "" {
		return pathValue
	}
	return pathValue + "?" + parsed.RawQuery
}

func readUploadedImportFile(r *http.Request) ([]byte, error) {
	if err := r.ParseMultipartForm(8 << 20); err != nil {
		return nil, fmt.Errorf("parse upload: %w", err)
	}
	file, _, err := r.FormFile("file")
	if err != nil {
		return nil, fmt.Errorf("read uploaded file: %w", err)
	}
	defer file.Close()
	data, err := io.ReadAll(io.LimitReader(file, 8<<20))
	if err != nil {
		return nil, fmt.Errorf("read uploaded file data: %w", err)
	}
	return data, nil
}

func decodeImportPayload(encoded string) ([]byte, error) {
	data, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		return nil, fmt.Errorf("decode import payload: %w", err)
	}
	return data, nil
}
