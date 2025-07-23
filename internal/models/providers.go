package models

// Provider - a typed constant for auth provider names.
type Provider string

const (
	ProviderGoogle Provider = "google"
	ProviderApple  Provider = "apple"
)
