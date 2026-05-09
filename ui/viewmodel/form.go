package viewmodel

type FormPageData struct {
	Lang      string
	Version   string
	Title     string
	HomeURL   string
	CancelURL string
	ReturnURL string
	PostURL   string
	IsEdit    bool
	ProfileID string

	ProfileTitle       string
	SSHHost            string
	SSHPort            int
	SSHUser            string
	SSHPassword        string
	SSHKeyPath         string
	LocalAddr          string
	LocalPort          int
	RemoteHost         string
	RemotePort         int
	ConnectTimeoutSec  int
	KnownHostsPath     string
	InsecureIgnoreHost bool
	AutoStart          bool
}
