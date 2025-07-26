package models

// Provider is special type for auth provider names.
// Using this type helps avoid mistakes with strings.
type Provider string

const (
	ProviderGoogle Provider = "google"
	ProviderApple  Provider = "apple"
)

func (p Provider) String() string {
	return string(p)
}
