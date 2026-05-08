package ui

import (
	"net/http"
	"net/url"
	"strings"

	"github.com/go-chi/chi/v5"

	"timber-ssh-go/ui/pages"
)

func (a *App) handleHome(w http.ResponseWriter, r *http.Request) {
	data := a.deps.HomeData(languageFromRequest(r))
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
		if detail, ok := a.deps.DetailData(selectedID, languageFromRequest(r)); ok {
			returnURL := selectedHomeURL(languageFromRequest(r), selectedID)
			detail.EditURL = appendRedirect(detail.EditURL, returnURL)
			data.SelectedDetail = &detail
		}
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := pages.Home(data).Render(r.Context(), w); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (a *App) handleDetail(w http.ResponseWriter, r *http.Request) {
	data, ok := a.deps.DetailData(chi.URLParam(r, "id"), languageFromRequest(r))
	if !ok {
		http.NotFound(w, r)
		return
	}
	data.ErrorMessage = r.URL.Query().Get("error")
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := pages.Detail(data).Render(r.Context(), w); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (a *App) handleDiagnostics(w http.ResponseWriter, r *http.Request) {
	data, ok := a.deps.DetailData(chi.URLParam(r, "id"), languageFromRequest(r))
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
	data := a.deps.HomeData(languageFromRequest(r))
	form := a.deps.NewFormData(languageFromRequest(r))
	form.CancelURL = withLang("/", languageFromRequest(r))
	form.ReturnURL = withLang("/", languageFromRequest(r))
	data.FormModal = &form
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := pages.Home(data).Render(r.Context(), w); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (a *App) handleEditProfile(w http.ResponseWriter, r *http.Request) {
	lang := languageFromRequest(r)
	profileID := chi.URLParam(r, "id")
	formData, ok := a.deps.EditFormData(chi.URLParam(r, "id"), languageFromRequest(r))
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
	data.EditURL = appendRedirect(data.EditURL, returnURL)
	if err := pages.Detail(data).Render(r.Context(), w); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (a *App) handleSaveProfile(w http.ResponseWriter, r *http.Request) {
	id, err := a.deps.SaveProfile(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	http.Redirect(w, r, withLang(fallbackRedirect(r, "/profiles/"+id), requestLang(r)), http.StatusSeeOther)
}

func (a *App) handleStartProfile(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	lang := requestLang(r)
	if err := a.deps.StartProfile(id); err != nil {
		http.Redirect(w, r, withError(fallbackRedirect(r, "/profiles/"+id), lang, err), http.StatusSeeOther)
		return
	}
	http.Redirect(w, r, withLang(fallbackRedirect(r, "/profiles/"+id), lang), http.StatusSeeOther)
}

func (a *App) handleStopProfile(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	lang := requestLang(r)
	if err := a.deps.StopProfile(id); err != nil {
		http.Redirect(w, r, withError(fallbackRedirect(r, "/profiles/"+id), lang, err), http.StatusSeeOther)
		return
	}
	http.Redirect(w, r, withLang(fallbackRedirect(r, "/profiles/"+id), lang), http.StatusSeeOther)
}

func (a *App) handleDeleteProfile(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	lang := requestLang(r)
	if err := a.deps.DeleteProfile(id); err != nil {
		http.Redirect(w, r, withError("/profiles/"+id, lang, err), http.StatusSeeOther)
		return
	}
	http.Redirect(w, r, withLang("/", lang), http.StatusSeeOther)
}

func languageFromRequest(r *http.Request) string {
	return r.URL.Query().Get("lang")
}

func requestLang(r *http.Request) string {
	if lang := languageFromRequest(r); lang != "" {
		return lang
	}
	if err := r.ParseForm(); err != nil {
		return ""
	}
	return r.FormValue("lang")
}

func withLang(path, lang string) string {
	if lang == "" {
		return path
	}
	if strings.Contains(path, "lang=") {
		return path
	}
	return path + "?lang=" + lang
}

func withError(path, lang string, err error) string {
	values := url.Values{}
	if lang != "" {
		values.Set("lang", lang)
	}
	if err != nil {
		values.Set("error", err.Error())
	}
	encoded := values.Encode()
	if encoded == "" {
		return path
	}
	return path + "?" + encoded
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
	sep := "?"
	if strings.Contains(path, "?") {
		sep = "&"
	}
	return path + sep + "redirect_to=" + url.QueryEscape(redirectTo)
}

func isHomeRedirect(path string) bool {
	return path == "/" || strings.HasPrefix(path, "/?")
}

func selectedHomeURL(lang, profileID string) string {
	path := "/?selected=" + profileID
	if lang == "" {
		return path
	}
	return "/?lang=" + lang + "&selected=" + profileID
}
