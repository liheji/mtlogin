package apimtbrowse

type AuthState struct {
	Token     string
	DID       string
	VisitorID string
}

type Profile struct {
	Username        string
	UploadedBytes   int64
	DownloadedBytes int64
	Bonus           float64
	LastLogin       string
	LastBrowse      string
}
