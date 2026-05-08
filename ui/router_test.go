package ui

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"timber-ssh-go/ui/viewmodel"
)

func TestRouterRendersHomePage(t *testing.T) {
	app := NewApp(RouterDeps{
		HomeData: func(lang string) viewmodel.HomePageData {
			return viewmodel.HomePageData{
				Lang:      lang,
				Title:     "Timber SSH Go",
				Version:   "0.0.2",
				HomeURL:   "/",
				NewURL:    "/profiles/new",
				LegacyURL: "/legacy",
				Cards: []viewmodel.ProfileCard{
					{
						Lang:            lang,
						Title:           "数据库隧道",
						SSHAddress:      "ssh.example.com:22",
						LocalEndpoint:   "127.0.0.1:15432",
						RemoteTarget:    "127.0.0.1:5432",
						LastEnabledText: "未启用",
						StateLabel:      "未连接",
						DetailURL:       "/profiles/profile-1",
						StartURL:        "/profiles/profile-1/start",
						StopURL:         "/profiles/profile-1/stop",
						CanStart:        true,
						RedirectURL:     "/",
					},
				},
			}
		},
		DetailData: func(id string, lang string) (viewmodel.DetailPageData, bool) {
			return viewmodel.DetailPageData{
				Lang:      lang,
				Title:     "Timber SSH Go",
				Version:   "0.0.2",
				HomeURL:   "/",
				EditURL:   "/profiles/" + id + "/edit",
				ProfileID: id,
				StartURL:  "/profiles/" + id + "/start",
				StopURL:   "/profiles/" + id + "/stop",
				DeleteURL: "/profiles/" + id + "/delete",
				CanStart:  true,
				Card: viewmodel.ProfileCard{
					Lang:          lang,
					Title:         "数据库隧道",
					SSHAddress:    "ssh.example.com:22",
					LocalEndpoint: "127.0.0.1:15432",
					RemoteTarget:  "127.0.0.1:5432",
					StateLabel:    "未连接",
				},
				Diagnostics: viewmodel.DiagnosticsPanel{
					ListenerReady:     true,
					SSHConnected:      false,
					ActiveConnections: "0",
					UploadRateText:    "0 B/s",
					DownloadRateText:  "0 B/s",
					UploadTotalText:   "0 B",
					DownloadTotalText: "0 B",
				},
			}, true
		},
		NewFormData: func(lang string) viewmodel.FormPageData {
			return viewmodel.FormPageData{Lang: lang, Title: "Timber SSH Go", Version: "0.0.2", HomeURL: "/", CancelURL: "/", PostURL: "/profiles/save", SSHPort: 22, LocalAddr: "127.0.0.1", LocalPort: 15432, RemoteHost: "127.0.0.1", RemotePort: 5432, ConnectTimeoutSec: 10}
		},
		EditFormData: func(id string, lang string) (viewmodel.FormPageData, bool) {
			return viewmodel.FormPageData{Lang: lang, Title: "Timber SSH Go", Version: "0.0.2", HomeURL: "/", CancelURL: "/profiles/" + id, PostURL: "/profiles/save", ProfileID: id, IsEdit: true, SSHPort: 22, LocalAddr: "127.0.0.1", LocalPort: 15432, RemoteHost: "127.0.0.1", RemotePort: 5432, ConnectTimeoutSec: 10}, true
		},
		SaveProfile:   func(r *http.Request) (string, error) { return "profile-1", nil },
		StartProfile:  func(id string) error { return nil },
		StopProfile:   func(id string) error { return nil },
		DeleteProfile: func(id string) error { return nil },
	})

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	app.Router().ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}
	body := rec.Body.String()
	for _, want := range []string{"Timber SSH Go", "数据库隧道", "ssh.example.com:22", "新增配置", "刷新"} {
		if !strings.Contains(body, want) {
			t.Fatalf("response missing %q:\n%s", want, body)
		}
	}
}

func TestRouterSwitchesLanguage(t *testing.T) {
	app := NewApp(RouterDeps{
		HomeData: func(lang string) viewmodel.HomePageData {
			return viewmodel.HomePageData{Lang: lang, Title: "Timber SSH Go", Version: "0.0.2", HomeURL: "/", NewURL: "/profiles/new"}
		},
		DetailData: func(id string, lang string) (viewmodel.DetailPageData, bool) {
			return viewmodel.DetailPageData{}, false
		},
		NewFormData:   func(lang string) viewmodel.FormPageData { return viewmodel.FormPageData{} },
		EditFormData:  func(id string, lang string) (viewmodel.FormPageData, bool) { return viewmodel.FormPageData{}, false },
		SaveProfile:   func(r *http.Request) (string, error) { return "", nil },
		StartProfile:  func(id string) error { return nil },
		StopProfile:   func(id string) error { return nil },
		DeleteProfile: func(id string) error { return nil },
	})

	req := httptest.NewRequest(http.MethodGet, "/?lang=en", nil)
	rec := httptest.NewRecorder()
	app.Router().ServeHTTP(rec, req)
	body := rec.Body.String()
	if !strings.Contains(body, "New Profile") || !strings.Contains(body, "Refresh") {
		t.Fatalf("english response missing translated text:\n%s", body)
	}
}

func TestRouterRendersNewProfileAsModal(t *testing.T) {
	app := NewApp(RouterDeps{
		HomeData: func(lang string) viewmodel.HomePageData {
			return viewmodel.HomePageData{
				Lang:    lang,
				Title:   "Timber SSH Go",
				Version: "0.0.2",
				HomeURL: "/",
				NewURL:  "/profiles/new",
				Cards: []viewmodel.ProfileCard{
					{Lang: lang, Title: "示例", StateLabel: "未连接"},
				},
			}
		},
		DetailData: func(id string, lang string) (viewmodel.DetailPageData, bool) {
			return viewmodel.DetailPageData{}, false
		},
		NewFormData: func(lang string) viewmodel.FormPageData {
			return viewmodel.FormPageData{Lang: lang, Version: "0.0.2", HomeURL: "/", CancelURL: "/", PostURL: "/profiles/save", SSHPort: 22, LocalAddr: "127.0.0.1", LocalPort: 15432, RemoteHost: "127.0.0.1", RemotePort: 5432, ConnectTimeoutSec: 10}
		},
		EditFormData:  func(id string, lang string) (viewmodel.FormPageData, bool) { return viewmodel.FormPageData{}, false },
		SaveProfile:   func(r *http.Request) (string, error) { return "", nil },
		StartProfile:  func(id string) error { return nil },
		StopProfile:   func(id string) error { return nil },
		DeleteProfile: func(id string) error { return nil },
	})

	req := httptest.NewRequest(http.MethodGet, "/profiles/new", nil)
	rec := httptest.NewRecorder()
	app.Router().ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}
	body := rec.Body.String()
	for _, want := range []string{"modal-backdrop", "新增配置", "关闭", "Timber SSH Go"} {
		if !strings.Contains(body, want) {
			t.Fatalf("response missing %q:\n%s", want, body)
		}
	}
}

func TestRouterRendersEditProfileAsModal(t *testing.T) {
	app := NewApp(RouterDeps{
		HomeData:      func(lang string) viewmodel.HomePageData { return viewmodel.HomePageData{} },
		NewFormData:   func(lang string) viewmodel.FormPageData { return viewmodel.FormPageData{} },
		SaveProfile:   func(r *http.Request) (string, error) { return "", nil },
		StartProfile:  func(id string) error { return nil },
		StopProfile:   func(id string) error { return nil },
		DeleteProfile: func(id string) error { return nil },
		DetailData: func(id string, lang string) (viewmodel.DetailPageData, bool) {
			return viewmodel.DetailPageData{
				Lang:    lang,
				Title:   "Timber SSH Go",
				Version: "0.0.2",
				HomeURL: "/",
				Card: viewmodel.ProfileCard{
					Lang:          lang,
					Title:         "数据库隧道",
					SSHAddress:    "ssh.example.com:22",
					LocalEndpoint: "127.0.0.1:15432",
					RemoteTarget:  "127.0.0.1:5432",
					StateLabel:    "未连接",
				},
				Diagnostics: viewmodel.DiagnosticsPanel{
					ActiveConnections: "0",
				},
			}, true
		},
		EditFormData: func(id string, lang string) (viewmodel.FormPageData, bool) {
			return viewmodel.FormPageData{
				Lang:              lang,
				Version:           "0.0.2",
				HomeURL:           "/",
				CancelURL:         "/profiles/" + id,
				PostURL:           "/profiles/save",
				ProfileID:         id,
				IsEdit:            true,
				ProfileTitle:      "数据库隧道",
				SSHPort:           22,
				LocalAddr:         "127.0.0.1",
				LocalPort:         15432,
				RemoteHost:        "127.0.0.1",
				RemotePort:        5432,
				ConnectTimeoutSec: 10,
			}, true
		},
	})

	req := httptest.NewRequest(http.MethodGet, "/profiles/profile-1/edit", nil)
	rec := httptest.NewRecorder()
	app.Router().ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}
	body := rec.Body.String()
	for _, want := range []string{"modal-backdrop", "编辑配置", "数据库隧道"} {
		if !strings.Contains(body, want) {
			t.Fatalf("response missing %q:\n%s", want, body)
		}
	}
}

func TestRouterRendersEmbeddedEditProfileModalWithHomeCancel(t *testing.T) {
	app := NewApp(RouterDeps{
		HomeData: func(lang string) viewmodel.HomePageData {
			return viewmodel.HomePageData{
				Lang:    lang,
				Title:   "Timber SSH Go",
				Version: "0.0.2",
				HomeURL: "/",
				NewURL:  "/profiles/new",
				Cards: []viewmodel.ProfileCard{
					{Lang: lang, Title: "数据库隧道", DetailURL: "/?selected=profile-1"},
				},
			}
		},
		NewFormData:   func(lang string) viewmodel.FormPageData { return viewmodel.FormPageData{} },
		SaveProfile:   func(r *http.Request) (string, error) { return "", nil },
		StartProfile:  func(id string) error { return nil },
		StopProfile:   func(id string) error { return nil },
		DeleteProfile: func(id string) error { return nil },
		DetailData: func(id string, lang string) (viewmodel.DetailPageData, bool) {
			return viewmodel.DetailPageData{
				Lang:    lang,
				Title:   "Timber SSH Go",
				Version: "0.0.2",
				HomeURL: "/",
				EditURL: "/profiles/" + id + "/edit",
				Card: viewmodel.ProfileCard{
					Lang:  lang,
					Title: "数据库隧道",
				},
			}, true
		},
		EditFormData: func(id string, lang string) (viewmodel.FormPageData, bool) {
			return viewmodel.FormPageData{
				Lang:         lang,
				Version:      "0.0.2",
				HomeURL:      "/",
				CancelURL:    "/profiles/" + id,
				PostURL:      "/profiles/save",
				ProfileID:    id,
				IsEdit:       true,
				ProfileTitle: "数据库隧道",
			}, true
		},
	})

	req := httptest.NewRequest(http.MethodGet, "/profiles/profile-1/edit?lang=zh-CN&redirect_to=%2F%3Flang%3Dzh-CN%26selected%3Dprofile-1&selected=profile-1", nil)
	rec := httptest.NewRecorder()
	app.Router().ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}
	body := rec.Body.String()
	for _, want := range []string{`href="/?lang=zh-CN&amp;selected=profile-1"`, `name="redirect_to" value="/?lang=zh-CN&amp;selected=profile-1"`} {
		if !strings.Contains(body, want) {
			t.Fatalf("response missing %q:\n%s", want, body)
		}
	}
}

func TestRouterPreservesLanguageWhenSavingProfile(t *testing.T) {
	app := NewApp(RouterDeps{
		HomeData: func(lang string) viewmodel.HomePageData { return viewmodel.HomePageData{} },
		DetailData: func(id string, lang string) (viewmodel.DetailPageData, bool) {
			return viewmodel.DetailPageData{}, false
		},
		NewFormData:   func(lang string) viewmodel.FormPageData { return viewmodel.FormPageData{} },
		EditFormData:  func(id string, lang string) (viewmodel.FormPageData, bool) { return viewmodel.FormPageData{}, false },
		SaveProfile:   func(r *http.Request) (string, error) { return "profile-1", nil },
		StartProfile:  func(id string) error { return nil },
		StopProfile:   func(id string) error { return nil },
		DeleteProfile: func(id string) error { return nil },
	})

	req := httptest.NewRequest(http.MethodPost, "/profiles/save", strings.NewReader("lang=en"))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rec := httptest.NewRecorder()
	app.Router().ServeHTTP(rec, req)

	if rec.Code != http.StatusSeeOther {
		t.Fatalf("status = %d, want 303", rec.Code)
	}
	if location := rec.Header().Get("Location"); location != "/profiles/profile-1?lang=en" {
		t.Fatalf("Location = %q, want %q", location, "/profiles/profile-1?lang=en")
	}
}

func TestRouterSaveProfileSupportsCustomRedirect(t *testing.T) {
	app := NewApp(RouterDeps{
		HomeData: func(lang string) viewmodel.HomePageData { return viewmodel.HomePageData{} },
		DetailData: func(id string, lang string) (viewmodel.DetailPageData, bool) {
			return viewmodel.DetailPageData{}, false
		},
		NewFormData:   func(lang string) viewmodel.FormPageData { return viewmodel.FormPageData{} },
		EditFormData:  func(id string, lang string) (viewmodel.FormPageData, bool) { return viewmodel.FormPageData{}, false },
		SaveProfile:   func(r *http.Request) (string, error) { return "profile-1", nil },
		StartProfile:  func(id string) error { return nil },
		StopProfile:   func(id string) error { return nil },
		DeleteProfile: func(id string) error { return nil },
	})

	req := httptest.NewRequest(http.MethodPost, "/profiles/save", strings.NewReader("lang=en&redirect_to=%2F%3Flang%3Den%26selected%3Dprofile-1"))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rec := httptest.NewRecorder()
	app.Router().ServeHTTP(rec, req)

	if rec.Code != http.StatusSeeOther {
		t.Fatalf("status = %d, want 303", rec.Code)
	}
	if location := rec.Header().Get("Location"); location != "/?lang=en&selected=profile-1" {
		t.Fatalf("Location = %q, want %q", location, "/?lang=en&selected=profile-1")
	}
}

func TestRouterStartProfileRedirectsToDetail(t *testing.T) {
	started := ""
	app := NewApp(RouterDeps{
		HomeData: func(lang string) viewmodel.HomePageData { return viewmodel.HomePageData{} },
		DetailData: func(id string, lang string) (viewmodel.DetailPageData, bool) {
			return viewmodel.DetailPageData{}, false
		},
		NewFormData:   func(lang string) viewmodel.FormPageData { return viewmodel.FormPageData{} },
		EditFormData:  func(id string, lang string) (viewmodel.FormPageData, bool) { return viewmodel.FormPageData{}, false },
		SaveProfile:   func(r *http.Request) (string, error) { return "", nil },
		StartProfile:  func(id string) error { started = id; return nil },
		StopProfile:   func(id string) error { return nil },
		DeleteProfile: func(id string) error { return nil },
	})

	req := httptest.NewRequest(http.MethodPost, "/profiles/profile-1/start", strings.NewReader("lang=en"))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rec := httptest.NewRecorder()
	app.Router().ServeHTTP(rec, req)

	if started != "profile-1" {
		t.Fatalf("started = %q, want %q", started, "profile-1")
	}
	if rec.Code != http.StatusSeeOther {
		t.Fatalf("status = %d, want 303", rec.Code)
	}
	if location := rec.Header().Get("Location"); location != "/profiles/profile-1?lang=en" {
		t.Fatalf("Location = %q, want %q", location, "/profiles/profile-1?lang=en")
	}
}

func TestRouterStartProfileSupportsCustomRedirect(t *testing.T) {
	app := NewApp(RouterDeps{
		HomeData: func(lang string) viewmodel.HomePageData { return viewmodel.HomePageData{} },
		DetailData: func(id string, lang string) (viewmodel.DetailPageData, bool) {
			return viewmodel.DetailPageData{}, false
		},
		NewFormData:   func(lang string) viewmodel.FormPageData { return viewmodel.FormPageData{} },
		EditFormData:  func(id string, lang string) (viewmodel.FormPageData, bool) { return viewmodel.FormPageData{}, false },
		SaveProfile:   func(r *http.Request) (string, error) { return "", nil },
		StartProfile:  func(id string) error { return nil },
		StopProfile:   func(id string) error { return nil },
		DeleteProfile: func(id string) error { return nil },
	})

	req := httptest.NewRequest(http.MethodPost, "/profiles/profile-1/start", strings.NewReader("lang=en&redirect_to=%2F"))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rec := httptest.NewRecorder()
	app.Router().ServeHTTP(rec, req)

	if location := rec.Header().Get("Location"); location != "/?lang=en" {
		t.Fatalf("Location = %q, want %q", location, "/?lang=en")
	}
}

func TestRouterRendersDiagnosticsFragment(t *testing.T) {
	app := NewApp(RouterDeps{
		HomeData:      func(lang string) viewmodel.HomePageData { return viewmodel.HomePageData{} },
		NewFormData:   func(lang string) viewmodel.FormPageData { return viewmodel.FormPageData{} },
		EditFormData:  func(id string, lang string) (viewmodel.FormPageData, bool) { return viewmodel.FormPageData{}, false },
		SaveProfile:   func(r *http.Request) (string, error) { return "", nil },
		StartProfile:  func(id string) error { return nil },
		StopProfile:   func(id string) error { return nil },
		DeleteProfile: func(id string) error { return nil },
		DetailData: func(id string, lang string) (viewmodel.DetailPageData, bool) {
			return viewmodel.DetailPageData{
				Lang: lang,
				Diagnostics: viewmodel.DiagnosticsPanel{
					ListenerReady:     true,
					SSHConnected:      true,
					ActiveConnections: "2",
					UploadRateText:    "12 KB/s",
					DownloadRateText:  "8 KB/s",
					UploadTotalText:   "100 KB",
					DownloadTotalText: "220 KB",
					RefreshURL:        "/profiles/" + id + "/diagnostics?lang=" + lang,
					Logs: []viewmodel.DiagnosticsLog{
						{TimeText: "2026-05-08 12:00:00", Message: "connected"},
					},
				},
			}, true
		},
	})

	req := httptest.NewRequest(http.MethodGet, "/profiles/profile-1/diagnostics?lang=en", nil)
	rec := httptest.NewRecorder()
	app.Router().ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}
	body := rec.Body.String()
	for _, want := range []string{"Diagnostics", "Auto Refresh", "12 KB/s", "connected"} {
		if !strings.Contains(body, want) {
			t.Fatalf("response missing %q:\n%s", want, body)
		}
	}
}
