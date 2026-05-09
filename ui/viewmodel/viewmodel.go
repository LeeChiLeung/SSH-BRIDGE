package viewmodel

type HomePageData struct {
	Title          string
	Version        string
	Lang           string
	CurrentURL     string
	HomeURL        string
	NewURL         string
	ExportURL      string
	ImportURL      string
	LegacyURL      string
	Cards          []ProfileCard
	FormModal      *FormPageData
	ImportModal    *ImportModalData
	ImportResult   *ImportResultData
	SelectedDetail *DetailPageData
}

type ProfileCard struct {
	Lang            string
	Title           string
	SSHAddress      string
	LocalEndpoint   string
	RemoteTarget    string
	LastEnabledText string
	StateLabel      string
	StateTone       string
	DetailURL       string
	StartURL        string
	StopURL         string
	CanStart        bool
	CanStop         bool
	RedirectURL     string
	Selected        bool
}
