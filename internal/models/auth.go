package models

// OIDCRequest is structure for OIDC authentication request.
// It contains provider name and ID token from that provider.
type OIDCRequest struct {
	Provider Provider `json:"provider"`
	IDToken  string   `json:"id_token"`
}

// AuthResponse is structure for successful authentication response.
// It contains internal JWT issued by our service.
type AuthResponse struct {
	Token string `json:"token"`
}
