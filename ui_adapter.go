package main

import (
	"fmt"
	"net/http"
	"time"

	"timber-ssh-go/ui"
	"timber-ssh-go/ui/i18n"
	"timber-ssh-go/ui/viewmodel"
)

func newUIRouter(controller *TunnelController, store *ProfileStore, legacyURL string) *ui.App {
	return ui.NewApp(ui.RouterDeps{
		HomeData: func(lang string) viewmodel.HomePageData {
			return viewmodel.HomePageData{
				Lang:      lang,
				Title:     "Timber SSH Go",
				Version:   currentBuildInfo().Version,
				HomeURL:   "/",
				NewURL:    withLangUI("/profiles/new", lang),
				ExportURL: withLangUI("/profiles/export", lang),
				ImportURL: withLangUI("/profiles/import", lang),
				LegacyURL: legacyURL,
				Cards:     buildUIProfileCards(controller, store, lang),
			}
		},
		DetailData: func(id string, lang string) (viewmodel.DetailPageData, bool) {
			profile, ok := store.Get(id)
			if !ok {
				return viewmodel.DetailPageData{}, false
			}
			snapshot := controller.Snapshot(profile.ID)
			card := buildProfileCard(profile, snapshot, lang)
			return viewmodel.DetailPageData{
				Lang:        lang,
				Title:       "Timber SSH Go",
				Version:     currentBuildInfo().Version,
				HomeURL:     homeSelectedURL(lang, profile.ID),
				EditURL:     withLangUI("/profiles/"+profile.ID+"/edit", lang),
				LegacyURL:   legacyURL,
				Card:        card,
				ProfileID:   profile.ID,
				StartURL:    withLangUI("/profiles/"+profile.ID+"/start", lang),
				StopURL:     withLangUI("/profiles/"+profile.ID+"/stop", lang),
				DeleteURL:   withLangUI("/profiles/"+profile.ID+"/delete", lang),
				CanStart:    snapshot.State == TunnelStateDisconnected,
				CanStop:     snapshot.State != TunnelStateDisconnected,
				Diagnostics: buildDiagnosticsPanel(profile.ID, lang, snapshot),
			}, true
		},
		NewFormData: func(lang string) viewmodel.FormPageData {
			return buildFormData(profileFromConfig("", defaultTunnelConfig()), false, lang)
		},
		NewImportData: func(lang string) viewmodel.ImportModalData {
			return buildImportModalData(lang)
		},
		EditFormData: func(id string, lang string) (viewmodel.FormPageData, bool) {
			profile, ok := store.Get(id)
			if !ok {
				return viewmodel.FormPageData{}, false
			}
			return buildFormData(profile, true, lang), true
		},
		SaveProfile: func(r *http.Request) (string, error) {
			profile, err := profileFromRequest(r)
			if err != nil {
				return "", err
			}
			saved, err := store.Save(profile)
			if err != nil {
				return "", err
			}
			return saved.ID, nil
		},
		StartProfile: func(id string) error {
			profile, ok := store.Get(id)
			if !ok {
				return fmt.Errorf("profile not found")
			}
			return controller.Start(id, profile.TunnelConfig())
		},
		StopProfile: func(id string) error {
			controller.Stop(id)
			return nil
		},
		DeleteProfile: func(id string) error {
			snapshot := controller.Snapshot(id)
			if snapshot.State != TunnelStateDisconnected {
				return fmt.Errorf("cannot delete a running profile")
			}
			return store.Delete(id)
		},
		ExportProfiles: func() (string, []byte, error) {
			data, err := store.ExportTransferJSON(time.Now())
			if err != nil {
				return "", nil, err
			}
			return "timber-ssh-go-profiles.json", data, nil
		},
		PreviewImport: func(data []byte) (viewmodel.ImportModalData, error) {
			preview, err := store.PreviewImport(data)
			if err != nil {
				return viewmodel.ImportModalData{}, err
			}
			return buildImportPreviewModalData(preview, data, ""), nil
		},
		ApplyImport: func(data []byte, policy string) (viewmodel.ImportResultData, error) {
			result, err := store.ApplyImport(data, policy)
			if err != nil {
				return viewmodel.ImportResultData{}, err
			}
			return viewmodel.ImportResultData(result), nil
		},
	})
}

func buildUIProfileCards(controller *TunnelController, store *ProfileStore, lang string) []viewmodel.ProfileCard {
	profiles := store.List()
	snapshots := controller.Snapshots()
	cards := make([]viewmodel.ProfileCard, 0, len(profiles))
	for _, profile := range profiles {
		cards = append(cards, buildProfileCard(profile, snapshots[profile.ID], lang))
	}
	return cards
}

func buildProfileCard(profile Profile, snapshot TunnelSnapshot, lang string) viewmodel.ProfileCard {
	cfg := profile.TunnelConfig()
	card := viewmodel.ProfileCard{
		Lang:            lang,
		Title:           profile.Title,
		SSHAddress:      fmt.Sprintf("%s:%d", profile.SSHHost, profile.SSHPort),
		LocalEndpoint:   cfg.LocalEndpoint(),
		RemoteTarget:    cfg.RemoteAddr(),
		LastEnabledText: formatTime(profile.LastEnabledAt),
		StateLabel:      statusLabelText(lang, TunnelStateDisconnected),
		StateTone:       stateToneText(TunnelStateDisconnected, ""),
		DetailURL:       homeSelectedURL(lang, profile.ID),
		StartURL:        withLangUI("/profiles/"+profile.ID+"/start", lang),
		StopURL:         withLangUI("/profiles/"+profile.ID+"/stop", lang),
		CanStart:        true,
		CanStop:         false,
		RedirectURL:     homeSelectedURL(lang, profile.ID),
	}
	if snapshot.Endpoint != "" {
		card.LocalEndpoint = snapshot.Endpoint
	}
	if snapshot.State != "" {
		card.StateLabel = statusLabelText(lang, snapshot.State)
		card.StateTone = stateToneText(snapshot.State, snapshot.LastError)
		card.CanStart = snapshot.State == TunnelStateDisconnected
		card.CanStop = snapshot.State != TunnelStateDisconnected
	}
	return card
}

func buildDiagnosticsPanel(profileID, lang string, snapshot TunnelSnapshot) viewmodel.DiagnosticsPanel {
	logs := make([]viewmodel.DiagnosticsLog, 0, len(snapshot.Logs))
	for _, entry := range snapshot.Logs {
		logs = append(logs, viewmodel.DiagnosticsLog{
			TimeText: entry.At.Local().Format("2006-01-02 15:04:05"),
			Message:  entry.Message,
		})
	}
	return viewmodel.DiagnosticsPanel{
		ListenerReady:     snapshot.Health.ListenerReady,
		SSHConnected:      snapshot.Health.SSHConnected,
		LastError:         snapshot.LastError,
		UploadRateText:    formatRate(snapshot.Traffic.UploadBytesPerSec),
		DownloadRateText:  formatRate(snapshot.Traffic.DownloadBytesPerSec),
		UploadTotalText:   formatBytes(snapshot.Traffic.UploadBytesTotal),
		DownloadTotalText: formatBytes(snapshot.Traffic.DownloadBytesTotal),
		ActiveConnections: fmt.Sprintf("%d", snapshot.Traffic.ActiveConnections),
		RefreshURL:        withLangUI("/profiles/"+profileID+"/diagnostics", lang),
		Logs:              logs,
	}
}

func buildFormData(profile Profile, isEdit bool, lang string) viewmodel.FormPageData {
	cancelURL := withLangUI("/", lang)
	if isEdit && profile.ID != "" {
		cancelURL = withLangUI("/profiles/"+profile.ID, lang)
	}
	return viewmodel.FormPageData{
		Lang:               lang,
		Version:            currentBuildInfo().Version,
		Title:              "Timber SSH Go",
		HomeURL:            withLangUI("/", lang),
		CancelURL:          cancelURL,
		ReturnURL:          cancelURL,
		PostURL:            withLangUI("/profiles/save", lang),
		IsEdit:             isEdit,
		ProfileID:          profile.ID,
		ProfileTitle:       profile.Title,
		SSHHost:            profile.SSHHost,
		SSHPort:            profile.SSHPort,
		SSHUser:            profile.SSHUser,
		SSHPassword:        profile.SSHPassword,
		SSHKeyPath:         profile.SSHKeyPath,
		LocalAddr:          profile.LocalAddr,
		LocalPort:          profile.LocalPort,
		RemoteHost:         profile.RemoteHost,
		RemotePort:         profile.RemotePort,
		ConnectTimeoutSec:  profile.ConnectTimeoutSec,
		KnownHostsPath:     profile.KnownHostsPath,
		InsecureIgnoreHost: profile.InsecureIgnoreHost,
	}
}

func buildImportModalData(lang string) viewmodel.ImportModalData {
	return viewmodel.ImportModalData{
		Lang:      lang,
		Title:     "Timber SSH Go",
		Version:   currentBuildInfo().Version,
		CancelURL: withLangUI("/", lang),
		PreviewURL: withLangUI("/profiles/import/preview", lang),
		ConfirmURL: withLangUI("/profiles/import/confirm", lang),
	}
}

func buildImportPreviewModalData(preview ProfileImportPreview, raw []byte, lang string) viewmodel.ImportModalData {
	modal := buildImportModalData(lang)
	modal.HasPreview = true
	modal.HasConflicts = len(preview.Conflicts) > 0
	modal.PayloadBase64 = encodeImportPayload(raw)
	for _, item := range preview.Added {
		modal.AddedTitles = append(modal.AddedTitles, item.Title)
	}
	for _, item := range preview.Conflicts {
		modal.ConflictTitles = append(modal.ConflictTitles, item.Title)
	}
	for _, item := range preview.Invalid {
		modal.InvalidItems = append(modal.InvalidItems, viewmodel.ImportInvalidItem{
			Title:  item.Title,
			Reason: item.Reason,
		})
	}
	return modal
}

func withLangUI(path, lang string) string {
	if lang == "" {
		return path
	}
	return path + "?lang=" + lang
}

func homeSelectedURL(lang, profileID string) string {
	if lang == "" {
		return "/?selected=" + profileID
	}
	return "/?lang=" + lang + "&selected=" + profileID
}

func statusLabelText(lang string, state TunnelState) string {
	t := i18n.New(lang)
	switch state {
	case TunnelStateConnected:
		return t.T("state.connected")
	case TunnelStateConnecting:
		return t.T("state.connecting")
	case TunnelStateRetrying:
		return t.T("state.retrying")
	default:
		return t.T("state.disconnected")
	}
}

func stateToneText(state TunnelState, lastError string) string {
	switch state {
	case TunnelStateConnected:
		return "green"
	case TunnelStateConnecting, TunnelStateRetrying:
		return "blue"
	case TunnelStateDisconnected:
		if lastError != "" {
			return "red"
		}
		return "gray"
	default:
		if lastError != "" {
			return "red"
		}
		return "gray"
	}
}
