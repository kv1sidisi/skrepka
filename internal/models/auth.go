package models

// OIDCRequest defines the structure for an OIDC-based authentication request.
type OIDCRequest struct {
	Provider Provider `json:"provider"`
	IDToken  string   `json:"id_token"`
}

// AuthResponse defines the structure for a successful authentication response.
// It contains the internal JWT issued by the service.
type AuthResponse struct {
	Token string `json:"token"`
}
