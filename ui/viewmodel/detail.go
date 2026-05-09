package viewmodel

type DetailPageData struct {
	Lang      string
	Title     string
	Version   string
	CurrentURL string
	HomeURL   string
	EditURL   string
	LegacyURL string
	Card      ProfileCard

	ProfileID    string
	StartURL     string
	StopURL      string
	DeleteURL    string
	ErrorMessage string
	CanStart     bool
	CanStop      bool
	Diagnostics  DiagnosticsPanel
	FormModal    *FormPageData
}

type DiagnosticsPanel struct {
	ListenerReady     bool
	SSHConnected      bool
	LastError         string
	UploadRateText    string
	DownloadRateText  string
	UploadTotalText   string
	DownloadTotalText string
	ActiveConnections string
	RefreshURL        string
	Logs              []DiagnosticsLog
}

type DiagnosticsLog struct {
	TimeText string
	Message  string
}
