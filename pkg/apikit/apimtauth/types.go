package apimtauth

type LoginRequest struct {
	Username   string
	Password   string
	TotpSecret string
	VisitorID  string
}

type AuthState struct {
	Token     string
	DID       string
	VisitorID string
}
