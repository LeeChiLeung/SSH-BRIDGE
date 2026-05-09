package viewmodel

type ImportModalData struct {
	Lang            string
	Title           string
	Version         string
	CancelURL       string
	PreviewURL      string
	ConfirmURL      string
	PayloadBase64   string
	HasPreview      bool
	HasConflicts    bool
	AddedTitles     []string
	ConflictTitles  []string
	InvalidItems    []ImportInvalidItem
	OverwriteNotice string
}

type ImportInvalidItem struct {
	Title  string
	Reason string
}

type ImportResultData struct {
	Added   int
	Merged  int
	Skipped int
	Invalid int
}
