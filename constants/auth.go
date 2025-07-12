package constants

type contextKey string

const UserIDKey contextKey = "userID"

type AuthProvider string

const (
	AuthProviderGoogle AuthProvider = "google"
	AuthProviderCreds  AuthProvider = "creds"
)
